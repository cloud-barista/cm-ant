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
				"ns_id = ? AND mcis_id = ? AND vm_id = ? AND username = ? AND agent_type = ?",
				param.NsId, param.McisId, param.VmId, param.Username, param.AgentType,
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

func (r *LoadRepository) DeleteAgentInstallInfoStatusTx(ctx context.Context, param *MonitoringAgentInfo) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Delete(&param).
			Error
	})

	return err

}

func (r *LoadRepository) GetPagingMonitoringAgentInfosTx(ctx context.Context, param GetAllMonitoringAgentInfosParam) ([]MonitoringAgentInfo, int64, error) {
	var monitoringAgentInfos []MonitoringAgentInfo
	var totalRows int64

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&monitoringAgentInfos)

		if param.NsId != "" {
			q = q.Where("ns_id = ?", param.NsId)
		}

		if param.McisId != "" {
			q = q.Where("mcis_id = ?", param.McisId)
		}

		if param.VmId != "" {
			q = q.Where("vm_id = ?", param.VmId)
		}

		if err := q.Count(&totalRows).Error; err != nil {
			return err
		}

		offset := (param.Page - 1) * param.Size
		if err := q.Offset(offset).Limit(param.Size).Find(&monitoringAgentInfos).Error; err != nil {
			return err
		}

		return nil
	})

	return monitoringAgentInfos, totalRows, err

}

func (r *LoadRepository) GetAllMonitoringAgentInfosTx(ctx context.Context, param MonitoringAgentInstallationParams) ([]MonitoringAgentInfo, error) {
	var monitoringAgentInfos []MonitoringAgentInfo

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&monitoringAgentInfos)

		if param.NsId != "" {
			q = q.Where("ns_id = ?", param.NsId)
		}

		if param.McisId != "" {
			q = q.Where("mcis_id = ?", param.McisId)
		}

		if param.VmIds != nil && len(param.VmIds) > 0 {
			q = q.Where("vm_id IN (?)", param.VmIds)
		}

		if err := q.Find(&monitoringAgentInfos).Error; err != nil {
			return err
		}

		return nil
	})

	return monitoringAgentInfos, err
}

func (r *LoadRepository) InsertLoadGeneratorInstallInfoTx(ctx context.Context, param *LoadGeneratorInstallInfo) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Where(
				"install_location = ? AND install_type = ? AND install_path = ? AND install_version = ? AND status = ?",
				param.InstallLocation, param.InstallType, param.InstallPath, param.InstallVersion, "completed",
			).
			FirstOrCreate(
				param,
			).Error
	})

	return err

}

func (r *LoadRepository) UpdateLoadGeneratorInstallInfoTx(ctx context.Context, param *LoadGeneratorInstallInfo) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Model(param).
			Save(param).
			Error
	})

	return err

}
