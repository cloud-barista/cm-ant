package services

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
)

func GetAllRemoteConnection() ([]model.LoadEnv, error) {
	allEnvs, err := repository.GetAllEnvironment()

	if err != nil {
		return nil, err
	}

	return allEnvs, nil
}

func DeleteRemoteConnection(envId string) error {
	fmt.Println("asdf")
	return nil
}
