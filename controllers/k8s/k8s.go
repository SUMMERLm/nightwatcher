package k8s

import (
	"errors"
	"log"
	"sync"

	"github.com/lmxia/nightwatcher/utils"

	gaiaclientset "github.com/lmxia/gaia/pkg/generated/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type ClientManager struct {
	K8sClient        *kubernetes.Clientset
	K8sDynamicClient *dynamic.DynamicClient
	Gaiaclient       *gaiaclientset.Clientset
	RestConfig       *restclient.Config
}

var (
	once      sync.Once
	k8sClient *ClientManager
)

func GetClientWithPanic() (*ClientManager, error) {
	once.Do(func() {
		var err error
		k8sClient, err = GetClient()
		if err != nil {
			log.Println("we can't get k8s client" + err.Error())
		}
	})
	if k8sClients == nil {
		return nil, errors.New("can't get k8s client")
	}
	return k8sClient, nil
}

var k8sClients = &sync.Map{} // 并发map

func GetClient() (*ClientManager, error) {
	// By default we get in cluster config.
	//localKubeConfig, err := utils.LoadsKubeConfig("/Users/lumingming/.kube/qingGlobal/gloal", 1)
	localKubeConfig, err := utils.LoadsKubeConfig("", 1)
	if err != nil {
		return nil, err
	}
	localKubeClientSet := kubernetes.NewForConfigOrDie(localKubeConfig)
	localGaiaClientSet := gaiaclientset.NewForConfigOrDie(localKubeConfig)
	localK8sDynamicClient, _ := dynamic.NewForConfig(localKubeConfig)

	return &ClientManager{
		K8sClient:        localKubeClientSet,
		Gaiaclient:       localGaiaClientSet,
		K8sDynamicClient: localK8sDynamicClient,
	}, nil
}

/*
const (
	// High enough QPS to fit all expected use cases.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases.
	defaultBurst = 1e6
	// full resyc cache resource time
	defaultResyncPeriod = 30 * time.Second
)
*/
