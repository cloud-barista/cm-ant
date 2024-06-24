package load

import (
	"gorm.io/gorm"
)

type LoadRepository struct {
	db *gorm.DB
}

func NewLoadRepository(db *gorm.DB) *LoadRepository {
	return &LoadRepository{
		db: db,
	}
}
