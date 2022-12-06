package database

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Model a basic GoLang struct which includes the following fields: ID, CreatedAt, UpdatedAt, DeletedAt
// It may be embedded into your model or you may build your own model without it
//
//	type User struct {
//	  database.Model
//	}
type Model struct {
	ID        int64                 `json:"id"         gorm:"primarykey"`
	CreatedAt int64                 `json:"created_at"`
	UpdatedAt int64                 `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"index"`
}

const (
	// DefaultPageSize pagination default items limit size
	DefaultPageSize = 30

	// MaxPageSize pagination default items max limit size
	MaxPageSize = 1000
)

// Pagination model
type Pagination[T any] struct {
	Page     int   `json:"page"      form:"page"      query:"page"`
	PageSize int   `json:"page_size" form:"page_size" query:"page_size"`
	Total    int64 `json:"total"`
	Items    []T   `json:"items"`
}

// Paginate callback
func (p *Pagination[T]) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if p.Page <= 0 {
			p.Page = 1
		}

		switch {
		case p.PageSize > MaxPageSize:
			p.PageSize = MaxPageSize
		case p.PageSize <= 0:
			p.PageSize = DefaultPageSize
		}

		offset := (p.Page - 1) * p.PageSize

		return db.Offset(offset).Limit(p.PageSize)
	}
}
