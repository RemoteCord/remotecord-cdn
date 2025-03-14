package router

import (
	"net/http"

	. "cdn/api/util"

	"github.com/gin-gonic/gin"
)

func pingEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func InitRoutes() {




	// Configure primary router
	router := gin.Default()
	router.MaxMultipartMemory = int64(Config.FileUploadLimit) << 20

	// Setup routes & paths
	api := router.Group("/api")
	{
		api.GET("/ping", pingEndpoint)
	}

	api.Group("/cdn")
	{
		api.POST("/upload/:uploadtoken", uploadEndpoint)
		api.GET("/download/:clientid", getFileEndpoint)
		api.GET("/upload-endpoint", getUploadEndpoint)
	}

	// router.Static("/download", "./uploads/images")

	// Run router
	router.Run(":" + Config.Port)
}