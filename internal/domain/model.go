package domain

type AgentInfo struct {
	AgentId string `form:"agentId,omitempty"`

	Hostname string `form:"hostname,omitempty"`
	Username string `form:"username,omitempty"`
	TcpPort  string `form:"tcpPort,omitempty"`
	Shutdown bool   `form:"shutdown,omitempty"`
}

func NewAgentInfo() AgentInfo {
	return AgentInfo{}
}

type LoadTestProperties struct {
	TestId string `form:"testId,omitempty"`

	Protocol string `form:"protocol,omitempty"`
	Hostname string `form:"hostname,omitempty"`
	Port     string `form:"port,omitempty"`
	Path     string `form:"port,omitempty"`
	BodyData string `form:"bodyData,omitempty"`

	Threads   string `form:"threads,omitempty"`
	RampTime  string `form:"rampTime,omitempty"`
	LoopCount string `form:"loopCount,omitempty"`

	Scheduled bool   `form:"scheduled,omitempty"`
	Infinite  bool   `form:"infinite,omitempty"`
	Duration  string `form:"duration,omitempty"`

	AgentHost string `form:"agentHost,omitempty"`
	AgentPort string `form:"agentPort,omitempty"`
}

func NewLoadTestProperties() LoadTestProperties {
	return LoadTestProperties{}
}
