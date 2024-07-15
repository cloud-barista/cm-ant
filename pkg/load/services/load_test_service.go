package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/load/managers"
)

func GetLoadTestResult(testKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(testKey)
	if err != nil {
		return nil, err
	}

	if loadExecutionState.ExecutionStatus == constant.Processing {
		return nil, errors.New("load test is processing")
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)

	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return nil, err
	}

	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetResult(loadEnv, testKey, format)
	if err != nil {
		return nil, fmt.Errorf("error on [GetLoadTestResult()]; %s", err)
	}
	return result, nil
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
		env.Username = (*loadEnv).Username
		env.PublicIp = (*loadEnv).PublicIp
		env.PemKeyPath = (*loadEnv).PemKeyPath
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

func GetLoadTestMetrics(loadTestKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKey)
	if err != nil {
		return nil, err
	}

	if loadExecutionState.ExecutionStatus == constant.Processing {
		return nil, errors.New("load test is processing")
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)

	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return nil, err
	}

	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetMetrics(loadEnv, loadTestKey, format)
	if err != nil {
		return nil, fmt.Errorf("error on [GetLoadTestMetrics()]; %s", err)
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
		load.Username = loadEnv.Username
		load.PublicIp = loadEnv.PublicIp
		load.PemKeyPath = loadEnv.PemKeyPath
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
				TotalSec:             state.TotalSec,
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
	load.Username = loadEnv.Username
	load.PublicIp = loadEnv.PublicIp
	load.PemKeyPath = loadEnv.PemKeyPath
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
