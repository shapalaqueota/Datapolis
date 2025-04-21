package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// GeoJSONCollection представляет коллекцию GeoJSON данных
type GeoJSONCollection struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"default:'FeatureCollection'"`
	CRS         JSONData  `json:"crs" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	UserID      int       `json:"user_id" gorm:"not null"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
}

// GeoJSONFeature представляет отдельный объект геоданных
type GeoJSONFeature struct {
	ID           int               `json:"id" gorm:"primaryKey"`
	Type         string            `json:"type" gorm:"default:'Feature'"`
	Properties   JSONData          `json:"properties" gorm:"type:jsonb"`
	Geometry     JSONData          `json:"geometry" gorm:"type:jsonb"`
	CollectionID int               `json:"collection_id" gorm:"not null"`
	Collection   GeoJSONCollection `json:"-" gorm:"foreignKey:CollectionID"`
	CreatedAt    time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

// JSONData представляет тип для хранения JSON данных
type JSONData json.RawMessage

// Value реализует интерфейс driver.Valuer
func (j JSONData) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan реализует интерфейс sql.Scanner
func (j *JSONData) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("тип не может быть преобразован в JSONData")
	}

	*j = JSONData(bytes)
	return nil
}

// MarshalJSON реализует интерфейс json.Marshaler
func (j JSONData) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler
func (j *JSONData) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONData: UnmarshalJSON на nil указателе")
	}
	*j = data
	return nil
}
