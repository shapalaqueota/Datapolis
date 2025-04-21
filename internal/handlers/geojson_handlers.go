package handlers

import (
	"Datapolis/internal/models"
	service "Datapolis/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type GeoJSONHandler struct {
	geoJSONService *service.GeoJSONService
}

// конструктор
func NewGeoJSONHandler(geoJSONService *service.GeoJSONService) *GeoJSONHandler {
	return &GeoJSONHandler{geoJSONService: geoJSONService}
}

// UploadGeoJSON загружает GeoJSON файл и создает новую коллекцию
func (h *GeoJSONHandler) UploadGeoJSON(c *gin.Context) {
	// Получаем ID пользователя из контекста (используя ключ как в middleware)
	userID, _ := c.Get("user_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}
	description := c.PostForm("description")

	collection, err := h.geoJSONService.ImportGeoJSON(c.Request.Context(), file, name, description, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при импорте GeoJSON: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, collection)
}

// GetCollections получает список коллекций пользователя
func (h *GeoJSONHandler) GetCollections(c *gin.Context) {
	userID, _ := c.Get("user_id")
	collections, err := h.geoJSONService.GetCollectionsByUser(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении коллекций: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, collections)
}

// GetCollection получает коллекцию по ID
func (h *GeoJSONHandler) GetCollection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	collection, err := h.geoJSONService.GetCollection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении коллекции: " + err.Error()})
		return
	}

	if collection == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Коллекция не найдена"})
		return
	}

	c.JSON(http.StatusOK, collection)
}

// ExportGeoJSON экспортирует коллекцию в GeoJSON формат
func (h *GeoJSONHandler) ExportGeoJSON(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	geoJSON, err := h.geoJSONService.ExportGeoJSON(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при экспорте GeoJSON: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Writer.Write(geoJSON)
}

// DeleteCollection удаляет коллекцию
func (h *GeoJSONHandler) DeleteCollection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	userID, _ := c.Get("user_id") // Исправлено с "userID" на "user_id"
	err = h.geoJSONService.DeleteCollection(c.Request.Context(), id, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении коллекции: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetFeatures получает все фичи коллекции
func (h *GeoJSONHandler) GetFeatures(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	features, err := h.geoJSONService.GetFeatures(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении фич: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, features)
}

// AddFeature добавляет новую фичу в коллекцию
func (h *GeoJSONHandler) AddFeature(c *gin.Context) {
	var feature models.GeoJSONFeature
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	collectionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID коллекции"})
		return
	}
	feature.CollectionID = collectionID

	err = h.geoJSONService.AddFeature(c.Request.Context(), &feature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении фичи: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, feature)
}

// UpdateFeature обновляет фичу
func (h *GeoJSONHandler) UpdateFeature(c *gin.Context) {
	var feature models.GeoJSONFeature
	if err := c.ShouldBindJSON(&feature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	featureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID фичи"})
		return
	}
	feature.ID = featureID

	err = h.geoJSONService.UpdateFeature(c.Request.Context(), &feature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении фичи: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, feature)
}

// DeleteFeature удаляет фичу
func (h *GeoJSONHandler) DeleteFeature(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	err = h.geoJSONService.DeleteFeature(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении фичи: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
