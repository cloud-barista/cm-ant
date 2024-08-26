package app

import "github.com/cloud-barista/cm-ant/internal/core/common/constant"

type MonitoringAgentInstallationReq struct {
	NsId  string   `json:"nsId"`
	MciId string   `json:"mciId"`
	VmIds []string `json:"vmIds,omitempty"`
}

type GetAllMonitoringAgentInfosReq struct {
	Page  int    `query:"page"`
	Size  int    `query:"size"`
	NsId  string `query:"nsId"`
	MciId string `query:"mciId"`
	VmId  string `query:"vmId"`
}

type InstallLoadGeneratorReq struct {
	InstallLocation constant.InstallLocation `json:"installLocation"`
}

type GetAllLoadGeneratorInstallInfoReq struct {
	Page   int    `query:"page"`
	Size   int    `query:"size"`
	Status string `query:"status"`
}

type RunLoadTestReq struct {
	InstallLoadGenerator       InstallLoadGeneratorReq `json:"installLoadGenerator"`
	LoadGeneratorInstallInfoId uint                    `json:"loadGeneratorInstallInfoId"`
	TestName                   string                  `json:"testName"`
	VirtualUsers               string                  `json:"virtualUsers"`
	Duration                   string                  `json:"duration"`
	RampUpTime                 string                  `json:"rampUpTime"`
	RampUpSteps                string                  `json:"rampUpSteps"`
	Hostname                   string                  `json:"hostname"`
	Port                       string                  `json:"port"`
	AgentInstalled             bool                    `json:"agentInstalled"`
	AgentHostname              string                  `json:"agentHostname"`

	HttpReqs []RunLoadGeneratorHttpReq `json:"httpReqs,omitempty"`
}

type RunLoadGeneratorHttpReq struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`
}

type GetAllLoadTestExecutionStateReq struct {
	Page            int                      `query:"page"`
	Size            int                      `query:"size"`
	LoadTestKey     string                   `query:"loadTestKey"`
	ExecutionStatus constant.ExecutionStatus `query:"executionStatus"`
}

type GetAllLoadTestExecutionHistoryReq struct {
	Page int `query:"page"`
	Size int `query:"size"`
}

type StopLoadTestReq struct {
	LoadTestKey string `json:"loadTestKey"`
}

type GetLoadTestResultReq struct {
	LoadTestKey string                `query:"loadTestKey"`
	Format      constant.ResultFormat `query:"format"`
}
