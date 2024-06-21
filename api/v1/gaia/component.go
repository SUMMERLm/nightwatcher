package gaia

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	appv1alpha1 "github.com/lmxia/gaia/pkg/apis/apps/v1alpha1"
	"github.com/lmxia/gaia/pkg/common"
	"github.com/lmxia/nightwatcher/app"
	"github.com/lmxia/nightwatcher/controllers/k8s"
	"k8s.io/apimachinery/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ComponentUri struct {
	Name     string `form:"name" json:"name" binding:"required"`
	DescName string `form:"descName" json:"descName" binding:"-"`
}

// @Tags		Component
// @Summary	查看全部具体某个component所在的field
// @accept		application/json
// @Produce	application/json
// @Param		name		query		string	true	"Name"
// @Param		descName	query		string	false	"DescName"
// @Success	200			{object}	app.Response
// @Failure	500			{object}	app.Response
// @Router		/gaia/component/field/query [get]
func GetComponentField(c *gin.Context) {
	appG := app.Gin{C: c}

	var u ComponentUri

	if err := appG.C.ShouldBind(&u); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}

	if len(u.Name) == 0 {
		appG.Fail(http.StatusBadRequest, errors.New("cannot get the params: name"), nil)
	} else {
		k8sClient, err := k8s.GetClientWithPanic()
		if err != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
			return
		}
		var rbs *appv1alpha1.ResourceBindingList
		if len(u.DescName) != 0 {
			// query according to component name and description name
			rbs, err = k8sClient.Gaiaclient.AppsV1alpha1().ResourceBindings(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{
				common.GaiaDescriptionLabel: u.DescName,
				common.StatusScheduler:      string(appv1alpha1.ResourceBindingSelected),
			}).String()})
		} else {
			// query according to component name
			rbs, err = k8sClient.Gaiaclient.AppsV1alpha1().ResourceBindings(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{
				common.StatusScheduler: string(appv1alpha1.ResourceBindingSelected),
			}).String()})
		}

		if err != nil {
			appG.Fail(http.StatusInternalServerError, err, nil)
			return
		}

		var field []string
		for _, rb := range rbs.Items {
			for _, rbApp := range rb.Spec.RbApps {
				if v, ok := rbApp.Replicas[u.Name]; ok && v > 0 {
					field = append(field, rbApp.ClusterName)
				}
			}
		}
		if len(field) == 0 {
			appG.Fail(http.StatusNotFound, errors.New("component not found"), nil)
			return
		}

		appG.Success(http.StatusOK, "ok", field)
	}
}
