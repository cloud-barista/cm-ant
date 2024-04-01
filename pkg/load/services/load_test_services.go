package services

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/load/managers"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"log"
)

func InstallLoadGenerator(installReq *api.LoadEnvReq) (uint, error) {
	loadTestManager := managers.NewLoadTestManager()

	err := loadTestManager.Install(installReq)
	if err != nil {
		return 0, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}

	createdEnvId, err := repository.SaveLoadTestInstallEnv(installReq)
	if err != nil {
		return 0, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}
	log.Println(createdEnvId, " load test env object is created")
	return createdEnvId, nil
}

func ExecuteLoadTest(property *api.LoadTestPropertyReq) (uint, string, error) {
	propertiesId := utils.CreateUniqIdBaseOnUnixTime()
	property.PropertiesId = propertiesId

	// check env
	if property.EnvId != "" {
		loadEnv, err := repository.GetEnvironment(property.EnvId)
		if err != nil {
			return 0, "", err
		}

		if loadEnv != nil && property.LoadEnvReq.InstallLocation == "" {
			var env api.LoadEnvReq
			env.InstallLocation = (*loadEnv).InstallLocation
			env.RemoteConnectionType = (*loadEnv).RemoteConnectionType
			env.Username = (*loadEnv).Username
			env.PublicIp = (*loadEnv).PublicIp
			env.Cert = (*loadEnv).Cert
			env.NsId = (*loadEnv).NsId
			env.McisId = (*loadEnv).McisId

			property.LoadEnvReq = env
		}
	}

	// installation jmeter
	envId, err := InstallLoadGenerator(&property.LoadEnvReq)
	if err != nil {
		return 0, "", err
	}

	log.Printf("[%s] start load test", propertiesId)
	loadTestManager := managers.NewLoadTestManager()

	testId, err := loadTestManager.Run(property)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return 0, "", fmt.Errorf("service - execute load test error; %w", err)
	}

	return envId, testId, nil
}

func StopLoadTest(property api.LoadTestPropertyReq) error {
	var env api.LoadEnvReq
	if property.EnvId != "" {
		loadEnv, err := repository.GetEnvironment(property.EnvId)
		if err != nil {
			return err
		}

		env.InstallLocation = (*loadEnv).InstallLocation
		env.RemoteConnectionType = (*loadEnv).RemoteConnectionType
		env.Username = (*loadEnv).Username
		env.PublicIp = (*loadEnv).PublicIp
		env.Cert = (*loadEnv).Cert
		env.NsId = (*loadEnv).NsId
		env.McisId = (*loadEnv).McisId

		property.LoadEnvReq = env
	}

	log.Printf("[%s] stop load test", property.PropertiesId)
	loadTestManager := managers.NewLoadTestManager()

	err := loadTestManager.Stop(property)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return fmt.Errorf("service - execute load test error; %w", err)
	}

	return nil
}

func GetLoadTestResult(testId string) (interface{}, error) {
	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetResult(testId)
	if err != nil {
		return nil, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}
	return result, nil
}
