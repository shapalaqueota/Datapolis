package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"Datapolis/internal/models"
	"Datapolis/internal/repository"
)

type GeoJSONService struct {
	repo *repository.GeoJSONRepository
}

func NewGeoJSONService(repo *repository.GeoJSONRepository) *GeoJSONService {
	return &GeoJSONService{repo: repo}
}

// ImportGeoJSON импортирует GeoJSON файл в БД
func (s *GeoJSONService) ImportGeoJSON(ctx context.Context, reader io.Reader, name, description string, userID int) (*models.GeoJSONCollection, error) {
	var geoJSON struct {
		Type     string            `json:"type"`
		Name     string            `json:"name"`
		CRS      json.RawMessage   `json:"crs"`
		Features []json.RawMessage `json:"features"`
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &geoJSON); err != nil {
		return nil, err
	}

	// Создаем коллекцию
	collection := &models.GeoJSONCollection{
		Name:        name,
		Description: description,
		Type:        geoJSON.Type,
		CRS:         models.JSONData(geoJSON.CRS),
		UserID:      userID,
	}

	if err = s.repo.CreateCollection(ctx, collection); err != nil {
		return nil, err
	}

	// Добавляем фичи
	for _, featureData := range geoJSON.Features {
		var feature struct {
			Type       string          `json:"type"`
			Properties json.RawMessage `json:"properties"`
			Geometry   json.RawMessage `json:"geometry"`
		}

		if err = json.Unmarshal(featureData, &feature); err != nil {
			continue
		}

		geoFeature := &models.GeoJSONFeature{
			Type:         feature.Type,
			Properties:   models.JSONData(feature.Properties),
			Geometry:     models.JSONData(feature.Geometry),
			CollectionID: collection.ID,
		}

		if err = s.repo.AddFeature(ctx, geoFeature); err != nil {
			continue
		}
	}

	return collection, nil
}

// GetCollection получает коллекцию по ID
func (s *GeoJSONService) GetCollection(ctx context.Context, id int) (*models.GeoJSONCollection, error) {
	return s.repo.GetCollectionByID(ctx, id)
}

// GetCollectionsByUser получает все коллекции пользователя
func (s *GeoJSONService) GetCollectionsByUser(ctx context.Context, userID int) ([]*models.GeoJSONCollection, error) {
	return s.repo.GetCollectionsByUserID(ctx, userID)
}

// ExportGeoJSON экспортирует коллекцию и все её фичи в GeoJSON формат
func (s *GeoJSONService) ExportGeoJSON(ctx context.Context, collectionID int) ([]byte, error) {
	collection, err := s.repo.GetCollectionByID(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, errors.New("коллекция не найдена")
	}

	features, err := s.repo.GetFeaturesByCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"type": collection.Type,
		"name": collection.Name,
		"crs":  json.RawMessage(collection.CRS),
	}

	featureList := make([]json.RawMessage, 0, len(features))
	for _, feature := range features {
		featureMap := map[string]interface{}{
			"type":       feature.Type,
			"properties": json.RawMessage(feature.Properties),
			"geometry":   json.RawMessage(feature.Geometry),
		}
		featureJSON, err := json.Marshal(featureMap)
		if err != nil {
			continue
		}
		featureList = append(featureList, featureJSON)
	}

	result["features"] = featureList
	return json.Marshal(result)
}

// DeleteCollection удаляет коллекцию
func (s *GeoJSONService) DeleteCollection(ctx context.Context, id, userID int) error {
	return s.repo.DeleteCollection(ctx, id, userID)
}

// GetFeatures получает все фичи коллекции
func (s *GeoJSONService) GetFeatures(ctx context.Context, collectionID int) ([]*models.GeoJSONFeature, error) {
	return s.repo.GetFeaturesByCollectionID(ctx, collectionID)
}

// AddFeature добавляет новую фичу в коллекцию
func (s *GeoJSONService) AddFeature(ctx context.Context, feature *models.GeoJSONFeature) error {
	return s.repo.AddFeature(ctx, feature)
}

// UpdateFeature обновляет фичу
func (s *GeoJSONService) UpdateFeature(ctx context.Context, feature *models.GeoJSONFeature) error {
	return s.repo.UpdateFeature(ctx, feature)
}

// DeleteFeature удаляет фичу
func (s *GeoJSONService) DeleteFeature(ctx context.Context, id int) error {
	feature, err := s.repo.GetFeatureByID(ctx, id)
	if err != nil {
		return err
	}
	if feature == nil {
		return errors.New("фича не найдена")
	}

	return s.repo.DeleteFeature(ctx, id, feature.CollectionID)
}
