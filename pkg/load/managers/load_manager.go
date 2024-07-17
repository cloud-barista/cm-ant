package managers

import (

	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/managers/jmetermanager"
)

type LoadTestManager interface {
	GetResult(loadEnv *model.LoadEnv, loadTestKey string, format string) (interface{}, error)
	GetMetrics(loadEnv *model.LoadEnv, loadTestKey string, format string) (interface{}, error)
}

func NewLoadTestManager() LoadTestManager {
	return &jmetermanager.JMeterLoadTestManager{}
}
