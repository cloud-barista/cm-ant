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
	// localhost for installing load generator; local | remote
	InstallLoadGenerator InstallLoadGeneratorReq `json:"installLoadGenerator"`

	// if already installed load generator simply put this field
	LoadGeneratorInstallInfoId uint `json:"loadGeneratorInstallInfoId"`

	// test scenario
	TestName     string `json:"testName"`
	VirtualUsers string `json:"virtualUsers"`
	Duration     string `json:"duration"`
	RampUpTime   string `json:"rampUpTime"`
	RampUpSteps  string `json:"rampUpSteps"`

	// for validate agent host and connect to tumblebug resources
	NsId  string `json:"nsId"`  // for metadata usage
	MciId string `json:"mciId"` // for metadata usage
	VmId  string `json:"vmId"`  // for metadata usage

	// agent tcp default port is 5555
	CollectAdditionalSystemMetrics bool   `json:"collectAdditionalSystemMetrics"`
	AgentHostname                  string `json:"agentHostname"` // basically, it is same as host for vm

	HttpReqs []RunLoadGeneratorHttpReq `json:"httpReqs,omitempty"`
}

type RunLoadGeneratorHttpReq struct {
	Method   string `json:"method" validate:"required"`             // GET or POST
	Protocol string `json:"protocol" validate:"required"`           // http or https
	Hostname string `json:"hostname,omitempty" validate:"required"` // xx.xx.xx.xx or asx.bbb.com
	Port     string `json:"port,omitempty" validate:"required"`     // 1 ~ 65353
	Path     string `json:"path,omitempty" validate:"required"`     // /xxx/www/sss or possibly empty
	BodyData string `json:"bodyData,omitempty" validate:"required"` // {"xxx": "tttt", "wwwww": "wotjkenr"}
}

type GetAllLoadTestExecutionStateReq struct {
	Page            int                      `query:"page"`
	Size            int                      `query:"size"`
	LoadTestKey     string                   `query:"loadTestKey"`
	ExecutionStatus constant.ExecutionStatus `query:"executionStatus"`
}

type GetLastLoadTestExecutionStateReq struct {
	NsId  string `query:"nsId"`
	MciId string `query:"mciId"`
	VmId  string `query:"vmId"`
}

type GetAllLoadTestExecutionHistoryReq struct {
	Page int `query:"page"`
	Size int `query:"size"`
}

type StopLoadTestReq struct {
	LoadTestKey string `json:"loadTestKey"`
	NsId        string `json:"nsId"`
	MciId       string `json:"mciId"`
	VmId        string `json:"vmId"`
}

type GetLoadTestResultReq struct {
	LoadTestKey string                `query:"loadTestKey"`
	Format      constant.ResultFormat `query:"format"`
}

type GetLastLoadTestResultReq struct {
	NsId   string                `query:"nsId"`
	MciId  string                `query:"mciId"`
	VmId   string                `query:"vmId"`
	Format constant.ResultFormat `query:"format"`
}
