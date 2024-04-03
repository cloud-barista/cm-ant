package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"strconv"
)

func SaveLoadTestExecution(loadTestReq *api.LoadExecutionConfigReq) (uint, error) {
	db := configuration.DB()
	tx := db.Begin()
	loadEnvId, err := strconv.ParseUint(loadTestReq.EnvId, 10, 64)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	loadExecutionState := model.LoadExecutionState{
		LoadEnvID:       uint(loadEnvId),
		LoadTestKey:     loadTestReq.LoadTestKey,
		ExecutionStatus: constant.Process,
	}

	if err := tx.Model(model.LoadExecutionState{}).
		Where(
			"load_env_id = ? AND load_test_key = ?",
			loadTestReq.EnvId, loadTestReq.LoadTestKey,
		).
		FirstOrCreate(&loadExecutionState).
		Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	loadExecutionConfig := model.LoadExecutionConfig{
		LoadEnvID:   uint(loadEnvId),
		LoadTestKey: loadTestReq.LoadTestKey,
		Threads:     loadTestReq.Threads,
		RampTime:    loadTestReq.RampTime,
		LoopCount:   loadTestReq.LoopCount,
		LoadExecutionHttps: []model.LoadExecutionHttp{
			{
				Method:   loadTestReq.HttpReqs.Method,
				Protocol: loadTestReq.HttpReqs.Protocol,
				Hostname: loadTestReq.HttpReqs.Hostname,
				Port:     loadTestReq.HttpReqs.Port,
				Path:     loadTestReq.HttpReqs.Path,
				BodyData: loadTestReq.HttpReqs.BodyData,
			},
		},
	}

	if err := tx.Create(&loadExecutionConfig).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return loadExecutionConfig.Model.ID, nil
}

func UpdateLoadExecutionState(envId, loadTestKey string, state constant.ExecutionStatus) error {
	db := configuration.DB()
	tx := db.Begin()

	loadEnvId, err := strconv.ParseUint(envId, 10, 64)

	if err != nil {
		return err
	}

	err = tx.Model(&model.LoadExecutionState{}).
		Where("load_env_id = ? AND load_test_key = ?", envId, loadTestKey).
		FirstOrCreate(&model.LoadExecutionState{
			LoadEnvID:       uint(loadEnvId),
			LoadTestKey:     loadTestKey,
			ExecutionStatus: state,
		}).
		Update("execution_status", state).
		Error

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func GetLoadExecutionConfigById(configId string) (model.LoadExecutionConfig, error) {
	db := configuration.DB()
	var loadExecutionConfig model.LoadExecutionConfig
	if err := db.Preload("LoadExecutionHttps").
		First(&loadExecutionConfig, configId).
		Error; err != nil {
		return loadExecutionConfig, err
	}

	return loadExecutionConfig, nil
}

func GetLoadExecutionState(envId, loadTestKey string) (model.LoadExecutionState, error) {
	db := configuration.DB()
	var loadExecutionState model.LoadExecutionState
	if err := db.Where("load_env_id = ? AND load_test_key = ?", envId, loadTestKey).
		First(&loadExecutionState).
		Error; err != nil {
		return loadExecutionState, err
	}

	return loadExecutionState, nil
}

func GetAllLoadExecutionState() ([]model.LoadExecutionState, error) {
	db := configuration.DB()

	var loadExecutionStates []model.LoadExecutionState

	result := db.Find(&loadExecutionStates)

	if err := result.Error; err != nil {
		return nil, err
	}
	return loadExecutionStates, nil
}
