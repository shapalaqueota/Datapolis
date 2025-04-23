package main

import (
	"Datapolis/internal/db"
	"Datapolis/internal/handlers"
	"Datapolis/internal/repository"
	"Datapolis/internal/routes"
	service "Datapolis/internal/services"
	"os"
)

func main() {
	db.ConnectDB()
	defer db.Pool.Close()

	userRepo := repository.NewUserRepository(db.Pool)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)
	geoJSONRepo := repository.NewGeoRepository(db.Pool)
	geoJSONService := service.NewGeoService(geoJSONRepo)
	geoJSONHandler := handlers.NewGeoJSONHandler(geoJSONService)

	router := routes.Router(userHandler, authHandler, geoJSONHandler)

	port := os.Getenv("PORT") // ← Heroku сам задаёт эту переменную
	if port == "" {
		port = "8080" // ← это только для запуска локально
	}
	router.Run(":" + port)
}
