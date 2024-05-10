package repository

import (
	"errors"
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"gorm.io/gorm"
)

func GetAllLoadEnvironments() ([]model.LoadEnv, error) {
	db := configuration.DB()

	var loadEnvs []model.LoadEnv

	result := db.Find(&loadEnvs)

	if err := result.Error; err != nil {
		return nil, err
	}
	return loadEnvs, nil
}

func GetEnvironment(envId string) (*model.LoadEnv, error) {
	db := configuration.DB()
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

func SaveLoadTestInstallEnv(loadEnv *model.LoadEnv) (uint, error) {
	db := configuration.DB()
	tx := db.Begin()

	if loadEnv == nil {
		return 0, errors.New("load test environment is empty")
	}

	if err := tx.
		Where(
			"install_location = ? AND remote_connection_type = ? AND ns_id = ? AND mcis_id = ? AND vm_id = ? AND username = ? AND public_ip = ? AND cert = ?",
			loadEnv.InstallLocation, loadEnv.RemoteConnectionType, loadEnv.NsId, loadEnv.McisId, loadEnv.VmId, loadEnv.Username, loadEnv.PublicIp, loadEnv.Cert,
		).
		FirstOrCreate(
			loadEnv,
		).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return loadEnv.Model.ID, nil
}

func DeleteLoadTestInstallEnv(loadEnvId string) error {
	db := configuration.DB()
	tx := db.Begin()

	if err := tx.
		Delete(
			&model.LoadEnv{}, loadEnvId,
		).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
