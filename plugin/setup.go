package plugin

import (
	"context"
	"flag"
	"time"

	"github.com/dixudx/yacht"
	"github.com/lmxia/gaia/pkg/apis/apps/v1alpha1"
	"github.com/lmxia/gaia/pkg/common"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	gaiaclientset "github.com/lmxia/gaia/pkg/generated/clientset/versioned"
	gaiainformers "github.com/lmxia/gaia/pkg/generated/informers/externalversions"
	"github.com/pkg/errors"
)

var (
	masterURL  string
	kubeconfig string
)

// Hook for unit tests.
var buildKubeConfigFunc = clientcmd.BuildConfigFromFlags

// init registers this plugin within the Caddy plugin framework. It uses "example" as the
// name, and couples it to the Action "setup".
func init() {
	caddy.RegisterPlugin("crossdns", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	klog.Infof("In setup")

	cd, err := CrossDNSParse(c)
	if err != nil {
		return plugin.Error("crossdns", err) // nolint:wrapcheck // No need to wrap this.
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		cd.Next = next
		return cd
	})

	return nil
}

func CrossDNSParse(c *caddy.Controller) (*CrossDNS, error) {
	cfg, err := buildKubeConfigFunc(masterURL, kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "error building kubeconfig")
	}
	indexers := cache.Indexers{
		FQDNINDEX: func(obj interface{}) ([]string, error) {
			desc := obj.(*v1alpha1.Description)
			fqdnSlice := make([]string, 0)
			for _, item := range desc.Spec.WorkloadComponents {
				if len(item.FQDN) != 0 {
					fqdnSlice = append(fqdnSlice, item.FQDN)
				}
			}
			if len(fqdnSlice) == 0 {
				return []string{NONFQDN}, nil
			}
			return fqdnSlice, nil
		},
	}

	indices := cache.Indices{}
	// 反向索引的cache store
	indexedStore := cache.NewThreadSafeStore(indexers, indices)

	ctx := context.Background()
	initCtx, cancel := context.WithCancel(ctx)
	cd := &CrossDNS{}

	localGaiaClientSet := gaiaclientset.NewForConfigOrDie(cfg)

	localAllGaiaInformerFactory := gaiainformers.NewSharedInformerFactory(localGaiaClientSet, 2*time.Hour)
	rbInformer := localAllGaiaInformerFactory.Apps().V1alpha1().ResourceBindings()
	descInformer := localAllGaiaInformerFactory.Apps().V1alpha1().Descriptions()
	cd.rbLister = rbInformer.Lister()
	cd.descLister = descInformer.Lister()
	cd.rbSynced = rbInformer.Informer().HasSynced
	cd.indexedStore = indexedStore

	yachtController := yacht.NewController("desc").
		WithCacheSynced(descInformer.Informer().HasSynced).
		WithHandlerFunc(cd.Handle).
		WithEnqueueFilterFunc(func(oldObj, newObj interface{}) (bool, error) {
			var tempObj interface{}
			if newObj != nil {
				tempObj = newObj
			} else {
				tempObj = oldObj
			}
			if tempObj != nil {
				// we will make all make changes to cache.
				desc := tempObj.(*v1alpha1.Description)
				// not scheduled? surely we will ignore it.
				if desc.Status.Phase != v1alpha1.DescriptionPhaseScheduled || desc.Namespace != common.GaiaReservedNamespace {
					return false, nil
				}
				return true, nil
			}
			return false, nil
		})

	localAllGaiaInformerFactory.Start(ctx.Done())
	_, err = descInformer.Informer().AddEventHandler(yachtController.DefaultResourceEventHandlerFuncs())
	if err != nil {
		cancel()
		return nil, err
	}
	go wait.UntilWithContext(initCtx, func(ctx context.Context) {
		yachtController.Run(ctx)
	}, time.Duration(0))

	c.OnShutdown(func() error {
		cancel()
		return nil
	})

	if c.Next() {
		cd.Zones = c.RemainingArgs()
		if len(cd.Zones) == 0 {
			cd.Zones = make([]string, len(c.ServerBlockKeys))
			copy(cd.Zones, c.ServerBlockKeys)
		}

		for i, str := range cd.Zones {
			cd.Zones[i] = plugin.Host(str).Normalize()
		}

		for c.NextBlock() {
			switch c.Val() {
			case "fallthrough":
				cd.Fall.SetZonesFromArgs(c.RemainingArgs())
			default:
				if c.Val() != "}" {
					return nil, c.Errf("unknown property '%s'", c.Val()) // nolint:wrapcheck // No need to wrap this.
				}
			}
		}
	}
	return cd, nil
}

// Handle Actually don't really need this, make it happened in filter is also fine,
// I just don't want slow down enqueue proceed.
func (cd *CrossDNS) Handle(obj interface{}) (requeueAfter *time.Duration, err error) {
	failedPeriod := 2 * time.Second
	key := obj.(string)
	namespace, descName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("invalid description key: %s", key)
		return &failedPeriod, nil
	}
	cachedDesc, err := cd.descLister.Descriptions(namespace).Get(descName)
	if err != nil {
		klog.Errorf("can't get description from: %s/%s", namespace, descName)
		return &failedPeriod, nil
	}
	if cachedDesc.DeletionTimestamp != nil {
		cd.indexedStore.Delete(cachedDesc.Name)
		return nil, nil
	}
	// 添加索引
	cd.indexedStore.Add(cachedDesc.Name, cachedDesc)
	return nil, nil
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
