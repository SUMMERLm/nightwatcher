package gaia

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmxia/nightwatcher/app"
	"github.com/lmxia/nightwatcher/controllers/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceBindingUri struct {
	Namespace string `uri:"namespace" binding:"-"`
	RBName    string `uri:"rbName" binding:"required"`
}

// UpdateSelectedRB update status from merged to selected
//
//	@Tags		ResourceBinding
//	@Summary	选择某个聚合后的调度方案 rb
//	@accept		application/json
//	@Produce	application/json
//	@Param		rbName	path		string	true	"RBName"
//	@Success	200		{object}	app.Response
//	@Failure	500		{object}	app.Response
//	@Router		/gaia/resourcebinding/gaia-merged/{rbName} [post]
func UpdateSelectedRB(c *gin.Context) {
	appG := app.Gin{C: c}
	var u ResourceBindingUri
	if err := appG.C.ShouldBindUri(&u); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}
	k8sClients, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	rb, err := k8sClients.Gaiaclient.AppsV1alpha1().ResourceBindings("gaia-merged").Get(context.TODO(), u.RBName, metav1.GetOptions{})
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	rbCopy := rb.DeepCopy()
	rbCopy.Spec.StatusScheduler = "selected"
	_, err = k8sClients.Gaiaclient.AppsV1alpha1().ResourceBindings("gaia-merged").Update(context.TODO(), rbCopy, metav1.UpdateOptions{})
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}

	appG.Success(http.StatusOK, "ok", u.RBName)
}
