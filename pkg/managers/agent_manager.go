package managers

import (
	"fmt"
	"log"

	"github.com/cloud-barista/cm-ant/internal/domain"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

const (
	agentWorkDir = "AGENT_WORK_DIR=/opt/perfmon-agent"
)

type AgentManager interface {
	Install(domain.AgentInfo) error
	Start(domain.AgentInfo) error
	Stop(domain.AgentInfo) error
	Remove(domain.AgentInfo) error
}

func NewAgentManager() AgentManager {
	return &LocalAgentManager{}
}

type LocalAgentManager struct {
}

func (l LocalAgentManager) Install(agentInfo domain.AgentInfo) error {
	installCmd := fmt.Sprintf("%s source ./script/install-server-agent.sh", agentWorkDir)
	err := utils.SyncSysCall(installCmd)

	if err != nil {
		log.Printf("error while installing server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully installed on localhost;")
	return nil
}

func (l LocalAgentManager) Start(agentInfo domain.AgentInfo) error {

	shutdown := agentInfo.Shutdown
	autoShutdown := ""

	if shutdown {
		autoShutdown = "--auto-shutdown"
	}

	agentInfo.TcpPort = "4444"

	startCmd := fmt.Sprintf("%s TCP_PORT=%s AUTO_SHUTDOWN=%s source ./script/start-server-agent.sh", agentWorkDir, agentInfo.TcpPort, autoShutdown)
	err := utils.AsyncSysCall(startCmd)

	if err != nil {
		log.Printf("error while start server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully started on localhost;")
	return nil
}

func (l LocalAgentManager) Stop(agentInfo domain.AgentInfo) error {

	agentInfo.TcpPort = "4444"

	stopCmd := fmt.Sprintf("%s TCP_PORT=%s source ./script/stop-server-agent.sh", agentWorkDir, agentInfo.TcpPort)
	err := utils.SyncSysCall(stopCmd)

	if err != nil {
		log.Printf("error while stop server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully stopped on localhost;")
	return nil
}

func (l LocalAgentManager) Remove(agentInfo domain.AgentInfo) error {

	stopCmd := fmt.Sprintf("%s source ./script/remove-server-agent.sh", agentWorkDir)
	err := utils.SyncSysCall(stopCmd)

	if err != nil {
		log.Printf("error while remove server agent; %s\n", err)
		return err
	}

	log.Println("[CM-ANT] server agent successfully remove on localhost;")
	return nil
}
