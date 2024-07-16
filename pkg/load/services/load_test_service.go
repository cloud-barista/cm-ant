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
