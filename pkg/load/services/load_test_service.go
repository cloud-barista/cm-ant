package services

import (
	"errors"
	"fmt"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"

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
