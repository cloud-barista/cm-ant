package cost

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type CostRepository struct {
	db *gorm.DB
}

func NewCostRepository(db *gorm.DB) *CostRepository {
	return &CostRepository{
		db: db,
	}
}

func (r *CostRepository) execInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction error: %w", tx.Error)
	}

	err := fn(tx)
	if err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("rollback error: %v, original error: %w", rbErr, err)
		}
		return err
	}

	return tx.Commit().Error
}
