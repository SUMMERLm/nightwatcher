package plugin

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/coredns/coredns/request"
	appv1alpha1 "github.com/lmxia/gaia/pkg/apis/apps/v1alpha1"
	"github.com/lmxia/gaia/pkg/common"
	"github.com/lmxia/gaia/pkg/generated/listers/apps/v1alpha1"
	"github.com/lmxia/nightwatcher/utils"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

type CrossDNS struct {
	Next     plugin.Handler
	Fall     fall.F
	Zones    []string
	rbLister v1alpha1.ResourceBindingLister
	rbSynced cache.InformerSynced

	indexedStore cache.ThreadSafeStore
	descLister   v1alpha1.DescriptionLister
}

type DNSRecord struct {
	IP string
}

func (c CrossDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := &request.Request{W: w, Req: r}
	qname := state.QName()

	zone := plugin.Zones(c.Zones).Matches(qname)
	if zone == "" {
		klog.Infof("Request does not match configured zones %v", c.Zones)
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, state.W, r) // nolint:wrapcheck // Let the caller wrap it.
	}

	klog.Infof("Request received for %q", qname)
	if state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA && state.QType() != dns.TypeSRV {
		msg := fmt.Sprintf("Query of type %d is not supported", state.QType())
		klog.Info(msg)
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, state.W, r) // nolint:wrapcheck // Let the caller wrap it.
	}
	zone = qname[len(qname)-len(zone):] // maintain case of original query
	state.Zone = zone

	pReq, pErr := parseRequest(state)

	if pErr != nil {
		// We only support svc type queries i.e. *.svc.*
		klog.Infof("Can't parse, request %q is not a valid query - err was %v", qname, pErr)
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, state.W, r) // nolint:wrapcheck // Let the caller wrap it.
	}

	return c.getDNSRecord(ctx, zone, state, w, r, pReq)
}

func (c *CrossDNS) getDNSRecord(ctx context.Context, zone string, state *request.Request, w dns.ResponseWriter,
	r *dns.Msg, pReq *recordRequest,
) (int, error) {
	// wait for rb synced.
	if !cache.WaitForCacheSync(ctx.Done(), c.rbSynced) {
		klog.Fatal("unable to sync caches for resource binding")
	}

	// 1. get which desc you belong.
	descNames, err := c.indexedStore.IndexKeys(FQDNINDEX, pReq.hostname)
	if err != nil || len(descNames) == 0 {
		klog.Infof("Couldn't find a scheduled description %q", state.QName())
		return c.emptyResponse(state)
	}
	// 2. get which component the fqdn belong
	componentName, err := utils.GetComponentFromDescriptionAndFQDN(c.descLister, descNames[0], pReq.hostname)
	if err != nil {
		klog.Infof("Get component name failed %s", err)
		return c.emptyResponse(state)
	}

	var dnsRecords []DNSRecord

	// 3 figure out which fields the fqdn were located.
	rbs, err := c.rbLister.ResourceBindings(common.GaiaRBMergedReservedNamespace).List(labels.SelectorFromSet(labels.Set{
		common.GaiaDescriptionLabel: descNames[0],
		// we suppose only fqdn is unique
		common.StatusScheduler: string(appv1alpha1.ResourceBindingSelected)}))
	var fields []string
	for _, rb := range rbs {
		for _, rbApp := range rb.Spec.RbApps {
			if v, ok := rbApp.Replicas[componentName]; ok && v > 0 {
				fields = append(fields, rbApp.ClusterName)
			}
		}
	}
	if len(fields) == 0 || err != nil {
		klog.Errorf("Failed to write message %v", err)
		return dns.RcodeServerFailure, errors.New("failed to write response")
	}

	// 4. Now we get fields, so get fields ip from hermes.
	realEndpoints, err := utils.FilterAccessServiceIPFrom(sets.New[string](fields...))
	if err != nil || len(realEndpoints) == 0 {
		realEndpoints = append(realEndpoints, utils.GetEnvDefault("ACCESS_SERVICE_DEFAULT_IP", "172.17.2.35"))
		klog.Errorf("We can't get real endpoints of access service from these fileds %s, so use default ip.", fields)
	}

	record := getAllRecordsFromField(realEndpoints)
	dnsRecords = append(dnsRecords, record...)
	if len(dnsRecords) == 0 {
		klog.Infof("Couldn't find a connected cluster or valid IPs for %q", state.QName())
		return c.emptyResponse(state)
	}

	records := make([]dns.RR, 0)

	if state.QType() == dns.TypeA {
		records = c.createARecords(dnsRecords, state)
	}

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	rand.Intn(len(records))

	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true

	// all
	//a.Answer = append(a.Answer, records...)
	// random one
	a.Answer = append(a.Answer, records[rand.Intn(len(records))])
	klog.Infof("Responding to query with '%s'", a.Answer)

	wErr := w.WriteMsg(a)
	if wErr != nil {
		// Error writing reply msg
		klog.Errorf("Failed to write message %#v: %v", a, wErr)
		return dns.RcodeServerFailure, errors.New("failed to write response")
	}

	return dns.RcodeSuccess, nil
}

func (c CrossDNS) Name() string {
	return "crossdns"
}

func (c CrossDNS) emptyResponse(state *request.Request) (int, error) {
	a := new(dns.Msg)
	a.SetReply(state.Req)

	return writeResponse(state, a)
}

func writeResponse(state *request.Request, a *dns.Msg) (int, error) {
	a.Authoritative = true

	wErr := state.W.WriteMsg(a)
	if wErr != nil {
		klog.Errorf("Failed to write message %#v: %v", a, wErr)
		return dns.RcodeServerFailure, errors.New("failed to write response")
	}

	return dns.RcodeSuccess, nil
}

func (c CrossDNS) createARecords(dnsrecords []DNSRecord, state *request.Request) []dns.RR {
	records := make([]dns.RR, 0)

	for _, record := range dnsrecords {
		dnsRecord := &dns.A{Hdr: dns.RR_Header{
			Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(),
			Ttl: uint32(5),
		}, A: net.ParseIP(record.IP).To4()}
		records = append(records, dnsRecord)
	}

	return records
}

func getAllRecordsFromField(slices []string) []DNSRecord {
	records := make([]DNSRecord, 0)
	for _, endpoint := range slices {
		record := DNSRecord{
			IP: endpoint,
		}
		records = append(records, record)
	}
	return records
}

var _ plugin.Handler = &CrossDNS{}
