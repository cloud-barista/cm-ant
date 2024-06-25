package load

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	u "github.com/cloud-barista/cm-ant/pkg/utils"
)

// LoadService represents a service for managing load operations.
type LoadService struct {
	loadRepo        *LoadRepository
	tumblebugClient *tumblebug.TumblebugClient
}

// NewLoadService creates a new instance of LoadService.
func NewLoadService(loadRepo *LoadRepository, client *tumblebug.TumblebugClient) *LoadService {
	return &LoadService{
		loadRepo:        loadRepo,
		tumblebugClient: client,
	}
}

// MonitoringAgentInstallationParams represents parameters for installing a monitoring agent.
type MonitoringAgentInstallationParams struct {
	NsId   string `json:"nsId"`
	McisId string `json:"mcisId"`
}

// MonitoringAgentInstallationResult represents the result of a monitoring agent installation.
type MonitoringAgentInstallationResult struct {
	ID      uint   `json:"id"`
	NsId    string `json:"nsId"`
	McisId  string `json:"mcisId"`
	VmCount int    `json:"vmCount"`
	Status  string `json:"status"`
}

// InstallMonitoringAgent installs a monitoring agent on specified VMs.
func (l *LoadService) InstallMonitoringAgent(param MonitoringAgentInstallationParams) (MonitoringAgentInstallationResult, error) {
	log.Println("[INFO] Starting installation of monitoring agent...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var res MonitoringAgentInstallationResult

	scriptPath := u.JoinRootPathWith("/script/install-server-agent.sh")
	log.Printf("[INFO] Reading installation script from %s", scriptPath)
	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Printf("[ERROR] Failed to read installation script: %v", err)
		return res, err
	}

	log.Printf("[INFO] Fetching VM IDs for NS: %s, MCIS: %s", param.NsId, param.McisId)
	vmIds, err := l.tumblebugClient.GetMcisIdsWithContext(ctx, param.NsId, param.McisId)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch VM IDs: %v", err)
		return res, err
	}

	if len(vmIds) == 0 {
		log.Println("[ERROR] No VMs found. Provision VM first.")
		return res, errors.New("provision vm first")
	}

	username := "cb-user"
	monitoringAgentInstallationInfo := MonitoringAgentInfo{
		Username:          username,
		Status:            "installing",
		AgentType:         "perfmon",
		AdditionalNsId:    param.NsId,
		AdditionalMcisId:  param.McisId,
		AdditionalVmId:    strings.Join(vmIds, ","),
		AdditionalVmCount: len(vmIds),
	}

	utils.LogInfo("Inserting monitoring agent installation info into database")
	err = l.loadRepo.InsertMonitoringAgentInfoTx(ctx, &monitoringAgentInstallationInfo)
	if err != nil {
		utils.LogErrorf("[ERROR] Failed to insert monitoring agent info: %v", err)
		return res, err
	}

	commandReq := tumblebug.SendCommandReq{
		Command:  []string{string(installScript)},
		UserName: username,
	}

	utils.LogInfof("Sending install command to MCIS. NS: %s, MCIS: %s", param.NsId, param.McisId)
	_, err = l.tumblebugClient.CommandToMcisWithContext(ctx, param.NsId, param.McisId, commandReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			monitoringAgentInstallationInfo.Status = "invalid"
			l.loadRepo.UpdateAgentInstallInfoStatusTx(ctx, &monitoringAgentInstallationInfo)
			utils.LogError("Timeout of context. Already 15 seconds has been passed.")
			return res, fmt.Errorf("invalid installation info: %w", err)
		} else {
			utils.LogErrorf("Error occurred during command execution: %v", err)
			monitoringAgentInstallationInfo.Status = "failed"
		}
		l.loadRepo.UpdateAgentInstallInfoStatusTx(ctx, &monitoringAgentInstallationInfo)
		return res, err
	}

	monitoringAgentInstallationInfo.Status = "completed"
	res.ID = monitoringAgentInstallationInfo.ID
	res.NsId = monitoringAgentInstallationInfo.AdditionalNsId
	res.McisId = monitoringAgentInstallationInfo.AdditionalMcisId
	res.VmCount = monitoringAgentInstallationInfo.AdditionalVmCount
	res.Status = monitoringAgentInstallationInfo.Status

	l.loadRepo.UpdateAgentInstallInfoStatusTx(ctx, &monitoringAgentInstallationInfo)
	utils.LogInfof("Complete installing monitoring agent on mics: %s", monitoringAgentInstallationInfo.AdditionalMcisId)
	return res, nil
}
