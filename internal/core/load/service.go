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
			Username:  username,
			Status:    "installing",
			AgentType: "perfmon",
			NsId:      param.NsId,
			McisId:    param.McisId,
			VmId:      vm.ID,
			VmCount:   len(mcis.VMs),
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
			NsId:      m.NsId,
			McisId:    m.McisId,
			VmId:      m.VmId,
			VmCount:   m.VmCount,
			Status:    m.Status,
			Username:  m.Username,
			AgentType: m.AgentType,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}

		res = append(res, r)
		utils.LogInfof(
			"Complete installing monitoring agent on mics: %s, vm: %s",
			m.McisId,
			m.VmId,
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
		r.NsId = monitoringAgentInfo.NsId
		r.McisId = monitoringAgentInfo.McisId
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

		_, err = l.tumblebugClient.CommandToVmWithContext(ctx, monitoringAgentInfo.NsId, monitoringAgentInfo.McisId, monitoringAgentInfo.VmId, commandReq)
		if err != nil {
			errorCollection = append(errorCollection, err)
			utils.LogErrorf("Failed to uninstall monitoring agent on Mcis: %s, VM: %s - Error: %v", monitoringAgentInfo.McisId, monitoringAgentInfo.VmId, err)
			continue
		}

		err = l.loadRepo.DeleteAgentInstallInfoStatusTx(ctx, &monitoringAgentInfo)

		if err != nil {
			utils.LogErrorf("Failed to delete agent installation status for Mcis: %s, VM: %s - Error: %v", monitoringAgentInfo.McisId, monitoringAgentInfo.VmId, err)
			errorCollection = append(errorCollection, err)
			continue
		}

		utils.LogInfof("Successfully uninstalled monitoring agent on Mcis: %s, VM: %s", monitoringAgentInfo.McisId, monitoringAgentInfo.VmId)

		time.Sleep(time.Second)
	}

	if len(errorCollection) > 0 {
		return effectedResults, fmt.Errorf("multiple errors: %v", errorCollection)
	}

	return effectedResults, nil
}

