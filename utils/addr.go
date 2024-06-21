package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

type QueryPromResp struct {
	QueryValV model.Vector
	QueryValM model.Matrix
}

// HermesQueryParam hermes查询条件
type HermesQueryParam struct {
	QueryValue string `form:"query_range" json:"query_range" url:"query_range,omitempty"` // 查询具体指标
	StartTime  string `form:"start" json:"start" url:"start,omitempty"`                   // 查询开始时间
	EndTime    string `form:"end" json:"end" url:"end,omitempty"`                         // 查询结束时间
}

func Int64Addr(i int64) *int64 {
	return &i
}

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

// FilterAccessServiceIPFrom  filter out from hermes which fields
func FilterAccessServiceIPFrom(fields sets.Set[string]) ([]string, error) {
	realEndpoints := make([]string, 0)
	param := HermesQueryParam{
		QueryValue: fmt.Sprintf("container_cpu_usage_seconds_total{component_name=\"%s\"}", GetEnvDefault("ACCESS_SERVICE_NAME", "gaia")),
		StartTime:  time.Now().Add(-time.Minute * 1).Format(time.RFC3339Nano),
		EndTime:    time.Now().Format(time.RFC3339Nano),
	}
	v, err := query.Values(param)
	if err != nil {
		klog.Infof("can't parse hermes request params")
		return nil, err
	}
	path := fmt.Sprintf("/query?%s", v.Encode())
	resp, err := NewHttpClient().Call(WithToHermes(), WithPath(path), WithUsePost())
	if err != nil {
		// we can't get from hermes, it's unstable. so give a default access service ip.
		//realEndpoints = append(realEndpoints, GetEnvDefault("ACCESS_SERVICE_DEFAULT_IP", "172.17.2.35"))
		return realEndpoints, err
	}

	defer resp.Body.Close()
	var result QueryPromResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		klog.Errorf("Can't decode result from hermes")
		return nil, err
	}
	for _, item := range result.QueryValM {
		cloneSet := item.Metric.Clone()
		// 当前所属的field名称，在我们查出来的field内
		if fieldName, ok := cloneSet["field_flag"]; ok && fields.Has(string(fieldName)) {
			if nodeIP, ok := cloneSet["node_ip"]; ok && len(nodeIP) != 0 {
				realEndpoints = append(realEndpoints, string(nodeIP))
			}
		}
	}
	return realEndpoints, nil
}
