package load

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
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
	NsId   string   `json:"nsId"`
	McisId string   `json:"mcisId"`
	VmIds  []string `json:"vmIds,omitempty"`
}

// MonitoringAgentInstallationResult represents the result of a monitoring agent installation.
type MonitoringAgentInstallationResult struct {
	ID        uint      `json:"id"`
	NsId      string    `json:"nsId"`
	McisId    string    `json:"mcisId"`
	VmId      string    `json:"vmId"`
	VmCount   int       `json:"vmCount"`
	Status    string    `json:"status"`
	Username  string    `json:"username"`
	AgentType string    `json:"agentType"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// InstallMonitoringAgent installs a monitoring agent on specified VMs or all VM on Mcis.
func (l *LoadService) InstallMonitoringAgent(param MonitoringAgentInstallationParams) ([]MonitoringAgentInstallationResult, error) {
	utils.LogInfo("Starting installation of monitoring agent...")

	var res []MonitoringAgentInstallationResult

	scriptPath := u.JoinRootPathWith("/script/install-server-agent.sh")
	utils.LogInfof("Reading installation script from %s", scriptPath)
	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Printf("[ERROR] Failed to read installation script: %v", err)
		return res, err
	}
	username := "cb-user"

	utils.LogInfof("Fetching Mcis object for NS: %s, MCIS: %s", param.NsId, param.McisId)
	mcis, err := l.tumblebugClient.GetMcisWithContext(context.Background(), param.NsId, param.McisId)
	if err != nil {
		utils.LogErrorf("Failed to fetch MCIS : %v", err)
		return res, err
	}

	if len(mcis.VMs) == 0 {
		utils.LogErrorf("No VMs found on mcis. Provision VM first.")
		return res, errors.New("there is no vm on mcis. provision vm first")
	}

	var mapSet map[string]struct{}
	if param.VmIds != nil && len(param.VmIds) > 0 {
		mapSet = utils.SliceToMap(param.VmIds)
	}

	var errorCollection []error

	for _, vm := range mcis.VMs {
		if mapSet != nil && !utils.Contains(mapSet, vm.ID) {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		m := MonitoringAgentInfo{
			Username:          username,
			Status:            "installing",
			AgentType:         "perfmon",
			AdditionalNsId:    param.NsId,
			AdditionalMcisId:  param.McisId,
			AdditionalVmId:    vm.ID,
			AdditionalVmCount: len(mcis.VMs),
		}
		utils.LogInfof("Inserting monitoring agent installation info into database vm id : %s", vm.ID)
		err = l.loadRepo.InsertMonitoringAgentInfoTx(ctx, &m)
		if err != nil {
			utils.LogErrorf("Failed to insert monitoring agent info for vm id %s : %v", vm.ID, err)
			errorCollection = append(errorCollection, err)
			continue
		}

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{string(installScript)},
			UserName: username,
		}

		utils.LogInfof("Sending install command to MCIS. NS: %s, MCIS: %s, VMID: %s", param.NsId, param.McisId, vm.ID)
		_, err = l.tumblebugClient.CommandToVmWithContext(ctx, param.NsId, param.McisId, vm.ID, commandReq)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				m.Status = "timeout"
				utils.LogErrorf("Timeout of context. Already 15 seconds has been passed. vm id : %s", vm.ID)
			} else {
				m.Status = "failed"
				utils.LogErrorf("Error occurred during command execution: %v", err)
			}
			errorCollection = append(errorCollection, err)
		} else {
			m.Status = "completed"
		}

		l.loadRepo.UpdateAgentInstallInfoStatusTx(ctx, &m)

		r := MonitoringAgentInstallationResult{
			ID:        m.ID,
			NsId:      m.AdditionalNsId,
			McisId:    m.AdditionalMcisId,
			VmId:      m.AdditionalVmId,
			VmCount:   m.AdditionalVmCount,
			Status:    m.Status,
			Username:  m.Username,
			AgentType: m.AgentType,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}

		res = append(res, r)
		utils.LogInfof(
			"Complete installing monitoring agent on mics: %s, vm: %s",
			m.AdditionalMcisId,
			m.AdditionalVmId,
		)

		time.Sleep(time.Second)
	}

	if len(errorCollection) > 0 {
		return res, fmt.Errorf("multiple errors: %v", errorCollection)
	}

	return res, nil
}

type GetAllMonitoringAgentInfosParam struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	NsId   string `json:"nsId,omitempty"`
	McisId string `json:"mcisId,omitempty"`
	VmId   string `json:"vmId,omitempty"`
}

type GetAllMonitoringAgentInfoResult struct {
	MonitoringAgentInfos []MonitoringAgentInstallationResult `json:"monitoringAgentInfos"`
	TotalRow             int64                               `json:"totalRow"`
}

func (l *LoadService) GetAllMonitoringAgentInfos(param GetAllMonitoringAgentInfosParam) (GetAllMonitoringAgentInfoResult, error) {
	var res GetAllMonitoringAgentInfoResult
	var monitoringAgentInfos []MonitoringAgentInstallationResult
	ctx := context.Background()

	utils.LogInfof("GetAllMonitoringAgentInfos called with param: %+v", param)
	result, totalRows, err := l.loadRepo.GetPagingMonitoringAgentInfosTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching monitoring agent infos: %v", err)
		return res, err
	}

	utils.LogInfof("Fetched %d monitoring agent infos", len(result))

	for _, monitoringAgentInfo := range result {
		var r MonitoringAgentInstallationResult
		r.ID = monitoringAgentInfo.ID
		r.NsId = monitoringAgentInfo.AdditionalNsId
		r.McisId = monitoringAgentInfo.AdditionalMcisId
		r.VmId = monitoringAgentInfo.AdditionalVmId
		r.VmCount = monitoringAgentInfo.AdditionalVmCount
		r.Status = monitoringAgentInfo.Status
		r.Username = monitoringAgentInfo.Username
		r.AgentType = monitoringAgentInfo.AgentType
		r.CreatedAt = monitoringAgentInfo.CreatedAt
		r.UpdatedAt = monitoringAgentInfo.UpdatedAt
		monitoringAgentInfos = append(monitoringAgentInfos, r)
	}

	res.MonitoringAgentInfos = monitoringAgentInfos
	res.TotalRow = totalRows

	return res, nil
}

// UninstallMonitoringAgent uninstalls a monitoring agent on specified VMs or all VM on Mcis.
// It takes MonitoringAgentInstallationParams as input and returns the number of affected results and any encountered error.
func (l *LoadService) UninstallMonitoringAgent(param MonitoringAgentInstallationParams) (int64, error) {

	ctx := context.Background()
	var effectedResults int64

	utils.LogInfo("Starting uninstallation of monitoring agent...")
	result, err := l.loadRepo.GetAllMonitoringAgentInfosTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Failed to fetch monitoring agent information: %v", err)
		return effectedResults, err
	}

	scriptPath := utils.JoinRootPathWith("/script/remove-server-agent.sh")
	utils.LogInfof("Reading uninstallation script from %s", scriptPath)

	uninstallPath, err := os.ReadFile(scriptPath)
	if err != nil {
		utils.LogErrorf("Failed to read uninstallation script: %v", err)
		return effectedResults, err
	}

	username := "cb-user"

	var errorCollection []error
	for _, monitoringAgentInfo := range result {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{string(uninstallPath)},
			UserName: username,
		}

		_, err = l.tumblebugClient.CommandToVmWithContext(ctx, monitoringAgentInfo.AdditionalNsId, monitoringAgentInfo.AdditionalMcisId, monitoringAgentInfo.AdditionalVmId, commandReq)
		if err != nil {
			errorCollection = append(errorCollection, err)
			utils.LogErrorf("Failed to uninstall monitoring agent on Mcis: %s, VM: %s - Error: %v", monitoringAgentInfo.AdditionalMcisId, monitoringAgentInfo.AdditionalVmId, err)
			continue
		}

		err = l.loadRepo.DeleteAgentInstallInfoStatusTx(ctx, &monitoringAgentInfo)

		if err != nil {
			utils.LogErrorf("Failed to delete agent installation status for Mcis: %s, VM: %s - Error: %v", monitoringAgentInfo.AdditionalMcisId, monitoringAgentInfo.AdditionalVmId, err)
			errorCollection = append(errorCollection, err)
			continue
		}

		utils.LogInfof("Successfully uninstalled monitoring agent on Mcis: %s, VM: %s", monitoringAgentInfo.AdditionalMcisId, monitoringAgentInfo.AdditionalVmId)

		time.Sleep(time.Second)
	}

	if len(errorCollection) > 0 {
		return effectedResults, fmt.Errorf("multiple errors: %v", errorCollection)
	}

	return effectedResults, nil
}
