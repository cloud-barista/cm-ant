package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/database"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)


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
