package api

import (
	"errors"

	"github.com/cloud-barista/cm-ant/pkg/load/constant"
)

type LoadExecutionConfigReq struct {
	LoadTestKey string `json:"loadTestKey,omitempty"`
	EnvId       string `json:"envId,omitempty"`
	TestName    string `json:"testName"`

	VirtualUsers string `json:"virtualUsers"`
	Duration     string `json:"duration"`
	RampUpTime   string `json:"rampUpTime"`
	RampUpSteps  string `json:"rampUpSteps"`

	HttpReqs []LoadExecutionHttpReq `json:"httpReqs,omitempty"`

	LoadEnvReq LoadEnvReq `json:"loadEnvReq,omitempty"`
}

type LoadExecutionHttpReq struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`
}

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
