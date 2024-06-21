package repository

import (
	"strconv"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/database"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)

func SaveLoadTestExecution(loadTestReq *api.LoadExecutionConfigReq) (uint, error) {
	db := database.DB()
	tx := db.Begin()
	loadEnvId, err := strconv.ParseUint(loadTestReq.EnvId, 10, 64)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	durationSec, err := strconv.Atoi(loadTestReq.Duration)
	if err != nil {
		return 0, err
	}

	rampUpSec, err := strconv.Atoi(loadTestReq.RampUpTime)
	if err != nil {
		return 0, err
	}

	totalSec := durationSec + rampUpSec

	loadExecutionState := model.LoadExecutionState{
		LoadEnvID:       uint(loadEnvId),
		LoadTestKey:     loadTestReq.LoadTestKey,
		ExecutionStatus: constant.Processing,
		StartAt:         time.Now(),
		TotalSec:        uint(totalSec),
	}

	if err := tx.
		Where("load_test_key = ?", loadTestReq.LoadTestKey).
		FirstOrCreate(&loadExecutionState).
		Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	loadExecutionHttps := []model.LoadExecutionHttp{}

	for _, loadExecutionHttp := range loadTestReq.HttpReqs {
		leh := model.LoadExecutionHttp{
			Method:   loadExecutionHttp.Method,
			Protocol: loadExecutionHttp.Protocol,
			Hostname: loadExecutionHttp.Hostname,
			Port:     loadExecutionHttp.Port,
			Path:     loadExecutionHttp.Path,
			BodyData: loadExecutionHttp.BodyData,
		}

		loadExecutionHttps = append(loadExecutionHttps, leh)
	}

	if len(loadExecutionHttps) < 1 {
		tx.Rollback()
		return 0, err
	}

	loadExecutionConfig := model.LoadExecutionConfig{
		LoadEnvID:          uint(loadEnvId),
		LoadTestKey:        loadTestReq.LoadTestKey,
		TestName:           loadTestReq.TestName,
		VirtualUsers:       loadTestReq.VirtualUsers,
		Duration:           loadTestReq.Duration,
		RampUpTime:         loadTestReq.RampUpTime,
		RampUpSteps:        loadTestReq.RampUpSteps,
		LoadExecutionHttps: loadExecutionHttps,
	}

	if err := tx.Create(&loadExecutionConfig).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return loadExecutionConfig.Model.ID, nil
}

func UpdateLoadExecutionStateWithNoTime(loadTestKey string, status constant.ExecutionStatus) error {
	db := database.DB()
	tx := db.Begin()

	err := tx.Model(&model.LoadExecutionState{}).
		Where("load_test_key = ?", loadTestKey).
		Updates(map[string]interface{}{"execution_status": status}).
		Error

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func UpdateLoadExecutionState(loadTestKey string, status constant.ExecutionStatus) error {
	db := database.DB()
	tx := db.Begin()

	now := time.Now()
	err := tx.Model(&model.LoadExecutionState{}).
		Where("load_test_key = ?", loadTestKey).
		Updates(map[string]interface{}{"execution_status": status, "end_at": &now}).
		Error

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func GetAllLoadExecutionConfig() ([]model.LoadExecutionConfig, error) {
	db := database.DB()
	var loadExecutionConfigs []model.LoadExecutionConfig
	if err := db.Preload("LoadExecutionHttps").
		Find(&loadExecutionConfigs).
		Error; err != nil {
		return loadExecutionConfigs, err
	}

	return loadExecutionConfigs, nil
}

func GetLoadExecutionConfig(testKey string) (model.LoadExecutionConfig, error) {
	db := database.DB()
	var loadExecutionConfig model.LoadExecutionConfig
	if err := db.Preload("LoadExecutionHttps").
		Where("load_test_key = ?", testKey).
		First(&loadExecutionConfig).
		Error; err != nil {
		return loadExecutionConfig, err
	}

	return loadExecutionConfig, nil
}

func GetLoadExecutionState(loadTestKey string) (model.LoadExecutionState, error) {
	db := database.DB()
	var loadExecutionState model.LoadExecutionState
	if err := db.Where("load_test_key = ?", loadTestKey).
		First(&loadExecutionState).
		Error; err != nil {
		return loadExecutionState, err
	}

	return loadExecutionState, nil
}

func GetAllLoadExecutionState() ([]model.LoadExecutionState, error) {
	db := database.DB()

	var loadExecutionStates []model.LoadExecutionState

	result := db.Find(&loadExecutionStates)

	if err := result.Error; err != nil {
		return nil, err
	}
	return loadExecutionStates, nil
}
