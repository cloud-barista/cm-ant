package repository

import (
	"errors"

	"github.com/cloud-barista/cm-ant/pkg/database"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"gorm.io/gorm"
)

func GetEnvironment(envId string) (*model.LoadEnv, error) {
	db := database.DB()
	var loadEnv model.LoadEnv

	result := db.First(&loadEnv, envId)

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &loadEnv, nil
		}
		return nil, err
	}

	return &loadEnv, nil
}
