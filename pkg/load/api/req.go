package api

import (
	"errors"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
)


type LoadEnvReq struct {
	InstallLocation      constant.InstallLocation      `json:"installLocation,omitempty"`
	RemoteConnectionType constant.RemoteConnectionType `json:"remoteConnectionType,omitempty"`
	Username             string                        `json:"username,omitempty"`

	PublicIp string `json:"publicIp,omitempty"`
	Cert     string `json:"cert,omitempty"`

	NsId   string `json:"nsId,omitempty"`
	McisId string `json:"mcisId,omitempty"`
}

func (l LoadEnvReq) Validate() error {

	if l.InstallLocation == constant.Remote {
		if l.RemoteConnectionType == "" {
			return errors.New("remote connection type should set")
		}

		switch l.RemoteConnectionType {
		case constant.BuiltIn:
			if l.NsId == "" ||
				l.McisId == "" ||
				l.Username == "" {
				return errors.New("check build in properties. all field has to filled")
			}
		case constant.PrivateKey, constant.Password:
			if l.PublicIp == "" ||
				l.Cert == "" ||
				l.Username == "" {
				return errors.New("check Secure shell properties. all field has to filled")
			}
		}
	}

	return nil
}

type LoadExecutionHttpReq struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path"`
	BodyData string `json:"bodyData"`
}

type LoadExecutionConfigReq struct {
	LoadTestKey string `json:"loadTestKey"`

	EnvId string `json:"envId"`

	Threads   string `json:"threads"`
	RampTime  string `json:"rampTime"`
	LoopCount string `json:"loopCount"`

	HttpReqs LoadExecutionHttpReq `json:"httpReqs"`

	LoadEnvReq LoadEnvReq `json:"loadEnvReq"`
}
