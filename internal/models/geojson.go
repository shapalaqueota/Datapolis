package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

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

type Pagination struct {
	Page     int `json:"page"`     // текущая страница
	Limit    int `json:"limit"`    // количество записей на странице
	Total    int `json:"total"`    // общее количество записей
	LastPage int `json:"lastPage"` // последняя страница
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

func NewPagination(page, limit int) *Pagination {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 || limit > 500 {
		limit = 500 // максимальное значение
	}

	return &Pagination{
		Page:  page,
		Limit: limit,
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p *Pagination) SetTotal(total int) {
	p.Total = total
	p.LastPage = (total + p.Limit - 1) / p.Limit
	if p.LastPage < 1 {
		p.LastPage = 1
	}
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
