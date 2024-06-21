package routers

import (
	"github.com/gin-gonic/gin"
	k8sv1 "github.com/lmxia/nightwatcher/api/v1/k8s"
)

func addK8sRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/k8s")

	router.GET("/pods", k8sv1.GetPods)
	router.GET("/watch/pods", k8sv1.WatchPods)
	router.GET("/pods/:namespace/:podName/ssh", k8sv1.PodWebSSH)
	router.GET("/pods/:namespace/:podName/log", k8sv1.GetPodLog)
	router.GET("/pods/:namespace/:podName/:containerName/download_log", k8sv1.DownloadPodContainerLog)
	router.GET("/pods/:namespace/:podName", k8sv1.GetPod)
	router.DELETE("/pods/:namespace/:podName", k8sv1.DeletePod)

	router.GET("/deployments", k8sv1.GetDeployments)
	router.GET("/deployments/:namespace/:deploymentName", k8sv1.GetDeployment)
	router.POST("/deployments", k8sv1.PostDeployment)
	router.POST("/deployments/:namespace/:deploymentName", k8sv1.DeploymentDoAction)
	router.DELETE("/deployments/:namespace/:deploymentName", k8sv1.DeleteDeployment)
	router.PUT("/deployments/:namespace/:deploymentName", k8sv1.PutDeployment)
	router.PATCH("/deployments/:namespace/:deploymentName", k8sv1.PatchDeployment)
	router.GET("/deployment_status/:namespace/:deploymentName", k8sv1.GetDeploymentStatus)
	router.GET("/deployment_pods/:namespace/:deploymentName", k8sv1.GetDeploymentPods)

	router.GET("/services", k8sv1.GetServices)
	router.GET("/services/:namespace/:serviceName", k8sv1.GetService)

	router.GET("/jobs", k8sv1.GetJobs)
	router.GET("/jobs/:namespace/:jobName", k8sv1.GetJob)
	router.DELETE("/jobs/:namespace/:jobName", k8sv1.DeleteJob)

	router.GET("/cronjobs", k8sv1.GetCronJobs)
	router.POST("/cronjobs", k8sv1.PostCronJob)
	router.GET("/cronjobs/:namespace/:cronjobName", k8sv1.GetCronJob)
	router.PUT("/cronjobs/:namespace/:cronjobName", k8sv1.PutCronJob)
	router.DELETE("/cronjobs/:namespace/:cronjobName", k8sv1.DeleteCronJob)

	router.GET("/events", k8sv1.GetEvents)

	router.GET("/nodes", k8sv1.GetNodes)
	router.GET("/namespaces", k8sv1.GetNamespaces)

	router.GET("/configmaps/:namespace/:name", k8sv1.GetConfigmap)
	router.PUT("/configmaps/:namespace/:name", k8sv1.PutConfigmap)
	router.DELETE("/configmaps/:namespace/:name", k8sv1.DeleteConfigmap)
}
