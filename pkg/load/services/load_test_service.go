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

func InstallLoadTester(installReq *api.LoadEnvReq) (uint, error) {
	loadTestManager := managers.NewLoadTestManager()

	if err := loadTestManager.Install(installReq); err != nil {
		return 0, fmt.Errorf("failed to install load tester: %w", err)
	}

	loadEnv := model.LoadEnv{
		InstallLocation:      (*installReq).InstallLocation,
		RemoteConnectionType: (*installReq).RemoteConnectionType,
		NsId:                 (*installReq).NsId,
		McisId:               (*installReq).McisId,
		VmId:                 (*installReq).VmId,
		Username:             (*installReq).Username,
		PublicIp:             (*installReq).PublicIp,
		Cert:                 (*installReq).Cert,
	}

	createdEnvId, err := repository.SaveLoadTestInstallEnv(&loadEnv)
	if err != nil {
		return 0, fmt.Errorf("failed to save load test installation environment: %w", err)
	}
	log.Printf("Environment ID %d for load test is successfully created", createdEnvId)

	return createdEnvId, nil
}

func UninstallLoadTester(loadEnvId string) error {
	loadTestManager := managers.NewLoadTestManager()

	var loadEnvReq api.LoadEnvReq
	if loadEnvId != "" {
		loadEnv, err := repository.GetEnvironment(loadEnvId)
		if err != nil {
			return err
		}

		loadEnvReq.InstallLocation = (*loadEnv).InstallLocation
		loadEnvReq.RemoteConnectionType = (*loadEnv).RemoteConnectionType
		loadEnvReq.Username = (*loadEnv).Username
		loadEnvReq.PublicIp = (*loadEnv).PublicIp
		loadEnvReq.Cert = (*loadEnv).Cert
		loadEnvReq.NsId = (*loadEnv).NsId
		loadEnvReq.McisId = (*loadEnv).McisId
		loadEnvReq.VmId = (*loadEnv).VmId
	}

	if err := loadTestManager.Uninstall(&loadEnvReq); err != nil {
		return fmt.Errorf("failed to uninstall load tester: %w", err)
	}

	//err := repository.DeleteLoadTestInstallEnv(loadEnvId)
	//if err != nil {
	//	return fmt.Errorf("failed to delete load test installation environment: %w", err)
	//}
	log.Println("load test environment is successfully deleted")

	return nil
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
		VmId:                 loadEnv.VmId,
	}
}

func runLoadTest(loadTestManager managers.LoadTestManager, loadTestReq *api.LoadExecutionConfigReq, loadTestKey string) {
	log.Printf("[%s] start load test", loadTestKey)
	if err := loadTestManager.Run(loadTestReq); err != nil {
		log.Printf("Error during load test: %v", err)
		if updateErr := repository.UpdateLoadExecutionState(loadTestKey, constant.Failed); updateErr != nil {
			log.Println(updateErr)
		}
	} else {
		log.Printf("load test complete!")

		if updateErr := repository.UpdateLoadExecutionState(loadTestKey, constant.Success); updateErr != nil {
			log.Println(updateErr)
		}
	}
}

func ExecuteLoadTest(loadTestReq *api.LoadExecutionConfigReq) (string, error) {
	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
	loadTestReq.LoadTestKey = loadTestKey

	// check env
	if err := prepareEnvironment(loadTestReq); err != nil {
		return "", err
	}

	// installation jmeter
	envId, err := InstallLoadTester(&loadTestReq.LoadEnvReq)
	if err != nil {
		return "", err
	}

	loadTestReq.EnvId = fmt.Sprintf("%d", envId)

	log.Printf("[%s] start load test", loadTestKey)
	loadTestManager := managers.NewLoadTestManager()

	go runLoadTest(loadTestManager, loadTestReq, loadTestKey)

	_, err = repository.SaveLoadTestExecution(loadTestReq)
	if err != nil {
		return "", err
	}

	return loadTestKey, nil
}

func StopLoadTest(loadTestKeyReq api.LoadTestKeyReq) error {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKeyReq.LoadTestKey)

	if err != nil {
		return err
	}

	if loadExecutionState.IsFinished() {
		return fmt.Errorf("load test is already finished")
	}

	loadTestReq := api.LoadExecutionConfigReq{
		LoadTestKey: loadTestKeyReq.LoadTestKey,
		EnvId:       fmt.Sprintf("%d", loadExecutionState.LoadEnvID),
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
		env.VmId = (*loadEnv).VmId

		loadTestReq.LoadEnvReq = env
	}

	log.Printf("[%s] stop load test. %+v\n", loadTestKeyReq.LoadTestKey, loadTestReq)
	loadTestManager := managers.NewLoadTestManager()

	err = loadTestManager.Stop(loadTestReq)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return fmt.Errorf("service - execute load test error; %w", err)
	}

	return nil
}

