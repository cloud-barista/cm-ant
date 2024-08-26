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
				"ns_id = ? AND mci_id = ? AND vm_id = ? AND username = ? AND agent_type = ?",
				param.NsId, param.MciId, param.VmId, param.Username, param.AgentType,
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

		if param.MciId != "" {
			q = q.Where("mci_id = ?", param.MciId)
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
		q := d.Model(&monitoringAgentInfos).
			Order("monitoring_agent_info.created_at desc")

		if param.NsId != "" {
			q = q.Where("ns_id = ?", param.NsId)
		}

		if param.MciId != "" {
			q = q.Where("mci_id = ?", param.MciId)
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
		err := d.
			Where(
				"install_location = ? AND install_type = ? AND install_path = ? AND install_version = ? AND status = ?",
				param.InstallLocation, param.InstallType, param.InstallPath, param.InstallVersion, "installed",
			).
			// Omit("LoadGeneratorServers").
			FirstOrCreate(param).Error

		if err != nil {
			return err
		}

		return nil
	})

	return err

}

func (r *LoadRepository) UpdateLoadGeneratorInstallInfoTx(ctx context.Context, param *LoadGeneratorInstallInfo) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		// Update() 와 동일한 방식으로 작동
		// key 충돌이 없는 경우 그냥 save 하기 때문에 association 이 insert 됨
		// return d.
		// 	Model(param).
		// 	Session(&gorm.Session{FullSaveAssociations: true}).
		// 	Save(param).Error

		// updates 의 경우 foreign key update 없이 키 충돌이 나는 경우 모든 필드 업데이트
		// key 충돌이 없는 경우 insert association 이 증가
		// return d.
		// 	Model(param).
		// 	Session(&gorm.Session{FullSaveAssociations: true}).
		// 	Updates(param).
		// 	Error

		// Association().Replace() 는 update 시 id 충돌의 경우 foreign key 를 null 로 업데이트
		return d.Model(param).
			Session(&gorm.Session{FullSaveAssociations: true}).
			Association("LoadGeneratorServers").
			Replace(param.LoadGeneratorServers)

	})

	return err

}

func (r *LoadRepository) GetValidLoadGeneratorInstallInfoByIdTx(ctx context.Context, loadGeneratorInstallInfoId uint) (LoadGeneratorInstallInfo, error) {
	var loadGeneratorInstallInfo LoadGeneratorInstallInfo

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Model(loadGeneratorInstallInfo).
			Preload("LoadGeneratorServers").
			First(&loadGeneratorInstallInfo, " id = ? AND status = ?", loadGeneratorInstallInfoId, "installed").
			Error
	})

	return loadGeneratorInstallInfo, err
}

func (r *LoadRepository) GetPagingLoadGeneratorInstallInfosTx(ctx context.Context, param GetAllLoadGeneratorInstallInfoParam) ([]LoadGeneratorInstallInfo, int64, error) {
	var loadGeneratorInstallInfos []LoadGeneratorInstallInfo
	var totalRows int64

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&LoadGeneratorInstallInfo{}).
			Preload("LoadGeneratorServers").
			Order("load_generator_install_infos.created_at desc")

		if param.Status != "" {
			q = q.Where("status = ?", param.Status)
		}

		if err := q.Count(&totalRows).Error; err != nil {
			return err
		}

		offset := (param.Page - 1) * param.Size
		return q.Offset(offset).
			Limit(param.Size).
			Find(&loadGeneratorInstallInfos).Error
	})

	return loadGeneratorInstallInfos, totalRows, err

}

func (r *LoadRepository) InsertLoadTestExecutionStateTx(ctx context.Context, param *LoadTestExecutionState) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		err := d.
			Where(
				"load_generator_install_info_id = ? AND load_test_key = ?",
				param.LoadGeneratorInstallInfoId, param.LoadTestKey,
			).
			FirstOrCreate(param).Error

		if err != nil {
			return err
		}

		return nil
	})

	return err

}

func (r *LoadRepository) SaveForLoadTestExecutionTx(ctx context.Context, loadParam *LoadTestExecutionInfo, stateParam *LoadTestExecutionState) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		err := d.
			Where(
				"load_test_key = ?", loadParam.LoadTestKey,
			).
			FirstOrCreate(loadParam).Error

		if err != nil {
			return err
		}

		stateParam.LoadTestExecutionInfoId = loadParam.ID
		err = d.
			Where("load_test_key = ?", stateParam.LoadTestKey).
			FirstOrCreate(&stateParam).
			Error

		if err != nil {
			return err
		}

		return nil
	})

	return err

}

