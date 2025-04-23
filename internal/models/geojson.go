package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// GeoJSONCollection holds high‑level metadata only. Geometry lives in geo_features.
type GeoJSONCollection struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SRID        int       `json:"srid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      int       `json:"user_id"`
}

type GeoJSONFeature struct {
	ID           int       `json:"id"`
	Type         string    `json:"type"`
	Properties   JSONData  `json:"properties"`
	Geometry     JSONData  `json:"geometry"`
	CollectionID int       `json:"collection_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type JSONData json.RawMessage

func (j JSONData) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONData) Scan(val interface{}) error {
	if val == nil {
		*j = nil
		return nil
	}
	switch v := val.(type) {
	case []byte:
		*j = JSONData(v)
	case string:
		*j = JSONData([]byte(v))
	default:
		return errors.New("unsupported type for JSONData")
	}
	return nil
}

func (j *JSONData) UnmarshalJSON(data []byte) error {
	*j = JSONData(data)
	return nil
}

func (j JSONData) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}
