package router

import (
	"fmt"
	"net/http"
	"strings"

	. "cdn/api/util"

	"github.com/gin-gonic/gin"
)

func pingEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func InitRoutes() {




	// Configure primary router
	router := gin.Default()
	 router.Use(corsMiddleware())

	//router.MaxMultipartMemory = int64(Config.FileUploadLimit) << 20

	// Setup routes & paths
	api := router.Group("/api")
	{
		api.GET("/ping", pingEndpoint)
	}

	api.Group("/cdn")
	{
		api.POST("/upload/:uploadtoken", uploadEndpoint)
		api.POST("/upload-large/:uploadtoken", uploadLargeEndpoint)

		api.GET("/download/:clientid", getFileEndpoint)
		api.GET("/download-large/:controllerid", getFileLargeEndpoint)

		api.GET("/upload-endpoint", getUploadEndpoint)
	}

	// router.Static("/download", "./uploads/images")

	// Run router
	router.Run(":" + Config.Port)
}

func corsMiddleware() gin.HandlerFunc {
 // Define allowed origins as a comma-separated string
 originsString := "https://preview.luqueee.dev,"
 var allowedOrigins []string
 if originsString != "" {
  // Split the originsString into individual origins and store them in allowedOrigins slice
  allowedOrigins = strings.Split(originsString, ",")
 }

 // Return the actual middleware handler function
 return func(c *gin.Context) {
  // Function to check if a given origin is allowed
  isOriginAllowed := func(origin string, allowedOrigins []string) bool {
	fmt.Println("Origin:", origin)
   for _, allowedOrigin := range allowedOrigins {
    if origin == allowedOrigin {
     return true
    }
   }
   return true
  }

  // Get the Origin header from the request
  origin := c.Request.Header.Get("Origin")

  // Check if the origin is allowed
  if isOriginAllowed(origin, allowedOrigins) {
   // If the origin is allowed, set CORS headers in the response
   c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
   c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
   c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
   c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
  }

  // Handle preflight OPTIONS requests by aborting with status 204
  if c.Request.Method == "OPTIONS" {
   c.AbortWithStatus(204)
   return
  }

  // Call the next handler
  c.Next()
 }
}