func (r *LoadRepository) UpdateLoadTestExecutionStateTx(ctx context.Context, param *LoadTestExecutionState) error {

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.
			Model(param).
			Save(param).
			Error
	})

	return err
}

func (r *LoadRepository) UpdateLoadTestExecutionInfoDuration(ctx context.Context, loadTestKey, compileDuration, executionDuration string) error {
	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		err := d.
			Model(&LoadTestExecutionInfo{}).
			Where(
				"load_test_key = ?", loadTestKey,
			).
			Updates(map[string]interface{}{"compile_duration": compileDuration, "execution_duration": executionDuration}).Error

		if err != nil {
			return err
		}

		return nil
	})

	return err

}

func (r *LoadRepository) GetPagingLoadTestExecutionStateTx(ctx context.Context, param GetAllLoadTestExecutionStateParam) ([]LoadTestExecutionState, int64, error) {
	var loadTestExecutionStates []LoadTestExecutionState
	var totalRows int64

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&LoadTestExecutionState{}).
			Preload("LoadGeneratorInstallInfo").
			Preload("LoadGeneratorInstallInfo.LoadGeneratorServers").
			Order("load_test_execution_states.created_at desc")

		if param.LoadTestKey != "" {
			q = q.Where("load_test_key like ?", "%"+param.LoadTestKey+"%")
		}

		if param.ExecutionStatus != "" {
			q = q.Where("execution_status = ?", param.ExecutionStatus)
		}

		if err := q.Count(&totalRows).Error; err != nil {
			return err
		}

		offset := (param.Page - 1) * param.Size
		if err := q.Offset(offset).Limit(param.Size).Find(&loadTestExecutionStates).Error; err != nil {
			return err
		}

		return nil
	})

	return loadTestExecutionStates, totalRows, err

}

func (r *LoadRepository) GetLoadTestExecutionStateTx(ctx context.Context, param GetLoadTestExecutionStateParam) (LoadTestExecutionState, error) {
	var loadTestExecutionState LoadTestExecutionState

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.Model(&loadTestExecutionState).
			Preload("LoadGeneratorInstallInfo").
			Preload("LoadGeneratorInstallInfo.LoadGeneratorServers").
			First(&loadTestExecutionState, "load_test_execution_states.load_test_key = ?", param.LoadTestKey).
			Error
	})

	return loadTestExecutionState, err
}

func (r *LoadRepository) GetPagingLoadTestExecutionHistoryTx(ctx context.Context, param GetAllLoadTestExecutionInfosParam) ([]LoadTestExecutionInfo, int64, error) {
	var loadTestExecutionInfo []LoadTestExecutionInfo
	var totalRows int64

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		q := d.Model(&LoadTestExecutionInfo{}).
			Preload("LoadTestExecutionState").
			Preload("LoadTestExecutionHttpInfos").
			Preload("LoadGeneratorInstallInfo").
			Preload("LoadGeneratorInstallInfo.LoadGeneratorServers").
			Order("load_test_execution_infos.created_at desc")

		if err := q.Count(&totalRows).Error; err != nil {
			return err
		}

		offset := (param.Page - 1) * param.Size
		if err := q.Offset(offset).Limit(param.Size).Find(&loadTestExecutionInfo).Error; err != nil {
			return err
		}

		return nil
	})

	return loadTestExecutionInfo, totalRows, err

}

func (r *LoadRepository) GetLoadTestExecutionInfoTx(ctx context.Context, param GetLoadTestExecutionInfoParam) (LoadTestExecutionInfo, error) {
	var loadTestExecutionInfo LoadTestExecutionInfo

	err := r.execInTransaction(ctx, func(d *gorm.DB) error {
		return d.Model(&loadTestExecutionInfo).
			Preload("LoadTestExecutionState").
			Preload("LoadTestExecutionHttpInfos").
			Preload("LoadGeneratorInstallInfo").
			Preload("LoadGeneratorInstallInfo.LoadGeneratorServers").
			First(&loadTestExecutionInfo, "load_test_execution_infos.load_test_key = ?", param.LoadTestKey).
			Error
	})

	return loadTestExecutionInfo, err
}
