package api

import "github.com/cloud-barista/cm-ant/pkg/load/constant"

type LoadEnvRes struct {
	LoadEnvId            uint                          `json:"loadEnvId"`
	InstallLocation      constant.InstallLocation      `json:"installLocation,omitempty"`
	RemoteConnectionType constant.RemoteConnectionType `json:"remoteConnectionType,omitempty"`
	Username             string                        `json:"username,omitempty"`

	PublicIp string `json:"publicIp,omitempty"`
	Cert     string `json:"cert,omitempty"`

	NsId   string `json:"nsId,omitempty"`
	McisId string `json:"mcisId,omitempty"`
}

type LoadExecutionRes struct {
	LoadExecutionConfigId uint   `json:"loadExecutionConfigId"`
	LoadTestKey           string `json:"loadTestKey"`
	Threads               string `json:"threads"`
	RampTime              string `json:"rampTime"`
	LoopCount             string `json:"loopCount"`

	LoadEnv           LoadEnvRes             `json:"loadEnv"`
	LoadExecutionHttp []LoadExecutionHttpRes `json:"loadExecutionHttp"`
}

type LoadExecutionHttpRes struct {
	LoadExecutionHttpId uint   `json:"loadExecutionHttpId"`
	Method              string `json:"method"`
	Protocol            string `json:"protocol"`
	Hostname            string `json:"hostname"`
	Port                string `json:"port"`
	Path                string `json:"path"`
	BodyData            string `json:"bodyData"`
}
