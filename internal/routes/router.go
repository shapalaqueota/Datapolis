package routes

import (
	"Datapolis/internal/handlers"
	"Datapolis/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func Router(
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	geoJSONHandler *handlers.GeoJSONHandler) *gin.Engine {

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/login", authHandler.Login)
	router.POST("/refresh", authHandler.RefreshToken)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/renovation")
	}

	geojson := protected.Group("/geojson")
	{
		// Коллекции - только чтение для обычных пользователей
		collections := geojson.Group("/collections")
		{
			collections.GET("", geoJSONHandler.GetCollections)
			collections.GET("/:id", geoJSONHandler.GetCollection)
			collections.GET("/:id/export", geoJSONHandler.ExportGeoJSON)
			collections.GET("/:id/features", geoJSONHandler.GetFeatures)
		}
	}

	admin := protected.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.POST("/register", userHandler.Register)
		admin.GET("/users", userHandler.GetUsers)
		admin.GET("/users/:id", userHandler.GetUser)

		adminGeoJSON := admin.Group("/geojson")
		{
			// Коллекции - администраторы могут создавать и удалять
			adminCollections := adminGeoJSON.Group("/collections")
			{
				adminCollections.POST("", geoJSONHandler.UploadGeoJSON)
				adminCollections.DELETE("/:id", geoJSONHandler.DeleteCollection)
				adminCollections.POST("/:id/features", geoJSONHandler.AddFeature)
			}

			// Фичи - администраторы могут обновлять и удалять
			adminFeatures := adminGeoJSON.Group("/features")
			{
				adminFeatures.PUT("/:id", geoJSONHandler.UpdateFeature)
				adminFeatures.DELETE("/:id", geoJSONHandler.DeleteFeature)
			}
		}
	}

	return router
}
