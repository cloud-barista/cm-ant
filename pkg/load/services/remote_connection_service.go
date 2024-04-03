package services

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
)

func GetAllRemoteConnection() ([]api.LoadEnvRes, error) {
	result, err := repository.GetAllEnvironment()

	if err != nil {
		return nil, err
	}

	var responseEnv []api.LoadEnvRes

	for _, loadEnv := range result {
		var load api.LoadEnvRes
		load.LoadEnvId = loadEnv.ID
		load.InstallLocation = loadEnv.InstallLocation
		load.RemoteConnectionType = loadEnv.RemoteConnectionType
		load.Username = loadEnv.Username
		load.PublicIp = loadEnv.PublicIp
		load.Cert = loadEnv.Cert
		load.NsId = loadEnv.NsId
		load.McisId = loadEnv.McisId

		responseEnv = append(responseEnv, load)
	}

	return responseEnv, nil
}

func DeleteRemoteConnection(envId string) error {
	fmt.Println("asdf")
	return nil
}
