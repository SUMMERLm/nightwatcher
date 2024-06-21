package utils

import (
	"errors"

	"github.com/lmxia/gaia/pkg/common"
	"github.com/lmxia/gaia/pkg/generated/listers/apps/v1alpha1"
	"k8s.io/klog/v2"
)

func GetComponentFromDescriptionAndFQDN(lister v1alpha1.DescriptionLister, descName, fqdn string) (string, error) {
	cachedDesc, err := lister.Descriptions(common.GaiaReservedNamespace).Get(descName)
	if err != nil {
		klog.Errorf("can't get description from: %s/%s", common.GaiaReservedNamespace, descName)
		return "", err
	}
	for _, item := range cachedDesc.Spec.WorkloadComponents {
		if item.FQDN == fqdn {
			return item.ComponentName, nil
		}
	}

	return "", errors.New("can't find component match that fqdn")
}
