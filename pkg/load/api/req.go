package api

import (
	"errors"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
)

type RemoteConnectionReq struct {
	RemoteConnectionType constant.RemoteConnectionType `json:"remoteConnectionType"`
	PublicId             string
	Username             string
	Cert                 string
	Port                 string
	NsId                 string
	McisId               string
}

func (r RemoteConnectionReq) Validate() error {
	if r.RemoteConnectionType == "" {
		return errors.New("remote connection type is required")
	}

	switch r.RemoteConnectionType {
	case constant.PrivateKey, constant.Password:
		if r.PublicId == "" || r.Username == "" || r.Cert == "" || r.Port == "" {
			return errors.New("pass all the arguments for requirement")
		}
	case constant.BuiltIn:
		if r.NsId == "" || r.McisId == "" || r.Username == "" {
			return errors.New("pass all the arguments for requirement")
		}
	default:
		return errors.New("invalid argument for remote connection type")

	}

	return nil
}

type LoadEnvReq struct {
	Type     constant.AccessType `json:"type"`
	NsId     string              `json:"nsId"`
	McisId   string              `json:"mcisId"`
	Username string              `json:"username"`
}

func (l LoadEnvReq) Validate() error {
	if l.Type == constant.Remote {
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
