package api

import (
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
)

type LoadTestKeyReq struct {
	LoadTestKey string `json:"loadTestKey,omitempty"`
}

type LoadEnvIdReq struct {
	LoadEnvId string `json:"loadEnvId,omitempty"`
}

type LoadExecutionConfigReq struct {
	LoadTestKey string `json:"loadTestKey,omitempty"`

	EnvId      string     `json:"envId,omitempty"`
	LoadEnvReq LoadEnvReq `json:"loadEnvReq,omitempty"`

	TestName     string `json:"testName"`
	VirtualUsers string `json:"virtualUsers"`
	Duration     string `json:"duration"`
	RampUpTime   string `json:"rampUpTime"`
	RampUpSteps  string `json:"rampUpSteps"`

	HttpReqs []LoadExecutionHttpReq `json:"httpReqs,omitempty"`
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
	InstallLocation constant.InstallLocation `json:"installLocation,omitempty"`
	Username        string                   `json:"username,omitempty"`

	PublicIp   string `json:"publicIp,omitempty"`
	PemKeyPath string `json:"pemKeyPath,omitempty"`

	NsId   string `json:"nsId,omitempty"`
	McisId string `json:"mcisId,omitempty"`
	VmId   string `json:"vmId,omitempty"`
}

type AntTargetServerReq struct {
	NsId   string `json:"nsId"`
	McisId string `json:"mcisId"`
	VmId   string `json:"vmId,omitempty"`
}
