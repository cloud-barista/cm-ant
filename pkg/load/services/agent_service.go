package services

import (
	"context"
	"errors"
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/outbound"
	"github.com/melbahja/goph"
	"log"
	"os"
	"time"
)

func InstallAgent(agentInstallReq api.AgentReq) (uint, error) {
	auth, err := goph.Key(agentInstallReq.PemKeyPath, "")
	if err != nil {
		return 0, err
	}
	client, err := goph.New(agentInstallReq.Username, agentInstallReq.PublicIp, auth)
	if err != nil {
		return 0, err
	}

	defer client.Close()

	scriptPath := configuration.JoinRootPathWith("/script/install-server-agent.sh")

	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	out, err := client.RunContext(ctx, string(installScript))

	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			log.Println(string(out))
			return 0, err
		}
	}

	log.Println(string(out))

	id, err := repository.SaveAgentInfo(&agentInstallReq)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UninstallAgent(agentId string) error {
	agentInfo, err := repository.GetAgentInfo(agentId)

	if err != nil {
		return err
	}

	auth, err := goph.Key(agentInfo.PemKeyPath, "")
	if err != nil {
		return err
	}
	client, err := goph.New(agentInfo.Username, agentInfo.PublicIp, auth)
	if err != nil {
		return err
	}

	defer client.Close()

	scriptPath := configuration.JoinRootPathWith("/script/remove-server-agent.sh")

	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	out, err := client.RunContext(context.Background(), string(installScript))

	if err != nil {
		log.Println(string(out))
		return err
	}

	log.Println(string(out))

	err = repository.DeleteAgentInfo(agentId)

	if err != nil {
		return err
	}

	return nil
}

func MockMigration(name string) error {

	createNamespaceReq := outbound.CreateNamespaceReq{
		Description: "description",
		Name:        "test01",
	}

	mcisDynamicReq := outbound.McisDynamicReq{
		Description:     "test 01 description",
		InstallMonAgent: "no",
		Label:           "DynamicVM",
		Name:            "mcis-test-01",
		SystemLabel:     "",
		VM: []outbound.VirtualMachineReq{{
			CommonImage:    "ubuntu20.04",
			CommonSpec:     "",
			ConnectionName: "asdfasdf",
			Description:    "test 01 vm description",
			Label:          "DynamicVM",
			Name:           "test-vm-t1",
			RootDiskSize:   "10",
			RootDiskType:   "default",
			SubGroupSize:   "3",
			VMUserPassword: "",
		}},
	}
	err := outbound.MockMigrate(createNamespaceReq, mcisDynamicReq)

	if err != nil {
		return err
	}
	return nil
}
