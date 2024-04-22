package managers

import (
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/managers/jmetermanager"
)

type LoadTestManager interface {
	Install(*api.LoadEnvReq) error
	Stop(api.LoadExecutionConfigReq) error
	Run(*api.LoadExecutionConfigReq) error
	GetResult(loadEnv *model.LoadEnv, loadTestKey string, format string) (interface{}, error)
}

func NewLoadTestManager() LoadTestManager {
	return &jmetermanager.JMeterLoadTestManager{}
}
