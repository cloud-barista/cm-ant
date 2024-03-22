package domain

type AgentInfo struct {
	AgentId string `json:"agentId" form:"agentId,omitempty"`

	Hostname string `json:"hostname" form:"hostname,omitempty"`
	Username string `json:"username" form:"username,omitempty"`
	TcpPort  string `json:"tcpPort" form:"tcpPort,omitempty"`
	Shutdown bool   `json:"shutdown" form:"shutdown,omitempty"`
}

func NewAgentInfo() AgentInfo {
	return AgentInfo{}
}

type LoadTest struct {
	TestId string `json:"testId" form:"TestId,omitempty"`

	Protocol string `json:"protocol" form:"protocol,omitempty"`
	Hostname string `json:"hostname" form:"hostname,omitempty"`
	Port     string `json:"port" form:"port,omitempty"`
	Path     string `json:"path" form:"port,omitempty"`
	BodyData string `json:"bodyData" form:"bodyData,omitempty"`

	Threads   string `json:"threads" form:"threads,omitempty"`
	RampTime  string `json:"rampTime" form:"rampTime,omitempty"`
	LoopCount string `json:"loopCount" form:"loopCount,omitempty"`
}

func NewLoadTest() LoadTest {
	return LoadTest{}
}
