package gaia

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmxia/gaia/pkg/common"
	"github.com/lmxia/nightwatcher/app"
	"github.com/lmxia/nightwatcher/controllers/k8s"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DescriptionUri struct {
	Namespace string `uri:"namespace" binding:"-"`
	DescName  string `uri:"descName" binding:"required"`
}

type SCNUri struct {
	ScnID1 string `form:"scn1" json:"scn1" binding:"required"`
	ScnID2 string `form:"scn2" json:"scn2" binding:"required"`
}

type DescriptionStatusUri struct {
	Namespace string `form:"namespace" json:"namespace" binding:"required"`
	Name      string `form:"name" json:"name" binding:"required"`
}

// DeleteDescription
//
//	@Tags		Descriptions
//	@Summary	删除某个部署的description
//	@accept		application/json
//	@Produce	application/json
//	@Param		description	path		string	true	"Description"
//	@Success	200			{object}	app.Response
//	@Failure	500			{object}	app.Response
//	@Router		/gaia/descriptions/gaia-reserved/{description} [delete]
func DeleteDescription(c *gin.Context) {
	appG := app.Gin{C: c}
	var desc DescriptionUri
	if err := appG.C.ShouldBindUri(&desc); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}
	k8sClients, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	err = k8sClients.Gaiaclient.AppsV1alpha1().Descriptions("gaia-reserved").Delete(context.TODO(), desc.DescName, metav1.DeleteOptions{})
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	appG.Success(http.StatusOK, "ok", desc.DescName)
}

// @Summary	查看全部descriptions
// @Tags		Descriptions
// @accept		application/json
// @Produce	application/json
// @Success	200	{object}	v1alpha1.DescriptionList
// @Failure	500	{object}	app.Response
// @Router		/gaia/descriptions [get]
func GetDescriptions(c *gin.Context) {
	appG := app.Gin{C: c}
	k8sClient, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		klog.Infof("get k8sClient failed: %v", err)
		return
	}

	descriptions, err := k8sClient.Gaiaclient.AppsV1alpha1().Descriptions(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		klog.Infof("get desc failed: %v", err)
		return
	}
	appG.Success(http.StatusOK, "ok", descriptions)
}

// @Summary	查看全部具体某个 description
// @Tags		Descriptions
// @accept		application/json
// @Produce	application/json
// @Param		namespace	path		string	true	"Namespace"
// @Param		descName	path		string	true	"DescName"
// @Success	200			{object}	app.Response
// @Failure	500			{object}	app.Response
// @Router		/gaia/descriptions/{namespace}/{descName} [get]
func GetDescription(c *gin.Context) {
	appG := app.Gin{C: c}
	var u DescriptionUri

	if err := appG.C.ShouldBindUri(&u); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}

	k8sClient, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}

	description, err := k8sClient.Gaiaclient.AppsV1alpha1().Descriptions(u.Namespace).Get(context.TODO(), u.DescName, metav1.GetOptions{})
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	appG.Success(http.StatusOK, "ok", description)
}

// @Summary	查看全部具体某个 description 的Status
// @Tags		Descriptions
// @accept		application/json
// @Produce	application/json
// @Param		namespace	query    string    true    "NAMESPACE"
// @Param		name	    query	 string    true    "NAME"
// @Success	200		{object}	app.Response
// @Failure	500		{object}	app.Response
// @Router		/gaia/descriptions/status [get]
func GetDescriptionStatus(c *gin.Context) {
	appG := app.Gin{C: c}
	var u DescriptionStatusUri

	if err := appG.C.ShouldBind(&u); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}

	if len(u.Namespace) != 0 && len(u.Name) != 0 {
		k8sClient, err := k8s.GetClientWithPanic()
		if err != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
			return
		}

		description, err := k8sClient.Gaiaclient.AppsV1alpha1().Descriptions(u.Namespace).Get(context.TODO(), u.Name, metav1.GetOptions{})
		if err != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
			return
		}
		appG.Success(http.StatusOK, "ok", description.Status)

	} else {
		appG.Fail(http.StatusBadRequest, errors.New("cannot get the params: namespace and name"), nil)
	}

}

// ScnIDMap declares a map whose keys are the Scn' ID and the values are null structs
type ScnIDMap map[string]struct{}

// DescScnIDMap declares a map whose keys are the descriptions' names and the values are the ScnIDMap
type DescScnIDMap map[string]ScnIDMap

// @Summary	判断两个scnid是否属于同一个蓝图
// @Tags		Descriptions
// @accept		application/json
// @Produce	application/json
// @Param		scn1	query		string	true	"SCNID1"
// @Param		scn2	query		string	true	"SCNID2"
// @Success	200		{object}	bool
// @Failure	500		{object}	app.Response
// @Router		/gaia/descriptions/scn [get]
func ScnsIsInTheSameDesc(c *gin.Context) {
	appG := app.Gin{C: c}
	var u SCNUri

	if err := appG.C.ShouldBind(&u); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}

	if len(u.ScnID1) != 0 && len(u.ScnID2) != 0 {
		k8sClient, err := k8s.GetClientWithPanic()
		if err != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
			klog.Infof("get k8sClient failed: %v", err)
			return
		}

		descriptions, err := k8sClient.Gaiaclient.AppsV1alpha1().Descriptions(common.GaiaReservedNamespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
			klog.Infof("get desc failed: %v", err)
			return
		}

		if len(descriptions.Items) == 0 {
			appG.Fail(http.StatusInternalServerError, errors.New("description not found whose namespace is 'gaia-reserved'"), nil)
			return
		}
		descScnIDMap := make(DescScnIDMap, len(descriptions.Items))

		for _, desc := range descriptions.Items {
			descScnIDMap[desc.Name] = make(map[string]struct{})
			for _, comn := range desc.Spec.WorkloadComponents {
				for _, container := range comn.Module.Spec.Containers {
					for _, containerEnv := range container.Env {
						if "SCNID" == containerEnv.Name {
							descScnIDMap[desc.Name][containerEnv.Value] = struct{}{}
						}
					}
				}
			}
		}

		var result bool
		for _, scnIDMap := range descScnIDMap {
			_, exist1 := scnIDMap[u.ScnID1]
			_, exist2 := scnIDMap[u.ScnID2]
			if exist1 && exist2 {
				result = true
				break
			} else {
				result = false
			}
		}
		appG.Success(http.StatusOK, "ok", result)
	} else {
		appG.Fail(http.StatusBadRequest, errors.New("cannot get the params: scn1 and scn2"), nil)
	}
}
