package services

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
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

func ExecuteLoadTest(loadTestReq *api.LoadExecutionConfigReq) (uint, string, uint, error) {
	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
	loadTestReq.LoadTestKey = loadTestKey

	// check env
	if loadTestReq.EnvId != "" {
		loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
		if err != nil {
			return 0, "", 0, err
		}

		if loadEnv != nil && loadTestReq.LoadEnvReq.InstallLocation == "" {
			var env api.LoadEnvReq
			env.InstallLocation = (*loadEnv).InstallLocation
			env.RemoteConnectionType = (*loadEnv).RemoteConnectionType
			env.Username = (*loadEnv).Username
			env.PublicIp = (*loadEnv).PublicIp
			env.Cert = (*loadEnv).Cert
			env.NsId = (*loadEnv).NsId
			env.McisId = (*loadEnv).McisId

			loadTestReq.LoadEnvReq = env
		}
	}

	// installation jmeter
	envId, err := InstallLoadGenerator(&loadTestReq.LoadEnvReq)
	if err != nil {
		return 0, "", 0, err
	}

	loadTestReq.EnvId = fmt.Sprintf("%d", envId)

	log.Printf("[%s] start load test", loadTestKey)
	loadTestManager := managers.NewLoadTestManager()

	go func() {
		err = loadTestManager.Run(loadTestReq)
		if err != nil {
			err = repository.UpdateLoadExecutionState(loadTestReq.EnvId, loadTestKey, constant.Failed)
			if err != nil {
				log.Println(err)
			}
		}
		err = repository.UpdateLoadExecutionState(loadTestReq.EnvId, loadTestKey, constant.Success)
		if err != nil {
			log.Println(err)
		}
	}()

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return 0, "", 0, fmt.Errorf("service - execute load test error; %w", err)
	}

	loadExecutionConfigId, err := repository.SaveLoadTestExecution(loadTestReq)
	if err != nil {
		return 0, "", 0, err
	}

	return envId, loadTestKey, loadExecutionConfigId, nil
}

func StopLoadTest(loadTestReq api.LoadExecutionConfigReq) error {
	var env api.LoadEnvReq
	if loadTestReq.EnvId != "" {
		loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
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

		loadTestReq.LoadEnvReq = env
	}

	log.Printf("[%s] stop load test", loadTestReq.LoadTestKey)
	loadTestManager := managers.NewLoadTestManager()

	err := loadTestManager.Stop(loadTestReq)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return fmt.Errorf("service - execute load test error; %w", err)
	}

	return nil
}

func GetLoadTestResult(envId, testKey, format string) (interface{}, error) {
	loadEnv, err := repository.GetEnvironment(envId)
	if err != nil {
		return nil, err
	}

	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetResult(loadEnv, testKey, format)
	if err != nil {
		return nil, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}
	return result, nil
}

func GetLoadExecutionConfigById(configId string) (interface{}, error) {
	loadExecutionConfig, err := repository.GetLoadExecutionConfigById(configId)
	if err != nil {
		return nil, err
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionConfig.LoadEnvID)
	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return nil, err
	}

	var load api.LoadEnvRes
	load.LoadEnvId = loadEnv.ID
	load.InstallLocation = loadEnv.InstallLocation
	load.RemoteConnectionType = loadEnv.RemoteConnectionType
	load.Username = loadEnv.Username
	load.PublicIp = loadEnv.PublicIp
	load.Cert = loadEnv.Cert
	load.NsId = loadEnv.NsId
	load.McisId = loadEnv.McisId

	loadExecutionHttps := make([]api.LoadExecutionHttpRes, 0)

	for _, v := range loadExecutionConfig.LoadExecutionHttps {
		loadHttp := api.LoadExecutionHttpRes{
			LoadExecutionHttpId: v.LoadExecutionConfigID,
			Method:              v.Method,
			Protocol:            v.Protocol,
			Hostname:            v.Hostname,
			Port:                v.Port,
			Path:                v.Path,
			BodyData:            v.BodyData,
		}
		loadExecutionHttps = append(loadExecutionHttps, loadHttp)
	}

	res := api.LoadExecutionRes{
		LoadExecutionConfigId: loadExecutionConfig.ID,
		LoadTestKey:           loadExecutionConfig.LoadTestKey,
		Threads:               loadExecutionConfig.Threads,
		RampTime:              loadExecutionConfig.RampTime,
		LoopCount:             loadExecutionConfig.LoopCount,
		LoadEnv:               load,
		LoadExecutionHttp:     loadExecutionHttps,
	}

	return res, nil
}
