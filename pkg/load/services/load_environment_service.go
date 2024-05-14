package services

import (
	"fmt"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
)

func GetAllLoadEnvironments() ([]api.LoadEnvRes, error) {
	result, err := repository.GetAllLoadEnvironments()

	if err != nil {
		return nil, err
	}

	var responseEnv []api.LoadEnvRes

	for _, loadEnv := range result {
		var load api.LoadEnvRes
		load.LoadEnvId = loadEnv.ID
		load.InstallLocation = loadEnv.InstallLocation
		load.Username = loadEnv.Username
		load.PublicIp = loadEnv.PublicIp
		load.PemKeyPath = loadEnv.PemKeyPath
		load.NsId = loadEnv.NsId
		load.McisId = loadEnv.McisId
		load.VmId = loadEnv.VmId

		responseEnv = append(responseEnv, load)
	}

	return responseEnv, nil
}

func DeleteLoadEnvironment(envId string) error {
	fmt.Println("asdf")
	return nil
}