func (l *LoadService) InstallLoadTester() error {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// defer cancel()

	// var loadEnvReq api.LoadEnvReq

	// 1. check the vm is provisioned or not and the state is under running or not.
	// 2. if the state is not valid, then provision the vm or run the vm.
	// 3. after validation check, install load generator on running vm.
	// TODO [medium] consider how to cluster load generator server and load generator.
	// load generator have to cluster depends on the load test scenario.

	// if antLoadEnvReq.InstallLocation == constant.Remote {
	// 	antTargetServerMcis, err := tumblebug.GetMcisObjectWithContext(ctx, antLoadEnvReq.NsId, antLoadEnvReq.McisId)

	// 	if err != nil {
	// 		return 0, err
	// 	}

	// 	if len(antTargetServerMcis.VMs) == 0 {
	// 		return 0, errors.New("cannot find any vm in target mcis")
	// 	}

	// 	var antTargetServerVm tumblebug.VmRes

	// 	for _, v := range antTargetServerMcis.VMs {
	// 		if strings.EqualFold(v.ID, antLoadEnvReq.VmId) {
	// 			antTargetServerVm = v
	// 		}
	// 	}

	// 	if antTargetServerVm.ID == "" {
	// 		return 0, fmt.Errorf("%s does not exist", antLoadEnvReq.VmId)
	// 	}

	// 	connectionId := antTargetServerVm.ConnectionName
	// 	antNsId := antLoadEnvReq.NsId
	// 	antVNetId := antTargetServerVm.VNetID
	// 	antSubnetId := antTargetServerVm.SubnetID
	// 	antCspRegion := antTargetServerVm.Region.Region
	// 	antCspImageId, ok := regionImageMap[antCspRegion]
	// 	antUsername := "cb-user"
	// 	antSgId := "ant-load-test-sg"
	// 	antSshId := "ant-load-test-ssh"
	// 	antImageId := "ant-load-test-image"
	// 	antSpecId := "ant-load-test-spec"
	// 	antCspSpecName := "t3.small"
	// 	antVmId := "ant-load-test-vm"
	// 	antMcisId := "ant-load-test-mcis"

	// 	log.Println(connectionId, antNsId, antVNetId, antSubnetId, antUsername, antCspRegion, antCspImageId, antSgId, antSshId, antImageId, antSpecId, antCspSpecName, antVmId, antMcisId)

	// 	if !ok {
	// 		return 0, errors.New("region base ubuntu 22.04 lts image doesn't exist")
	// 	}

	// 	var wg sync.WaitGroup
	// 	goroutine := 4
	// 	errChan := make(chan error, goroutine)

	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()

	// 		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
	// 		defer c()

	// 		err2 := tumblebug.GetSecurityGroupWithContext(tc, antNsId, antSgId)
	// 		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
	// 			errChan <- err2
	// 			return
	// 		} else if err2 == nil {
	// 			return
	// 		}

	// 		sg := tumblebug.SecurityGroupReq{
	// 			Name:           antSgId,
	// 			ConnectionName: connectionId,
	// 			VNetID:         antVNetId,
	// 			Description:    "Default Security Group for Ant load test",
	// 			FirewallRules: []tumblebug.FirewallRuleReq{
	// 				{FromPort: "22", ToPort: "22", IPProtocol: "tcp", Direction: "inbound", CIDR: "0.0.0.0/0"},
	// 			},
	// 		}
	// 		_, err2 = tumblebug.CreateSecurityGroupWithContext(tc, antNsId, sg)
	// 		if err2 != nil {
	// 			errChan <- err2
	// 			return
	// 		}
	// 	}()

	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()

	// 		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
	// 		defer c()

	// 		_, err2 := tumblebug.GetSecureShellWithContext(tc, antNsId, antSshId)
	// 		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
	// 			errChan <- err2
	// 			return
	// 		} else if err2 == nil {
	// 			return
	// 		}

	// 		ssh := tumblebug.SecureShellReq{
	// 			ConnectionName: connectionId,
	// 			Name:           antSshId,
	// 			Username:       antUsername,
	// 			Description:    "Default secure shell key for Ant load test",
	// 		}
	// 		sshResult, err2 := tumblebug.CreateSecureShellWithContext(ctx, antNsId, ssh)
	// 		if err2 != nil {
	// 			errChan <- err2
	// 			return
	// 		}
	// 		home, err2 := os.UserHomeDir()
	// 		if err2 != nil {
	// 			errChan <- err2
	// 			return
	// 		}
	// 		pemFilePath := fmt.Sprintf("%s/.ssh/%s.pem", home, sshResult.Id)

	// 		err2 = os.WriteFile(pemFilePath, []byte(sshResult.PrivateKey), 0600)
	// 		if err2 != nil {
	// 			errChan <- err2
	// 			return
	// 		}

	// 		log.Printf("%s.pem ssh private key file save to default ssh path", sshResult.Id)
	// 	}()

	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()

	// 		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
	// 		defer c()

	// 		err2 := tumblebug.GetImageWithContext(tc, antNsId, antImageId)
	// 		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
	// 			errChan <- err2
	// 			return
	// 		} else if err2 == nil {
	// 			return
	// 		}
	// 		// TODO add dynamic spec integration in advanced version
	// 		image := tumblebug.ImageReq{
	// 			ConnectionName: connectionId,
	// 			Name:           antImageId,
	// 			CspImageId:     antCspImageId,
	// 			Description:    "Default machine image for Ant load test",
	// 		}
	// 		_, err2 = tumblebug.CreateImageWithContext(ctx, antNsId, image)
	// 		if err2 != nil {
	// 			errChan <- err2
	// 			return
	// 		}
	// 	}()

	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()

	// 		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
	// 		defer c()

	// 		err2 := tumblebug.GetSpecWithContext(tc, antNsId, antSpecId)
	// 		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
	// 			errChan <- err2
	// 			return
	// 		} else if err2 == nil {
	// 			return
	// 		}

	// 		// TODO add dynamic spec integration in advanced version
	// 		spec := tumblebug.SpecReq{
	// 			ConnectionName: connectionId,
	// 			Name:           antSpecId,
	// 			CspSpecName:    antCspSpecName,
	// 			Description:    "Default spec for Ant load test",
	// 		}
	// 		_, err2 = tumblebug.CreateSpecWithContext(ctx, antNsId, spec)
	// 		if err2 != nil {
	// 			errChan <- err2
	// 			return
	// 		}
	// 	}()

	// 	wg.Wait()
	// 	close(errChan)

	// 	if len(errChan) != 0 {
	// 		err = <-errChan
	// 		return 0, err
	// 	}

	// 	antLoadGenerateServerMcis, err := tumblebug.GetMcisWithContext(ctx, antNsId, antMcisId)

	// 	if err != nil && !errors.Is(err, tumblebug.ResourcesNotFound) {
	// 		return 0, err
	// 	} else if err == nil {
	// 		if !antLoadGenerateServerMcis.IsRunning(antVmId) {
	// 			// TODO - need to change mcis or antTargetServerVm status execution
	// 			return 0, errors.New("vm is not running condition")
	// 		}
	// 	} else {
	// 		antLoadGenerateServerReq := tumblebug.McisReq{
	// 			Name:            antMcisId,
	// 			Description:     "Default mcis for Ant load test",
	// 			InstallMonAgent: "no",
	// 			Label:           "ANT",
	// 			SystemLabel:     "ANT",
	// 			Vm: []tumblebug.VmReq{
	// 				{
	// 					SubGroupSize:     "1",
	// 					Name:             antVmId,
	// 					ImageId:          antImageId,
	// 					VmUserAccount:    antUsername,
	// 					ConnectionName:   connectionId,
	// 					SshKeyId:         antSshId,
	// 					SpecId:           antSpecId,
	// 					SecurityGroupIds: []string{antSgId},
	// 					VNetId:           antVNetId,
	// 					SubnetId:         antSubnetId,
	// 					Description:      "Default vm for Ant load test",
	// 					VmUserPassword:   "",
	// 					RootDiskType:     "default",
	// 					RootDiskSize:     "default",
	// 				},
	// 			},
	// 		}

	// 		antLoadGenerateServerMcis, err = tumblebug.CreateMcisWithContext(ctx, antNsId, antLoadGenerateServerReq)

	// 		if err != nil {
	// 			return 0, err
	// 		}

	// 		log.Println("******************created*******************\n", antLoadGenerateServerMcis)
	// 		time.Sleep(3 * time.Second)
	// 	}

	// 	log.Println("ant load generate server mcis is ready with running condition")

	// 	ssh, err := tumblebug.GetSecureShellWithContext(ctx, antNsId, antSshId)
	// 	if err != nil {
	// 		return 0, err
	// 	}

	// 	vm := antLoadGenerateServerMcis.VMs[0]

	// 	loadEnvReq = api.LoadEnvReq{
	// 		InstallLocation: constant.Remote,
	// 		NsId:            antNsId,
	// 		McisId:          antLoadGenerateServerMcis.ID,
	// 		VmId:            antLoadGenerateServerMcis.VmId(),
	// 		Username:        antUsername,
	// 		PublicIp:        vm.PublicIP,
	// 		PemKeyPath:      filepath.Join(os.Getenv("HOME"), ".ssh", ssh.Id+".pem"),
	// 	}
	// } else {
	// 	loadEnvReq = api.LoadEnvReq{
	// 		InstallLocation: constant.Local,
	// 	}
	// }

	// manager := managers.NewLoadTestManager()

	// err := manager.Install(&loadEnvReq)
	// if err != nil {
	// 	return 0, err
	// }

	// loadEnv := model.LoadEnv{
	// 	InstallLocation: loadEnvReq.InstallLocation,
	// 	NsId:            loadEnvReq.NsId,
	// 	McisId:          loadEnvReq.McisId,
	// 	VmId:            loadEnvReq.VmId,
	// 	Username:        loadEnvReq.Username,
	// 	PublicIp:        loadEnvReq.PublicIp,
	// 	PemKeyPath:      loadEnvReq.PemKeyPath,
	// }

	// createdEnvId, err := repository.SaveLoadTestInstallEnv(&loadEnv)
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to save load test installation environment: %w", err)
	// }
	// log.Printf("Environment ID %d for load test is successfully created", createdEnvId)

	// return createdEnvId, nil

	return nil
}
