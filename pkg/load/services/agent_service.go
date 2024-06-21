package services

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

func InstallAgent(agentReq api.AntTargetServerReq) error {
	scriptPath := utils.JoinRootPathWith("/script/install-server-agent.sh")

	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	mcisObject, err := tumblebug.GetMcisObjectWithContext(ctx, agentReq.NsId, agentReq.McisId)

	if err != nil {
		return err
	}

	if len(mcisObject.VMs) == 0 {
		return errors.New("provision vm first")
	}

	vm := mcisObject.VMs[0]

	if mcisObject.IsRunning(mcisObject.VmId()) {
		return errors.New("start vm first")
	}

	var vmId string

	if agentReq.VmId == "" {
		vmId = vm.ID
	} else {
		vmId = agentReq.VmId
	}

	vmUserAccount := vm.VMUserAccount

	agentInstallInfo := model.AgentInstallInfo{
		NsId:     agentReq.NsId,
		McisId:   agentReq.McisId,
		VmId:     vmId,
		Username: vmUserAccount,
		Status:   "install",
	}

	err = repository.InsertAgentInstallInfo(&agentInstallInfo)

	if err != nil {
		return err
	}

	commandReq := tumblebug.SendCommandReq{
		Command:  []string{string(installScript)},
		UserName: vmUserAccount,
	}

	_, err = tumblebug.CommandToVmWithContext(ctx, agentReq.NsId, agentReq.McisId, vmId, commandReq)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			agentInstallInfo.Status = "completed"
			repository.UpdateAgentInstallInfoStatus(&agentInstallInfo)
			return nil
		}

		log.Printf("error occured; %s\n", err)
		agentInstallInfo.Status = "failed"
		repository.UpdateAgentInstallInfoStatus(&agentInstallInfo)
		return err
	}

	agentInstallInfo.Status = "invalid"
	repository.UpdateAgentInstallInfoStatus(&agentInstallInfo)

	return errors.New("invalid installation info")
}

func GetAllAgentInstallInfo() ([]api.AgentInstallInfoRes, error) {

	result, err := repository.GetAllAgentInstallInfos()

	if err != nil {
		return nil, err
	}

	var agentInstallInfos []api.AgentInstallInfoRes

	for _, agentInstallInfo := range result {
		var a api.AgentInstallInfoRes
		a.AgentInstallInfoId = agentInstallInfo.Model.ID
		a.NsId = agentInstallInfo.NsId
		a.McisId = agentInstallInfo.McisId
		a.VmId = agentInstallInfo.VmId
		a.Status = agentInstallInfo.Status
		a.CreatedAt = agentInstallInfo.Model.CreatedAt

		agentInstallInfos = append(agentInstallInfos, a)
	}

	return agentInstallInfos, nil
}

func UninstallAgent(agentInstallInfoId string) error {
	agentInstallInfo, err := repository.GetAgentInstallInfo(agentInstallInfoId)
	if err != nil {
		return err
	}

	scriptPath := utils.JoinRootPathWith("/script/remove-server-agent.sh")

	uninstallPath, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	commandReq := tumblebug.SendCommandReq{
		Command:  []string{string(uninstallPath)},
		UserName: agentInstallInfo.Username,
	}

	_, err = tumblebug.CommandToVm(agentInstallInfo.NsId, agentInstallInfo.McisId, agentInstallInfo.VmId, commandReq)

	if err != nil {
		return err
	}

	err = repository.DeleteAgentInstallInfo(agentInstallInfoId)

	if err != nil {
		return err
	}

	return nil
}
