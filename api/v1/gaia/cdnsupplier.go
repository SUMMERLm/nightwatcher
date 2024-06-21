package gaia

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"github.com/lmxia/nightwatcher/app"
	"github.com/lmxia/nightwatcher/controllers/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SupplierName struct {
	SupplierName string `SupplierName:"label" binding:"required"`
}

type CdnSupplier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CdnSupplierSpec `json:"spec"`
	Status            SupplierStatus  `json:"status,omitempty"`
}

type CdnSupplierSpec struct {
	// +required
	SupplierName string `json:"supplierName,omitempty"`
	// +required
	CloudAccessKeyid string `json:"cloudAccessKeyid,omitempty"`
	// +required
	CloudAccessKeysecret string `json:"cloudAccessKeysecret,omitempty"`
}

type SupplierStatus struct{}

// CdnSupplierList contains a list of CdnSupplier
type CdnSupplierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CdnSupplier `json:"items"`
}

// CreateCDNSupplier 用于创建一个新的CDN Provider
//
//		@Tags		CDNSupplier
//		@Summary	用于创建一个新的CDN Provider
//		@accept		application/json
//		@Produce	application/json
//	 @Param		RequestBody	 body	gaia.CdnSupplierSpec	true	"RequestBody"
//		@Success	200		{object}	app.Response
//		@Failure	500		{object}	app.Response
//		@Router		/gaia/cdnsuppliers/new [post]
func CreateCDNSupplier(c *gin.Context) {
	appG := app.Gin{C: c}
	k8sClients, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}

	var spec CdnSupplierSpec
	if errPare := c.BindJSON(&spec); errPare != nil {
		appG.Fail(http.StatusInternalServerError, errPare, nil)
		return
	}
	cdnGvr := schema.GroupVersionResource{Group: "apps.gaia.io", Version: "v1alpha1", Resource: "cdnsuppliers"}

	resultCdn, err := k8sClients.K8sDynamicClient.Resource(cdnGvr).Namespace("gaia-frontend").Get(context.TODO(), spec.SupplierName, metav1.GetOptions{})
	klog.V(1).Infof("resultCDN msg is %s", resultCdn)
	//cdnSupplier := CdnSupplier{}
	//if errUnstructured := runtime.DefaultUnstructuredConverter.FromUnstructured(resultCdn.UnstructuredContent(), cdnSupplier); errUnstructured != nil {
	//	appG.Fail(http.StatusInternalServerError, err, nil)
	//}
	if resultCdn == nil {
		resultCdn := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps.gaia.io/v1alpha1",
				"kind":       "CdnSupplier",
				"metadata": map[string]interface{}{
					"name":      spec.SupplierName,
					"namespace": "gaia-frontend",
					"labels": map[string]interface{}{
						"cdnsupplier/aliyun": spec.SupplierName,
					},
				},
				"spec": map[string]interface{}{
					"supplierName":         spec.SupplierName,
					"cloudAccessKeyid":     spec.CloudAccessKeyid,
					"cloudAccessKeysecret": spec.CloudAccessKeysecret,
				},
			},
		}
		_, newCdnErr := k8sClients.K8sDynamicClient.Resource(cdnGvr).Namespace("gaia-frontend").Create(context.TODO(), resultCdn, metav1.CreateOptions{})
		if newCdnErr != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
		}
	}
	return
}

// DeleteCDNSupplier 用于删除一个CDN Provider
//
//	@Tags		CDNSupplier
//	@Summary	用于删除一个新的CDN Provider
//	@accept		application/json
//	@Produce	application/json
//	@Param		RequestBody	 body	gaia.SupplierName	true	"RequestBody"
//	@Success	200		{object}	app.Response
//	@Failure	500		{object}	app.Response
//	@Router		/gaia/cdnsuppliers/recycle [post]
func DeleteCDNSupplier(c *gin.Context) {
	appG := app.Gin{C: c}
	k8sClients, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	var supplierName SupplierName
	if errPare := c.BindJSON(&supplierName); errPare != nil {
		appG.Fail(http.StatusInternalServerError, errPare, nil)
		return
	}
	cdnGvr := schema.GroupVersionResource{Group: "apps.gaia.io", Version: "v1alpha1", Resource: "cdnsuppliers"}

	resultCdn, err := k8sClients.K8sDynamicClient.Resource(cdnGvr).Namespace("gaia-frontend").Get(context.TODO(), supplierName.SupplierName, metav1.GetOptions{})
	klog.V(1).Infof("resultCDN msg is %s", resultCdn)
	if err == nil {
		deleteCdnErr := k8sClients.K8sDynamicClient.Resource(cdnGvr).Namespace("gaia-frontend").Delete(context.TODO(), resultCdn.GetName(), metav1.DeleteOptions{})
		if deleteCdnErr != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
		}
	}
	return
}

// UpdateCDNSupplier 用于更新一个CDN Provider
//
// @Tags		CDNSupplier
// @Summary	用于更新的CDN Provider
// @accept		application/json
// @Produce	application/json
// @Param		RequestBody	 body	gaia.CdnSupplierSpec	true	"RequestBody"
// @Success	200		{object}	app.Response
// @Failure	500		{object}	app.Response
// @Router		/gaia/cdnsuppliers/update [post]
func UpdateCDNSupplier(c *gin.Context) {
	appG := app.Gin{C: c}
	k8sClients, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}

	var spec CdnSupplierSpec
	if errPare := c.BindJSON(&spec); errPare != nil {
		appG.Fail(http.StatusInternalServerError, errPare, nil)
		return
	}

	cdnGvr := schema.GroupVersionResource{Group: "apps.gaia.io", Version: "v1alpha1", Resource: "cdnsuppliers"}
	resultCdnLocal, err := k8sClients.K8sDynamicClient.Resource(cdnGvr).Namespace("gaia-frontend").Get(context.TODO(), spec.SupplierName, metav1.GetOptions{})
	klog.V(1).Infof("resultCDN msg is %s", resultCdnLocal)
	resultCdn := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps.gaia.io/v1alpha1",
			"kind":       "CdnSupplier",
			"metadata": map[string]interface{}{
				"name":      spec.SupplierName,
				"namespace": "gaia-frontend",
				"labels": map[string]interface{}{
					"cdnsupplier/aliyun": spec.SupplierName,
				},
				"resourceVersion": resultCdnLocal.GetResourceVersion(),
			},
			"spec": map[string]interface{}{
				"supplierName":         spec.SupplierName,
				"cloudAccessKeyid":     spec.CloudAccessKeyid,
				"cloudAccessKeysecret": spec.CloudAccessKeysecret,
			},
		},
	}
	_, updateCdnErr := k8sClients.K8sDynamicClient.Resource(cdnGvr).Namespace("gaia-frontend").Update(context.TODO(), resultCdn, metav1.UpdateOptions{})
	if updateCdnErr != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
	}
	return
}
