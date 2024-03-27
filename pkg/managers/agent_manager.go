package managers

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	antsys "github.com/cloud-barista/cm-ant/pkg/utils"
	"log"
)

const (
	agentWorkDir = "AGENT_WORK_DIR=/opt/perfmon-agent"
)

type AgentManager interface {
	Install(model.AgentInfo) error
	Start(model.AgentInfo) error
	Stop(model.AgentInfo) error
	Remove(model.AgentInfo) error
}

func NewAgentManager() AgentManager {
	return &LocalAgentManager{}
}

type LocalAgentManager struct {
}

func (l LocalAgentManager) Install(agentInfo model.AgentInfo) error {
	installCmd := fmt.Sprintf("%s source ./script/install-server-agent.sh", agentWorkDir)
	err := antsys.InlineCmd(installCmd)

	if err != nil {
		log.Printf("error while installing server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully installed on localhost;")
	return nil
}

func (l LocalAgentManager) Start(agentInfo model.AgentInfo) error {

	shutdown := agentInfo.Shutdown
	autoShutdown := ""

	if shutdown {
		autoShutdown = "--auto-shutdown"
	}

	agentInfo.TcpPort = "4444"

	startCmd := fmt.Sprintf("%s TCP_PORT=%s AUTO_SHUTDOWN=%s source ./script/start-server-agent.sh", agentWorkDir, agentInfo.TcpPort, autoShutdown)
	err := antsys.InlineCmdAsync(startCmd)

	if err != nil {
		log.Printf("error while start server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully started on localhost;")
	return nil
}

func (l LocalAgentManager) Stop(agentInfo model.AgentInfo) error {

	agentInfo.TcpPort = "4444"

	stopCmd := fmt.Sprintf("%s TCP_PORT=%s source ./script/stop-server-agent.sh", agentWorkDir, agentInfo.TcpPort)
	err := antsys.InlineCmd(stopCmd)

	if err != nil {
		log.Printf("error while stop server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully stopped on localhost;")
	return nil
}

func (l LocalAgentManager) Remove(agentInfo model.AgentInfo) error {

	stopCmd := fmt.Sprintf("%s source ./script/remove-server-agent.sh", agentWorkDir)
	err := antsys.InlineCmd(stopCmd)

	if err != nil {
		log.Printf("error while remove server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully remove on localhost;")
	return nil
}
