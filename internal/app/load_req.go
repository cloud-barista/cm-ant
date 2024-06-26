package app

import "github.com/cloud-barista/cm-ant/internal/core/common/constant"

type MonitoringAgentInstallationReq struct {
	NsId   string   `json:"nsId"`
	McisId string   `json:"mcisId"`
	VmIds  []string `json:"vmIds,omitempty"`
}

type GetAllMonitoringAgentInfosReq struct {
	Page   int    `query:"page"`
	Size   int    `query:"size"`
	NsId   string `query:"nsId"`
	McisId string `query:"mcisId"`
	VmId   string `query:"vmId"`
}

type InstallLoadGeneratorReq struct {
	InstallLocation   constant.InstallLocation   `json:"installLocation,omitempty"`
	LoadGeneratorType constant.LoadGeneratorType `json:"loadGeneratorType,omitempty"`
	Username          string                     `json:"username,omitempty"`
	PublicIp          string                     `json:"publicIp,omitempty"`
	PemKeyPath        string                     `json:"pemKeyPath,omitempty"`

	NsId   string `json:"nsId,omitempty"`
	McisId string `json:"mcisId,omitempty"`
	VmId   string `json:"vmId,omitempty"`
}
