package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)

func SaveLoadTestInstallEnv(installReq api.LoadEnvReq) (int64, error) {
	db := configuration.DB()
	tx := db.Begin()
	loadEnv := model.LoadEnv{
		Type:     installReq.Type,
		NsId:     &installReq.NsId,
		McisId:   &installReq.McisId,
		Username: &installReq.Username,
	}

	if err := tx.FirstOrCreate(
		&loadEnv,
		"type = ? AND ns_id = ? AND mcis_id = ? AND username = ?",
		installReq.Type, installReq.NsId, installReq.McisId, installReq.Username,
	).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return tx.RowsAffected, nil
}
