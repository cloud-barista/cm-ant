package domain

import "errors"

type LoadEnvReq struct {
	Type     AccessType `json:"type" form:"type,required"`
	NsId     string     `json:"nsId" form:"nsId,omitempty"`
	McisId   string     `json:"mcisId" form:"mcisId,omitempty"`
	Username string     `json:"username" form:"username,omitempty"`
}

func (l LoadEnvReq) Validate() error {
	if l.Type == REMOTE {
		if l.NsId == "" || l.McisId == "" || l.Username == "" {
			return errors.New("invalid user request for load env request")
		}
	}

	return nil
}

type LoadHttpReq struct {
	Method   string `json:"method" form:"method,omitempty"`
	Protocol string `json:"protocol" form:"protocol,omitempty"`
	Hostname string `json:"hostname" form:"hostname,omitempty"`
	Port     string `json:"port" form:"port,omitempty"`
	Path     string `json:"path" form:"port,omitempty"`
	BodyData string `json:"bodyData" form:"bodyData,omitempty"`
}

type LoadTestPropertyReq struct {
	PropertiesId string `json:"propertiesId" form:"propertiesId,omitempty"`

	Threads   string `json:"threads" form:"threads,omitempty"`
	RampTime  string `json:"rampTime" form:"rampTime,omitempty"`
	LoopCount string `json:"loopCount" form:"loopCount,omitempty"`

	HttpReqs LoadHttpReq `json:"httpReqs" form:"httpReqs,omitempty"`

	LoadEnvReq LoadEnvReq `json:"loadEnvReq" form:"loadEnvReq,omitempty"`
}
