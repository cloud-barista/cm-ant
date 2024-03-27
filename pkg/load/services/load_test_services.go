package services

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/load/managers"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"log"
)

func InstallLoadGenerator(installReq api.LoadEnvReq) error {
	loadTestManager := managers.NewLoadTestManager()

	err := loadTestManager.Install(installReq)
	if err != nil {
		return fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}

	created, err := repository.SaveLoadTestInstallEnv(installReq)
	if err != nil {
		return fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}
	log.Println(created, " load test env object is created")
	return nil
}

func ExecuteLoadTest(properties api.LoadTestPropertyReq) (string, error) {
	propertiesId := utils.CreateUniqIdBaseOnUnixTime()
	properties.PropertiesId = propertiesId

	log.Printf("[%s] start load test", propertiesId)
	loadTestManager := managers.NewLoadTestManager()

	testId, err := loadTestManager.Run(properties)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return "", fmt.Errorf("service - execute load test error; %w", err)
	}

	return testId, nil
}

func StopLoadTest(properties api.LoadTestPropertyReq) error {

	log.Printf("[%s] stop load test", properties.PropertiesId)
	loadTestManager := managers.NewLoadTestManager()

	err := loadTestManager.Stop(properties)

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
