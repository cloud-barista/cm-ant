package api

import (
	"errors"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
)

type LoadEnvReq struct {
	Type     constant.AccessType `json:"type"`
	NsId     string              `json:"nsId"`
	McisId   string              `json:"mcisId"`
	Username string              `json:"username"`
}

func (l LoadEnvReq) Validate() error {
	if l.Type == constant.REMOTE {
		if l.NsId == "" || l.McisId == "" || l.Username == "" {
			return errors.New("invalid user request for load env request")
		}
	}

	return nil
}

type LoadHttpReq struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path"`
	BodyData string `json:"bodyData"`
}

type LoadTestPropertyReq struct {
	PropertiesId string `json:"propertiesId"`

	Threads   string `json:"threads"`
	RampTime  string `json:"rampTime"`
	LoopCount string `json:"loopCount"`

	HttpReqs LoadHttpReq `json:"httpReqs"`

	LoadEnvReq LoadEnvReq `json:"loadEnvReq"`
}