func GetLoadTestResult(testKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(testKey)
	if err != nil {
		return nil, err
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)

	loadEnv, err := repository.GetEnvironment(loadEnvId)
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

func GetLoadTestMetrics(testKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(testKey)
	if err != nil {
		return nil, err
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)

	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return nil, err
	}

	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetMetrics(loadEnv, testKey, format)
	if err != nil {
		return nil, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
	}
	return result, nil
}

func GetAllLoadExecutionConfig() ([]api.LoadExecutionRes, error) {
	loadExecutionConfigs, err := repository.GetAllLoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	var loadExecutionConfigResponses []api.LoadExecutionRes
	for _, v := range loadExecutionConfigs {
		loadEnvId := fmt.Sprintf("%d", v.LoadEnvID)
		loadEnv, err := repository.GetEnvironment(loadEnvId)
		if err != nil {
			return nil, err
		}
		state, err := repository.GetLoadExecutionState(v.LoadTestKey)
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
		load.VmId = loadEnv.VmId

		loadExecutionHttps := make([]api.LoadExecutionHttpRes, 0)

		for _, v := range v.LoadExecutionHttps {
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
			LoadExecutionConfigId: v.ID,
			LoadTestKey:           v.LoadTestKey,
			VirtualUsers:          v.VirtualUsers,
			Duration:              v.Duration,
			RampUpTime:            v.RampUpTime,
			RampUpSteps:           v.RampUpSteps,
			LoadEnv:               load,
			LoadExecutionHttp:     loadExecutionHttps,
			TestName:              v.TestName,
			LoadExecutionState: api.LoadExecutionStateRes{
				LoadExecutionStateId: state.ID,
				LoadTestKey:          state.LoadTestKey,
				ExecutionStatus:      state.ExecutionStatus,
				StartAt:              state.StartAt,
				EndAt:                state.EndAt,
			},
		}

		loadExecutionConfigResponses = append(loadExecutionConfigResponses, res)
	}

	return loadExecutionConfigResponses, nil
}

func GetLoadExecutionConfig(loadTestKey string) (api.LoadExecutionRes, error) {
	loadExecutionConfig, err := repository.GetLoadExecutionConfig(loadTestKey)
	if err != nil {
		return api.LoadExecutionRes{}, err
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionConfig.LoadEnvID)
	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return api.LoadExecutionRes{}, err
	}
	state, err := repository.GetLoadExecutionState(loadTestKey)
	if err != nil {
		return api.LoadExecutionRes{}, err
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
	load.VmId = loadEnv.VmId

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
		VirtualUsers:          loadExecutionConfig.VirtualUsers,
		Duration:              loadExecutionConfig.Duration,
		RampUpTime:            loadExecutionConfig.RampUpTime,
		RampUpSteps:           loadExecutionConfig.RampUpSteps,
		TestName:              loadExecutionConfig.TestName,
		LoadEnv:               load,
		LoadExecutionHttp:     loadExecutionHttps,
		LoadExecutionState: api.LoadExecutionStateRes{
			LoadExecutionStateId: state.ID,
			LoadTestKey:          state.LoadTestKey,
			ExecutionStatus:      state.ExecutionStatus,
			StartAt:              state.StartAt,
			EndAt:                state.EndAt,
		},
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
			LoadTestKey:          v.LoadTestKey,
			ExecutionStatus:      v.ExecutionStatus,
			StartAt:              v.StartAt,
			EndAt:                v.EndAt,
		}

		responseStates = append(responseStates, loadExecutionStateRes)
	}

	return responseStates, nil
}

func GetLoadExecutionState(loadTestKey string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKey)

	if err != nil {
		return nil, err
	}

	loadExecutionStateRes := api.LoadExecutionStateRes{
		LoadExecutionStateId: loadExecutionState.ID,
		LoadTestKey:          loadExecutionState.LoadTestKey,
		ExecutionStatus:      loadExecutionState.ExecutionStatus,
		StartAt:              loadExecutionState.StartAt,
		EndAt:                loadExecutionState.EndAt,
	}

	return loadExecutionStateRes, nil
}
