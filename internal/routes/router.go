package routes

import (
	"Datapolis/internal/handlers"
	"Datapolis/internal/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func Router(
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	geoJSONHandler *handlers.GeoJSONHandler) *gin.Engine {

	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/sign-in", authHandler.Login)
	router.POST("/refresh", authHandler.RefreshToken)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/renovation")
	}

	geojson := protected.Group("/geojson")
	{
		collections := geojson.Group("/collections")
		{
			collections.GET("", geoJSONHandler.GetAllCollections)
			collections.GET("/:id", geoJSONHandler.GetCollection)
			collections.GET("/:id/features", geoJSONHandler.GetFeatures)
		}
	}

	admin := protected.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.POST("/sign-up", userHandler.Register)
		admin.POST("/register", userHandler.Register) // Added this line to support both routes
		admin.GET("/users", userHandler.GetUsers)
		admin.GET("/users/:id", userHandler.GetUser)
		admin.PUT("/users/update/:id", userHandler.UpdateUser)
		admin.PUT("/users/update-password/:id", userHandler.UpdatePassword)

		adminGeoJSON := admin.Group("/geojson")
		{
			adminCollections := adminGeoJSON.Group("/collections")
			{
				adminCollections.POST("", geoJSONHandler.UploadGeoJSONBulk)
				adminCollections.DELETE("/:id", geoJSONHandler.DeleteCollection)
				adminCollections.POST("/:id/features", geoJSONHandler.AddSingleFeature)

			}
			adminFeatures := adminGeoJSON.Group("/features")
			{
				adminFeatures.PUT("/:id", geoJSONHandler.UpdateFeature)
				adminFeatures.DELETE("/:id", geoJSONHandler.DeleteFeature)
			}
		}
	}

	return router
}
