package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)

func SaveAgentInfo(a *api.AgentReq) (uint, error) {
	db := configuration.DB()
	tx := db.Begin()

	agentInfo := model.AgentInfo{
		Username:   a.Username,
		PublicIp:   a.PublicIp,
		PemKeyPath: a.PemKeyPath,
	}

	if err := tx.
		Where(
			"username = ? AND public_ip = ? AND pem_key_path = ?",
			agentInfo.Username, agentInfo.PublicIp, agentInfo.PemKeyPath,
		).
		FirstOrCreate(
			&agentInfo,
		).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return agentInfo.Model.ID, nil
}

func InsertAgentInstallInfo(agentInstallInfo *model.AgentInstallInfo) error {
	db := configuration.DB()
	tx := db.Begin()

	if err := tx.
		Where(
			"ns_id = ? AND mcis_id = ? AND vm_id = ? AND username = ?",
			agentInstallInfo.NsId, agentInstallInfo.McisId, agentInstallInfo.VmId, agentInstallInfo.Username,
		).
		FirstOrCreate(
			agentInstallInfo,
		).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func UpdateAgentInstallInfoStatus(agentInstallInfo *model.AgentInstallInfo) error {
	db := configuration.DB()
	tx := db.Begin()

	if err := tx.
		Model(agentInstallInfo).
		Update(
			"status", agentInstallInfo.Status,
		).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func GetAgentInfo(agentId string) (*model.AgentInfo, error) {
	db := configuration.DB()
	var agentInfo model.AgentInfo
	result := db.First(&agentInfo, agentId)

	if err := result.Error; err != nil {
		return nil, err
	}

	return &agentInfo, nil
}

func GetAllAgentInstallInfos() ([]model.AgentInstallInfo, error) {
	db := configuration.DB()

	var agentInstallInfos []model.AgentInstallInfo

	result := db.Find(&agentInstallInfos)

	if err := result.Error; err != nil {
		return nil, err
	}

	return agentInstallInfos, nil
}

func GetAgentInstallInfo(agentInstallInfoId string) (model.AgentInstallInfo, error) {
	db := configuration.DB()

	var agentInstallInfo model.AgentInstallInfo

	result := db.First(&agentInstallInfo, agentInstallInfoId)

	if err := result.Error; err != nil {
		return agentInstallInfo, err
	}

	return agentInstallInfo, nil
}

func DeleteAgentInstallInfo(agentInstallInfoId string) error {
	db := configuration.DB()
	tx := db.Begin()

	if err := tx.
		Delete(
			&model.AgentInstallInfo{}, agentInstallInfoId,
		).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func DeleteAgentInfo(agentId string) error {
	db := configuration.DB()
	tx := db.Begin()

	if err := tx.
		Delete(
			&model.AgentInfo{}, agentId,
		).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
