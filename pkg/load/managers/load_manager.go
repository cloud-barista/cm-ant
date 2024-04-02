package managers

import (
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/managers/jmetermanager"
)

type LoadTestManager interface {
	Install(*api.LoadEnvReq) error
	Stop(api.LoadTestReq) error
	Run(*api.LoadTestReq) error
	GetResult(*model.LoadEnv, string, string) (interface{}, error)
}

func NewLoadTestManager() LoadTestManager {
	return &jmetermanager.JMeterLoadTestManager{}
}
