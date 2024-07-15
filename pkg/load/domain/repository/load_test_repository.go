package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/database"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)

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
