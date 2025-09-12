package load

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
)

// InstallMonitoringAgent installs a monitoring agent on specified VMs or all VM on mci.
func (l *LoadService) InstallMonitoringAgent(param MonitoringAgentInstallationParams) ([]MonitoringAgentInstallationResult, error) {
	log.Info().Msg("Starting installation of monitoring agent...")
	timeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.MonitoringAgentInstall)
	if err != nil {
		timeout = 5 * time.Minute // 기본값
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var res []MonitoringAgentInstallationResult

	scriptPath := utils.JoinRootPathWith("/script/install-server-agent.sh")
	log.Info().Msgf("Reading installation script from %s", scriptPath)
	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Error().Msgf("Failed to read installation script; %v", err)
		return res, err
	}
	username := "cb-user"

	log.Info().Msgf("Fetching mci object for NS: %s, msi id: %s", param.NsId, param.MciId)
	mci, err := l.tumblebugClient.GetMciWithContext(ctx, param.NsId, param.MciId)
	if err != nil {
		log.Error().Msgf("Failed to fetch mci; %v", err)
		return res, err
	}

	if len(mci.Vm) == 0 {
		log.Error().Msg("No VMs found on mci. Provision VM first.")
		return res, errors.New("there is no vm on mci. provision vm first")
	}

	var mapSet map[string]struct{}
	if len(param.VmIds) > 0 {
		mapSet = utils.SliceToMap(param.VmIds)
	}

	var errorCollection []error

	for _, vm := range mci.Vm {
		if mapSet != nil && !utils.Contains(mapSet, vm.Id) {
			continue
		}
		commandTimeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.CommandExecution)
		if err != nil {
			commandTimeout = 5 * time.Minute // 기본값
		}
		ctx, cancel := context.WithTimeout(ctx, commandTimeout)
		defer cancel()
		m := MonitoringAgentInfo{
			Username:  username,
			Status:    "installing",
			AgentType: "perfmon",
			NsId:      param.NsId,
			MciId:     param.MciId,
			VmId:      vm.Id,
			VmCount:   len(mci.Vm),
		}
		log.Info().Msgf("Inserting monitoring agent installation info into database vm id : %s", vm.Id)
		err = l.loadRepo.InsertMonitoringAgentInfoTx(ctx, &m)
		if err != nil {
			log.Error().Msgf("Failed to insert monitoring agent info for vm id %s : %v", vm.Id, err)
			errorCollection = append(errorCollection, err)
			continue
		}

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{string(installScript)},
			UserName: username,
		}

		log.Info().Msgf("Sending install command to mci. NS: %s, mci: %s, VMID: %s", param.NsId, param.MciId, vm.Id)
		_, err = l.tumblebugClient.CommandToVmWithContext(ctx, param.NsId, param.MciId, vm.Id, commandReq)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				m.Status = "timeout"
				log.Error().Msgf("Timeout of context. Already %s has been passed. vm id; %s", commandTimeout, vm.Id)
			} else {
				m.Status = "failed"
				log.Error().Msgf("Error occurred during command execution; %v", err)
			}
			errorCollection = append(errorCollection, err)
		} else {
			m.Status = "completed"
		}

		l.loadRepo.UpdateAgentInstallInfoStatusTx(ctx, &m)

		r := MonitoringAgentInstallationResult{
			ID:        m.ID,
			NsId:      m.NsId,
			MciId:     m.MciId,
			VmId:      m.VmId,
			VmCount:   m.VmCount,
			Status:    m.Status,
			Username:  m.Username,
			AgentType: m.AgentType,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}

		res = append(res, r)
		log.Info().Msgf(
			"Complete installing monitoring agent on mics: %s, vm: %s",
			m.MciId,
			m.VmId,
		)

		time.Sleep(time.Second)
	}

	if len(errorCollection) > 0 {
		return res, fmt.Errorf("multiple errors: %v", errorCollection)
	}

	return res, nil
}

func (l *LoadService) GetAllMonitoringAgentInfos(param GetAllMonitoringAgentInfosParam) (GetAllMonitoringAgentInfoResult, error) {
	var res GetAllMonitoringAgentInfoResult
	var monitoringAgentInfos []MonitoringAgentInstallationResult
	timeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.CommandExecution)
	if err != nil {
		timeout = 1 * time.Minute // 기본값
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info().Msgf("GetAllMonitoringAgentInfos called with param: %+v", param)
	result, totalRows, err := l.loadRepo.GetPagingMonitoringAgentInfosTx(ctx, param)

	if err != nil {
		log.Error().Msgf("Error fetching monitoring agent infos; %v", err)
		return res, err
	}

	log.Info().Msgf("Fetched %d monitoring agent infos", len(result))

	for _, monitoringAgentInfo := range result {
		var r MonitoringAgentInstallationResult
		r.ID = monitoringAgentInfo.ID
		r.NsId = monitoringAgentInfo.NsId
		r.MciId = monitoringAgentInfo.MciId
		r.VmId = monitoringAgentInfo.VmId
		r.VmCount = monitoringAgentInfo.VmCount
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

// UninstallMonitoringAgent uninstalls a monitoring agent on specified VMs or all VM on Mci.
// It takes MonitoringAgentInstallationParams as input and returns the number of affected results and any encountered error.
func (l *LoadService) UninstallMonitoringAgent(param MonitoringAgentInstallationParams) (int64, error) {

	ctx := context.Background()
	var effectedResults int64

	log.Info().Msg("Starting uninstallation of monitoring agent...")
	result, err := l.loadRepo.GetAllMonitoringAgentInfosTx(ctx, param)

	if err != nil {
		log.Error().Msgf("Failed to fetch monitoring agent information; %v", err)
		return effectedResults, err
	}

	scriptPath := utils.JoinRootPathWith("/script/remove-server-agent.sh")
	log.Info().Msgf("Reading uninstallation script from %s", scriptPath)

	uninstallPath, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Error().Msgf("Failed to read uninstallation script; %v", err)
		return effectedResults, err
	}

	username := "cb-user"

	var errorCollection []error
	for _, monitoringAgentInfo := range result {
		timeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.UninstallAgent)
		if err != nil {
			timeout = 2 * time.Minute // 기본값
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		commandReq := tumblebug.SendCommandReq{
			Command:  []string{string(uninstallPath)},
			UserName: username,
		}

		_, err = l.tumblebugClient.CommandToVmWithContext(ctx, monitoringAgentInfo.NsId, monitoringAgentInfo.MciId, monitoringAgentInfo.VmId, commandReq)
		if err != nil {
			errorCollection = append(errorCollection, err)
			log.Error().Msgf("Failed to uninstall monitoring agent on mci: %s, VM: %s - Error: %v", monitoringAgentInfo.MciId, monitoringAgentInfo.VmId, err)
			continue
		}

		err = l.loadRepo.DeleteAgentInstallInfoStatusTx(ctx, &monitoringAgentInfo)

		if err != nil {
			log.Error().Msgf("Failed to delete agent installation status for mci: %s, VM: %s - Error: %v", monitoringAgentInfo.MciId, monitoringAgentInfo.VmId, err)
			errorCollection = append(errorCollection, err)
			continue
		}

		log.Info().Msgf("Successfully uninstalled monitoring agent on mci: %s, VM: %s", monitoringAgentInfo.MciId, monitoringAgentInfo.VmId)

		time.Sleep(time.Second)
	}

	if len(errorCollection) > 0 {
		return effectedResults, fmt.Errorf("multiple errors: %v", errorCollection)
	}

	return effectedResults, nil
}
