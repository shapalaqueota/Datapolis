package main

import (
	"Datapolis/internal/db"
	"Datapolis/internal/handlers"
	"Datapolis/internal/repository"
	"Datapolis/internal/routes"
	service "Datapolis/internal/services"
	"log"
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

	router := routes.Router(userHandler, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
