package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)

func GetAllEnvironment() ([]model.LoadEnv, error) {
	db := configuration.DB()
	var loadEnvs []model.LoadEnv

	result := db.Find(&loadEnvs)

	if err := result.Error; err != nil {
		return nil, err
	}

	return loadEnvs, nil
}

func SaveLoadTestInstallEnv(installReq api.LoadEnvReq) (uint, error) {
	db := configuration.DB()
	tx := db.Begin()

	loadEnv := model.LoadEnv{
		InstallLocation:      installReq.InstallLocation,
		RemoteConnectionType: installReq.RemoteConnectionType,
		NsId:                 installReq.NsId,
		McisId:               installReq.McisId,
		Username:             installReq.Username,
		PublicIp:             installReq.PublicIp,
		Cert:                 installReq.Cert,
	}

	if err := tx.FirstOrCreate(
		&loadEnv,
		"install_location = ? AND remote_connection_type = ? AND ns_id = ? AND mcis_id = ? AND username = ? AND public_ip = ? AND cert = ?",
		loadEnv.InstallLocation, loadEnv.RemoteConnectionType, loadEnv.NsId, loadEnv.McisId, loadEnv.Username, loadEnv.PublicIp, loadEnv.Cert,
	).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return loadEnv.Model.ID, nil
}
