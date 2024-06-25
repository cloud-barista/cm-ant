package load

import (
	"context"
	"fmt"

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

func (r *LoadRepository) execInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
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

func (r *LoadRepository) InsertMonitoringAgentInfoTx(ctx context.Context, param *MonitoringAgentInfo) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Where(
				"additional_ns_id = ? AND additional_mcis_id = ? AND username = ? AND agent_type = ?",
				param.AdditionalNsId, param.AdditionalMcisId, param.Username, param.AgentType,
			).
			FirstOrCreate(
				param,
			).Error
	})

	return err

}

func (r *LoadRepository) UpdateAgentInstallInfoStatusTx(ctx context.Context, param *MonitoringAgentInfo) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Model(param).
			Update(
				"status", param.Status,
			).Error
	})

	return err

}
