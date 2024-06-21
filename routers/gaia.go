package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/lmxia/nightwatcher/api/v1/gaia"
)

func addGaiaRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/gaia")
	router.GET("/clusters", gaia.GetClusters)
	router.GET("/clusters/:namespace/:cluster/:label", gaia.GetClusterLabel)
	router.GET("/component/field/query", gaia.GetComponentField)

	router.POST("/resourcebinding/gaia-merged/:rbName", gaia.UpdateSelectedRB)

	router.POST("/cdnsuppliers/new", gaia.CreateCDNSupplier)
	router.POST("/cdnsuppliers/recycle", gaia.DeleteCDNSupplier)
	router.POST("/cdnsuppliers/update", gaia.UpdateCDNSupplier)
	
	descRouter := router.Group("/descriptions")
	{
		descRouter.GET("", gaia.GetDescriptions)
		descRouter.GET("/:namespace/:descName", gaia.GetDescription)
		descRouter.GET("/status", gaia.GetDescriptionStatus)
		descRouter.GET("/scn", gaia.ScnsIsInTheSameDesc)
		descRouter.DELETE("/gaia-reserved/:descName", gaia.DeleteDescription)
	}
}
