package v1

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmxia/nightwatcher/app"
	"github.com/lmxia/nightwatcher/controllers/k8s"
	v1 "k8s.io/api/core/v1"
)

// GetConfigmap
//
//	@Summary	获取Configmap资源
//	@accept		application/json
//	@Param		cluster		path		string	true	"Cluster"
//	@Param		namespace	path		string	true	"Namespace"
//	@Param		name		path		string	true	"Name"
//	@Success	200			{object}	app.Response
//	@Failure	500			{object}	app.Response
//	@Router		/k8s/configmaps/{namespace}/{name} [get]
func GetConfigmap(c *gin.Context) {
	appG := app.Gin{C: c}
	param, err := app.GetPathParameterString(c, "namespace", "name")
	if err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}
	k8sClient, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	configMapOperation := k8s.NewConfigmapOperation(k8sClient.K8sClient)
	configMap, err := configMapOperation.Get(context.TODO(), param["namespace"], param["name"])
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	appG.Success(http.StatusOK, "ok", configMap)
}

// PutConfigmap
//
//	@Summary	更新Configmap资源
//	@accept		application/json
//	@Param		cluster		path		string	true	"Cluster"
//	@Param		namespace	path		string	true	"Namespace"
//	@Param		name		path		string	true	"Name"
//	@Success	200			{object}	app.Response
//	@Failure	500			{object}	app.Response
//	@Router		/k8s/configmaps/{namespace}/{name} [put]
func PutConfigmap(c *gin.Context) {
	appG := app.Gin{C: c}
	var configMap v1.ConfigMap
	param, err := app.GetPathParameterString(c, "namespace", "name")
	if err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}
	if err = appG.C.ShouldBind(&configMap); err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}
	k8sClient, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	configMapOperation := k8s.NewConfigmapOperation(k8sClient.K8sClient)
	result, err := configMapOperation.Update(context.TODO(), param["namespace"], param["name"], &configMap)
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	appG.Success(http.StatusOK, "ok", result)
}

// DeleteConfigmap
//
//	@Summary	删除Configmap资源
//	@accept		application/json
//	@Param		cluster		path		string	true	"Cluster"
//	@Param		namespace	path		string	true	"Namespace"
//	@Param		name		path		string	true	"Name"
//	@Success	200			{object}	app.Response
//	@Failure	500			{object}	app.Response
//	@Router		/k8s/configmaps/{namespace}/{name} [delete]
func DeleteConfigmap(c *gin.Context) {
	appG := app.Gin{C: c}
	param, err := app.GetPathParameterString(c, "namespace", "name")
	if err != nil {
		appG.Fail(http.StatusBadRequest, err, nil)
		return
	}
	k8sClient, err := k8s.GetClientWithPanic()
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}

	configMapOperation := k8s.NewConfigmapOperation(k8sClient.K8sClient)
	err = configMapOperation.Delete(context.TODO(), param["namespace"], param["name"])
	if err != nil {
		appG.Fail(http.StatusInternalServerError, err, nil)
		return
	}
	appG.Success(http.StatusOK, "ok", nil)
}
