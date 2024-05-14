package api

import (
	"time"

	"github.com/cloud-barista/cm-ant/pkg/load/constant"
)

type LoadEnvRes struct {
	LoadEnvId       uint                     `json:"loadEnvId"`
	InstallLocation constant.InstallLocation `json:"installLocation,omitempty"`
	Username        string                   `json:"username,omitempty"`

	PublicIp   string `json:"publicIp,omitempty"`
	PemKeyPath string `json:"pemKeyPath,omitempty"`

	NsId   string `json:"nsId,omitempty"`
	McisId string `json:"mcisId,omitempty"`
	VmId   string `json:"vmId,omitempty"`
}

type LoadExecutionRes struct {
	LoadExecutionConfigId uint   `json:"loadExecutionConfigId"`
	LoadTestKey           string `json:"loadTestKey"`
	TestName              string `json:"testName"`

	VirtualUsers string `json:"virtualUsers"`
	Duration     string `json:"duration"`
	RampUpTime   string `json:"rampUpTime"`
	RampUpSteps  string `json:"rampUpSteps"`

	LoadEnv            LoadEnvRes             `json:"loadEnv"`
	LoadExecutionHttp  []LoadExecutionHttpRes `json:"loadExecutionHttp,omitempty"`
	LoadExecutionState LoadExecutionStateRes  `json:"loadExecutionState,omitempty"`
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

type LoadExecutionStateRes struct {
	LoadExecutionStateId uint
	LoadTestKey          string
	ExecutionStatus      constant.ExecutionStatus
	StartAt              time.Time
	EndAt                *time.Time
}

type AgentInstallInfoRes struct {
	AgentInstallInfoId uint      `json:"agentInstallInfoId,omitempty"`
	NsId               string    `json:"nsId,omitempty"`
	McisId             string    `json:"mcisId,omitempty"`
	VmId               string    `json:"vmId,omitempty"`
	Status             string    `json:"status,omitempty"`
	CreatedAt          time.Time `json:"createdAt,omitempty"`
}
