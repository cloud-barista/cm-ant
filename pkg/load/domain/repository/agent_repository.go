package repository

import (
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
)

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
