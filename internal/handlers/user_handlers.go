package handlers

import (
	"Datapolis/internal/models"
	"Datapolis/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	userService *service.UserService
}

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	log.Printf("Попытка регистрации пользователя: %s", user.Username)

	err := h.userService.Register(c.Request.Context(), &user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidUserData):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			log.Printf("Неизвестная ошибка при регистрации пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		}
		return
	}

	user.Password = ""
	log.Printf("Успешная регистрация пользователя: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно зарегистрирован"})

}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenPair, err := h.authService.Login(c, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Успешный вход",
		"token":              tokenPair.AccessToken,
		"refresh_token":      tokenPair.RefreshToken,
		"expires_in":         tokenPair.ExpiresIn,
		"refresh_expires_in": tokenPair.RefreshExpiresIn,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Токен обновлен",
		"token":              tokenPair.AccessToken,
		"expires_in":         tokenPair.ExpiresIn,
		"refresh_expires_in": tokenPair.RefreshExpiresIn,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	result := struct {
		ID        int       `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, result)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}, len(users))

	for i, user := range users {
		result[i].ID = user.ID
		result[i].Username = user.Username
		result[i].Role = user.Role
	}

	c.JSON(http.StatusOK, result)
}
