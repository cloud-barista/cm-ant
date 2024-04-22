package services

import (
	"fmt"
	"log"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/load/managers"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

func InstallLoadGenerator(installReq *api.LoadEnvReq) (uint, error) {
	loadTestManager := managers.NewLoadTestManager()

	if err := loadTestManager.Install(installReq); err != nil {
		return 0, fmt.Errorf("failed to install load generator: %w", err)
	}

	createdEnvId, err := repository.SaveLoadTestInstallEnv(installReq)
	if err != nil {
		return 0, fmt.Errorf("failed to save load test installation environment: %w", err)
	}
	log.Printf("Environment ID %d for load test is successfully created", createdEnvId)

	return createdEnvId, nil
}

func prepareEnvironment(loadTestReq *api.LoadExecutionConfigReq) error {
	if loadTestReq.EnvId == "" {
		return nil
	}

	loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	if loadEnv != nil && loadTestReq.LoadEnvReq.InstallLocation == "" {
		loadTestReq.LoadEnvReq = convertToLoadEnvReq(loadEnv)
	}

	return nil
}

func convertToLoadEnvReq(loadEnv *model.LoadEnv) api.LoadEnvReq {
	return api.LoadEnvReq{
		InstallLocation:      loadEnv.InstallLocation,
		RemoteConnectionType: loadEnv.RemoteConnectionType,
		Username:             loadEnv.Username,
		PublicIp:             loadEnv.PublicIp,
		Cert:                 loadEnv.Cert,
		NsId:                 loadEnv.NsId,
		McisId:               loadEnv.McisId,
	}
}

func runLoadTest(loadTestManager managers.LoadTestManager, loadTestReq *api.LoadExecutionConfigReq, loadTestKey string) {
	log.Printf("[%s] start load test", loadTestKey)
	if err := loadTestManager.Run(loadTestReq); err != nil {
		log.Printf("Error during load test: %v", err)
		if updateErr := repository.UpdateLoadExecutionState(loadTestReq.EnvId, loadTestKey, constant.Failed); updateErr != nil {
			log.Println(updateErr)
		}
	} else {
		if updateErr := repository.UpdateLoadExecutionState(loadTestReq.EnvId, loadTestKey, constant.Success); updateErr != nil {
			log.Println(updateErr)
		}
	}
}

func ExecuteLoadTest(loadTestReq *api.LoadExecutionConfigReq) (uint, string, uint, error) {
	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
	loadTestReq.LoadTestKey = loadTestKey

	// check env
	if err := prepareEnvironment(loadTestReq); err != nil {
		return 0, "", 0, err
	}

	// installation jmeter
	envId, err := InstallLoadGenerator(&loadTestReq.LoadEnvReq)
	if err != nil {
		return 0, "", 0, err
	}

	loadTestReq.EnvId = fmt.Sprintf("%d", envId)

	log.Printf("[%s] start load test", loadTestKey)
	loadTestManager := managers.NewLoadTestManager()

	go runLoadTest(loadTestManager, loadTestReq, loadTestKey)

	loadExecutionConfigId, err := repository.SaveLoadTestExecution(loadTestReq)
	if err != nil {
		return 0, "", 0, err
	}

	return envId, loadTestKey, loadExecutionConfigId, nil
}

func StopLoadTest(loadTestReq api.LoadExecutionConfigReq) error {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestReq.EnvId, loadTestReq.LoadTestKey)

	if err != nil {
		return err
	}

	if loadExecutionState.IsFinished() {
		return fmt.Errorf("load test is already finished")
	}

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

	err = loadTestManager.Stop(loadTestReq)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return fmt.Errorf("service - execute load test error; %w", err)
	}

	return nil
}

func GetLoadTestResult(envId, testKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(envId, testKey)

	if err != nil {
		return nil, err
	}

	if !loadExecutionState.IsFinished() {
		return nil, fmt.Errorf("load test is under executing")
	}

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
		Threads:               loadExecutionConfig.VirtualUsers,
		RampTime:              loadExecutionConfig.Duration,
		LoopCount:             loadExecutionConfig.RampUpTime,
		LoadEnv:               load,
		LoadExecutionHttp:     loadExecutionHttps,
	}

	return res, nil
}

func GetAllLoadExecutionState() (interface{}, error) {
	loadExecutionStates, err := repository.GetAllLoadExecutionState()

	if err != nil {
		return nil, err
	}

	responseStates := make([]api.LoadExecutionStateRes, 0)
	for _, v := range loadExecutionStates {
		loadExecutionStateRes := api.LoadExecutionStateRes{
			LoadExecutionStateId: v.ID,
			LoadEnvID:            v.LoadEnvID,
			LoadTestKey:          v.LoadTestKey,
			ExecutionStatus:      v.ExecutionStatus,
			ExecutionDate:        v.CreatedAt,
		}

		responseStates = append(responseStates, loadExecutionStateRes)
	}

	return responseStates, nil
}
