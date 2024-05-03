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

func GetAgentInfo(agentId string) (*model.AgentInfo, error) {
	db := configuration.DB()
	var agentInfo model.AgentInfo
	result := db.First(&agentInfo, agentId)

	if err := result.Error; err != nil {
		return nil, err
	}

	return &agentInfo, nil
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
