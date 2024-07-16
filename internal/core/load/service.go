package load

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"gorm.io/gorm"
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
	ID        uint      `json:"id,omitempty"`
	NsId      string    `json:"nsId,omitempty"`
	McisId    string    `json:"mcisId,omitempty"`
	VmId      string    `json:"vmId,omitempty"`
	VmCount   int       `json:"vmCount,omitempty"`
	Status    string    `json:"status,omitempty"`
	Username  string    `json:"username,omitempty"`
	AgentType string    `json:"agentType,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// InstallMonitoringAgent installs a monitoring agent on specified VMs or all VM on Mcis.
func (l *LoadService) InstallMonitoringAgent(param MonitoringAgentInstallationParams) ([]MonitoringAgentInstallationResult, error) {
	utils.LogInfo("Starting installation of monitoring agent...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var res []MonitoringAgentInstallationResult

	scriptPath := utils.JoinRootPathWith("/script/install-server-agent.sh")
	utils.LogInfof("Reading installation script from %s", scriptPath)
	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		utils.LogErrorf("Failed to read installation script: %v", err)
		return res, err
	}
	username := "cb-user"

	utils.LogInfof("Fetching Mcis object for NS: %s, MCIS: %s", param.NsId, param.McisId)
	mcis, err := l.tumblebugClient.GetMcisWithContext(ctx, param.NsId, param.McisId)
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
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
	MonitoringAgentInfos []MonitoringAgentInstallationResult `json:"monitoringAgentInfos,omitempty"`
	TotalRow             int64                               `json:"totalRow,omitempty"`
}

func (l *LoadService) GetAllMonitoringAgentInfos(param GetAllMonitoringAgentInfosParam) (GetAllMonitoringAgentInfoResult, error) {
	var res GetAllMonitoringAgentInfoResult
	var monitoringAgentInfos []MonitoringAgentInstallationResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

type InstallLoadGeneratorParam struct {
	InstallLocation constant.InstallLocation `json:"installLocation,omitempty"`
	Coordinates     []string                 `json:"coordinate"`
}

type LoadGeneratorServerResult struct {
	ID              uint      `json:"id,omitempty"`
	Csp             string    `json:"csp,omitempty"`
	Region          string    `json:"region,omitempty"`
	Zone            string    `json:"zone,omitempty"`
	PublicIp        string    `json:"publicIp,omitempty"`
	PrivateIp       string    `json:"privateIp,omitempty"`
	PublicDns       string    `json:"publicDns,omitempty"`
	MachineType     string    `json:"machineType,omitempty"`
	Status          string    `json:"status,omitempty"`
	SshPort         string    `json:"sshPort,omitempty"`
	Lat             string    `json:"lat,omitempty"`
	Lon             string    `json:"lon,omitempty"`
	Username        string    `json:"username,omitempty"`
	VmId            string    `json:"vmId,omitempty"`
	StartTime       time.Time `json:"startTime,omitempty"`
	AdditionalVmKey string    `json:"additionalVmKey,omitempty"`
	Label           string    `json:"label,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}

type LoadGeneratorInstallInfoResult struct {
	ID              uint                     `json:"id,omitempty"`
	InstallLocation constant.InstallLocation `json:"installLocation,omitempty"`
	InstallType     string                   `json:"installType,omitempty"`
	InstallPath     string                   `json:"installPath,omitempty"`
	InstallVersion  string                   `json:"installVersion,omitempty"`
	Status          string                   `json:"status,omitempty"`
	CreatedAt       time.Time                `json:"createdAt,omitempty"`
	UpdatedAt       time.Time                `json:"updatedAt,omitempty"`

	PublicKeyName        string                      `json:"publicKeyName,omitempty"`
	PrivateKeyName       string                      `json:"privateKeyName,omitempty"`
	LoadGeneratorServers []LoadGeneratorServerResult `json:"loadGeneratorServers,omitempty"`
}

const (
	antNsId            = "ant-default-ns"
	antMcisDescription = "Default MCIS for Cloud Migration Verification"
	antInstallMonAgent = "no"
	antMcisLabel       = "DynamicMcis,AntDefault"
	antMcisId          = "ant-default-mcis"

	antVmDescription  = "Default VM for Cloud Migration Verification"
	antVmLabel        = "DynamicVm,AntDefault"
	antVmName         = "ant-default-vm"
	antVmRootDiskSize = "default"
	antVmRootDiskType = "default"
	antVmSubGroupSize = "1"
	antVmUserPassword = ""

	antPubKeyName  = "id_rsa_ant.pub"
	antPrivKeyName = "id_rsa_ant"

	defaultDelay = 20 * time.Second
	imageOs      = "ubuntu22.04"
)

// InstallLoadGenerator installs the load generator either locally or remotely.
// Currently remote request is executing via cb-tumblebug.
func (l *LoadService) InstallLoadGenerator(param InstallLoadGeneratorParam) (LoadGeneratorInstallInfoResult, error) {
	utils.LogInfo("Starting InstallLoadGenerator")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var result LoadGeneratorInstallInfoResult

	loadGeneratorInstallInfo := LoadGeneratorInstallInfo{
		InstallLocation: param.InstallLocation,
		InstallType:     "jmeter",
		InstallPath:     config.AppConfig.Load.JMeter.Dir,
		InstallVersion:  config.AppConfig.Load.JMeter.Version,
		Status:          "starting",
	}

	err := l.loadRepo.InsertLoadGeneratorInstallInfoTx(ctx, &loadGeneratorInstallInfo)
	if err != nil {
		utils.LogError("Failed to insert LoadGeneratorInstallInfo:", err)
		return result, err
	}

	utils.LogInfo("LoadGeneratorInstallInfo inserted successfully")

	installLocation := param.InstallLocation
	installScriptPath := utils.JoinRootPathWith("/script/install-jmeter.sh")

	switch installLocation {
	case constant.Local:
		utils.LogInfo("Starting local installation of JMeter")
		err := utils.Script(installScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", config.AppConfig.Load.JMeter.Dir),
			fmt.Sprintf("JMETER_VERSION=%s", config.AppConfig.Load.JMeter.Version),
		})
		if err != nil {
			utils.LogError("Error while installing JMeter locally:", err)
			return result, fmt.Errorf("error while installing jmeter; %s", err)
		}
		utils.LogInfo("Local installation of JMeter completed successfully")
	case constant.Remote:
		utils.LogInfo("Starting remote installation of JMeter")
		// get the spec and image information
		recommendVm, err := l.getRecommendVm(ctx, param.Coordinates)
		if err != nil {
			utils.LogError("Failed to get recommended VM:", err)
			return result, err
		}
		antVmCommonSpec := recommendVm[0].Name
		antVmConnectionName := recommendVm[0].ConnectionName
		antVmCommonImage, err := utils.ReplaceAtIndex(antVmCommonSpec, imageOs, "+", 2)

		if err != nil {
			utils.LogError("Error replacing VM spec index:", err)
			return result, err
		}

		// check namespace is valid or not
		err = l.validDefaultNs(ctx, antNsId)
		if err != nil {
			utils.LogError("Error validating default namespace:", err)
			return result, err
		}

		// get the ant default mcis
		antMcis, err := l.getAndDefaultMcis(ctx, antVmCommonSpec, antVmCommonImage, antVmConnectionName)
		if err != nil {
			utils.LogError("Error getting or creating default MCIS:", err)
			return result, err
		}

		// if server is not running state, try to resume and get mcis information
		retryCount := config.AppConfig.Load.Retry
		for retryCount > 0 && antMcis.StatusCount.CountRunning < 1 {
			utils.LogInfof("Attempting to resume MCIS, retry count: %d", retryCount)

			err = l.tumblebugClient.ControlLifecycleWithContext(ctx, antNsId, antMcis.ID, "resume")
			if err != nil {
				utils.LogError("Error resuming MCIS:", err)
				return result, err
			}
			time.Sleep(defaultDelay)
			antMcis, err = l.getAndDefaultMcis(ctx, antVmCommonSpec, antVmCommonImage, antVmConnectionName)
			if err != nil {
				utils.LogError("Error getting MCIS after resume attempt:", err)
				return result, err
			}

			retryCount = retryCount - 1
		}

		if antMcis.StatusCount.CountRunning < 1 {
			utils.LogError("No running VM on ant default MCIS")
			return result, errors.New("there is no running vm on ant default mcis")
		}

		addAuthorizedKeyCommand, err := getAddAuthorizedKeyCommand()
		if err != nil {
			utils.LogError("Error getting add authorized key command:", err)
			return result, err
		}

		installationCommand, err := utils.ReadToString(installScriptPath)
		if err != nil {
			utils.LogError("Error reading installation script:", err)
			return result, err
		}

		commandReq := tumblebug.SendCommandReq{
			Command: []string{installationCommand, addAuthorizedKeyCommand},
		}

		_, err = l.tumblebugClient.CommandToMcisWithContext(ctx, antNsId, antMcis.ID, commandReq)
		if err != nil {
			utils.LogError("Error sending command to MCIS:", err)
			return result, err
		}

		utils.LogInfo("Commands sent to MCIS successfully")

		loadGeneratorServers := make([]LoadGeneratorServer, 0)

		for i, vm := range antMcis.VMs {
			loadGeneratorServer := LoadGeneratorServer{
				Csp:             vm.ConnectionConfig.ProviderName,
				Region:          vm.Region.Region,
				Zone:            vm.Region.Zone,
				PublicIp:        vm.PublicIP,
				PrivateIp:       vm.PrivateIP,
				PublicDns:       vm.PublicDNS,
				MachineType:     vm.CspViewVMDetail.VMSpecName,
				Status:          vm.Status,
				SshPort:         vm.SSHPort,
				Lat:             fmt.Sprintf("%f", vm.Location.Latitude),
				Lon:             fmt.Sprintf("%f", vm.Location.Longitude),
				Username:        vm.CspViewVMDetail.VMUserID,
				VmId:            vm.CspViewVMDetail.IID.SystemID,
				StartTime:       vm.CspViewVMDetail.StartTime,
				AdditionalVmKey: vm.ID,
				Label:           vm.Label,
				IsCluster:       false,
				IsMaster:        i == 0,
				ClusterSize:     uint64(len(antMcis.VMs)),
			}

			loadGeneratorServers = append(loadGeneratorServers, loadGeneratorServer)
		}

		loadGeneratorInstallInfo.LoadGeneratorServers = loadGeneratorServers
		loadGeneratorInstallInfo.PublicKeyName = antPubKeyName
		loadGeneratorInstallInfo.PrivateKeyName = antPrivKeyName
	}

	loadGeneratorInstallInfo.Status = "installed"
	err = l.loadRepo.UpdateLoadGeneratorInstallInfoTx(ctx, &loadGeneratorInstallInfo)
	if err != nil {
		utils.LogError("Error updating LoadGeneratorInstallInfo:", err)
		return result, err
	}

	utils.LogInfo("LoadGeneratorInstallInfo updated successfully")

	loadGeneratorServerResults := make([]LoadGeneratorServerResult, 0)
	for _, l := range loadGeneratorInstallInfo.LoadGeneratorServers {
		lr := LoadGeneratorServerResult{
			ID:              l.ID,
			Csp:             l.Csp,
			Region:          l.Region,
			Zone:            l.Zone,
			PublicIp:        l.PublicIp,
			PrivateIp:       l.PrivateIp,
			PublicDns:       l.PublicDns,
			MachineType:     l.MachineType,
			Status:          l.Status,
			SshPort:         l.SshPort,
			Lat:             l.Lat,
			Lon:             l.Lon,
			Username:        l.Username,
			VmId:            l.VmId,
			StartTime:       l.StartTime,
			AdditionalVmKey: l.AdditionalVmKey,
			Label:           l.Label,
			CreatedAt:       l.CreatedAt,
			UpdatedAt:       l.UpdatedAt,
		}
		loadGeneratorServerResults = append(loadGeneratorServerResults, lr)
	}

	result.ID = loadGeneratorInstallInfo.ID
	result.InstallLocation = loadGeneratorInstallInfo.InstallLocation
	result.InstallType = loadGeneratorInstallInfo.InstallType
	result.InstallPath = loadGeneratorInstallInfo.InstallPath
	result.InstallVersion = loadGeneratorInstallInfo.InstallVersion
	result.Status = loadGeneratorInstallInfo.Status
	result.PublicKeyName = loadGeneratorInstallInfo.PublicKeyName
	result.PrivateKeyName = loadGeneratorInstallInfo.PrivateKeyName
	result.CreatedAt = loadGeneratorInstallInfo.CreatedAt
	result.UpdatedAt = loadGeneratorInstallInfo.UpdatedAt
	result.LoadGeneratorServers = loadGeneratorServerResults

	utils.LogInfo("InstallLoadGenerator completed successfully")

	return result, nil
}

// getAddAuthorizedKeyCommand returns a command to add the authorized key.
func getAddAuthorizedKeyCommand() (string, error) {
	pubKeyPath, _, err := validateKeyPair()
	if err != nil {
		return "", err
	}

	pub, err := utils.ReadToString(pubKeyPath)
	if err != nil {
		return "", err
	}

	addAuthorizedKeyScript := utils.JoinRootPathWith("/script/add-authorized-key.sh")

	addAuthorizedKeyCommand, err := utils.ReadToString(addAuthorizedKeyScript)
	if err != nil {
		return "", err
	}

	addAuthorizedKeyCommand = strings.Replace(addAuthorizedKeyCommand, `PUBLIC_KEY=""`, fmt.Sprintf(`PUBLIC_KEY="%s"`, pub), 1)
	return addAuthorizedKeyCommand, nil
}

// validateKeyPair checks and generates SSH key pair if it doesn't exist.
func validateKeyPair() (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	privKeyPath := fmt.Sprintf("%s/.ssh/%s", homeDir, antPrivKeyName)
	pubKeyPath := fmt.Sprintf("%s/.ssh/%s", homeDir, antPubKeyName)

	err = utils.CreateFolderIfNotExist(fmt.Sprintf("%s/.ssh", homeDir))
	if err != nil {
		return pubKeyPath, privKeyPath, err
	}

	exist := utils.ExistCheck(privKeyPath)
	if !exist {
		err := utils.GenerateSSHKeyPair(4096, privKeyPath, pubKeyPath)
		if err != nil {
			return pubKeyPath, privKeyPath, err
		}
	}
	return pubKeyPath, privKeyPath, nil
}

// getAndDefaultMcis retrieves or creates the default MCIS.
func (l *LoadService) getAndDefaultMcis(ctx context.Context, antVmCommonSpec, antVmCommonImage, antVmConnectionName string) (tumblebug.McisRes, error) {
	var antMcis tumblebug.McisRes
	var err error
	antMcis, err = l.tumblebugClient.GetMcisWithContext(ctx, antNsId, antMcisId)
	if err != nil {
		if errors.Is(err, tumblebug.ErrNotFound) {
			dynamicMcisArg := tumblebug.DynamicMcisReq{
				Description:     antMcisDescription,
				InstallMonAgent: antInstallMonAgent,
				Label:           antMcisLabel,
				Name:            antMcisId,
				SystemLabel:     "",
				VM: []tumblebug.DynamicVmReq{
					{
						CommonImage:    antVmCommonImage,
						CommonSpec:     antVmCommonSpec,
						ConnectionName: antVmConnectionName,
						Description:    antVmDescription,
						Label:          antVmLabel,
						Name:           antVmName,
						RootDiskSize:   antVmRootDiskSize,
						RootDiskType:   antVmRootDiskType,
						SubGroupSize:   antVmSubGroupSize,
						VMUserPassword: antVmUserPassword,
					},
				},
			}
			antMcis, err = l.tumblebugClient.DynamicMcisWithContext(ctx, antNsId, dynamicMcisArg)
			time.Sleep(defaultDelay)
			if err != nil {
				return antMcis, err
			}
		} else {
			return antMcis, err
		}
	} else if antMcis.VMs != nil && len(antMcis.VMs) == 0 {

		dynamicVmArg := tumblebug.DynamicVmReq{
			CommonImage:    antVmCommonImage,
			CommonSpec:     antVmCommonSpec,
			ConnectionName: antVmConnectionName,
			Description:    antVmDescription,
			Label:          antVmLabel,
			Name:           antVmName,
			RootDiskSize:   antVmRootDiskSize,
			RootDiskType:   antVmRootDiskType,
			SubGroupSize:   antVmSubGroupSize,
			VMUserPassword: antVmUserPassword,
		}

		antMcis, err = l.tumblebugClient.DynamicVmWithContext(ctx, antNsId, antMcisId, dynamicVmArg)
		time.Sleep(defaultDelay)
		if err != nil {
			return antMcis, err
		}
	}
	return antMcis, nil
}

// getRecommendVm retrieves recommendVm to specify the location of provisioning.
func (l *LoadService) getRecommendVm(ctx context.Context, coordinates []string) (tumblebug.RecommendVmResList, error) {
	recommendVmArg := tumblebug.RecommendVmReq{
		Filter: tumblebug.Filter{
			Policy: []tumblebug.FilterPolicy{
				{
					Condition: []tumblebug.Condition{
						{
							Operand:  "2",
							Operator: ">=",
						},
						{
							Operand:  "8",
							Operator: "<=",
						},
					},
					Metric: "vCPU",
				},
				{
					Condition: []tumblebug.Condition{
						{
							Operand:  "4",
							Operator: ">=",
						},
						{
							Operand:  "8",
							Operator: "<=",
						},
					},
					Metric: "memoryGiB",
				},
				{
					Condition: []tumblebug.Condition{
						{
							Operand: "aws",
						},
					},
					Metric: "providerName",
				},
			},
		},
		Limit: "1",
		Priority: tumblebug.Priority{
			Policy: []tumblebug.Policy{
				{
					Metric: "location",
					Parameter: []tumblebug.Parameter{
						{
							Key: "coordinateClose",
							Val: coordinates,
						},
					},
				},
			},
		},
	}

	recommendRes, err := l.tumblebugClient.GetRecommendVmWithContext(ctx, recommendVmArg)

	if err != nil {
		return nil, err
	}

	if len(recommendRes) == 0 {
		return nil, errors.New("there is no recommended vm list")
	}

	return recommendRes, nil
}

// validDefaultNs checks if the default namespace exists, and creates it if not.
func (l *LoadService) validDefaultNs(ctx context.Context, antNsId string) error {
	_, err := l.tumblebugClient.GetNsWithContext(ctx, antNsId)
	if err != nil && errors.Is(err, tumblebug.ErrNotFound) {

		arg := tumblebug.CreateNsReq{
			Name:        antNsId,
			Description: "cm-ant default ns for validating migration",
		}

		err = l.tumblebugClient.CreateNsWithContext(ctx, arg)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

type UninstallLoadGeneratorParam struct {
	LoadGeneratorInstallInfoId uint
}

func (l *LoadService) UninstallLoadGenerator(param UninstallLoadGeneratorParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	loadGeneratorInstallInfo, err := l.loadRepo.GetValidLoadGeneratorInstallInfoByIdTx(ctx, param.LoadGeneratorInstallInfoId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.LogError("Cannot find valid load generator install info:", err)
			return errors.New("cannot find valid load generator install info")
		}
		utils.LogError("Error retrieving load generator install info:", err)
		return err
	}

	uninstallScriptPath := utils.JoinRootPathWith("/script/uninstall-jmeter.sh")

	switch loadGeneratorInstallInfo.InstallLocation {
	case constant.Local:
		err := utils.Script(uninstallScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", config.AppConfig.Load.JMeter.Dir),
			fmt.Sprintf("JMETER_VERSION=%s", config.AppConfig.Load.JMeter.Version),
		})
		if err != nil {
			utils.LogErrorf("Error while uninstalling load generator: %s", err)
			return fmt.Errorf("error while uninstalling load generator: %s", err)
		}
	case constant.Remote:

		uninstallCommand, err := utils.ReadToString(uninstallScriptPath)
		if err != nil {
			utils.LogError("Error reading uninstall script:", err)
			return err
		}

		commandReq := tumblebug.SendCommandReq{
			Command: []string{uninstallCommand},
		}

		_, err = l.tumblebugClient.CommandToMcisWithContext(ctx, antNsId, antMcisId, commandReq)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				utils.LogError("VM is not in running state. Cannot connect to the VMs.")
				return errors.New("vm is not running state. cannot connect to the vms")
			}
			utils.LogError("Error sending uninstall command to MCIS:", err)
			return err
		}

		// err = l.tumblebugClient.ControlLifecycleWithContext(ctx, antNsId, antMcisId, "suspend")
		// if err != nil {
		// 	return err
		// }
	}

	loadGeneratorInstallInfo.Status = "deleted"
	for i := range loadGeneratorInstallInfo.LoadGeneratorServers {
		loadGeneratorInstallInfo.LoadGeneratorServers[i].Status = "deleted"
	}

	err = l.loadRepo.UpdateLoadGeneratorInstallInfoTx(ctx, &loadGeneratorInstallInfo)
	if err != nil {
		utils.LogError("Error updating load generator install info:", err)
		return err
	}

	utils.LogInfo("Successfully uninstalled load generator.")
	return nil
}

type GetAllLoadGeneratorInstallInfoParam struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	Status string `json:"Status"`
}

type GetAllLoadGeneratorInstallInfoResult struct {
	LoadGeneratorInstallInfoResults []LoadGeneratorInstallInfoResult `json:"loadGeneratorInstallInfoResults,omitempty"`
	TotalRows                       int64                            `json:"totalRows,omitempty"`
}

func (l *LoadService) GetAllLoadGeneratorInstallInfo(param GetAllLoadGeneratorInstallInfoParam) (GetAllLoadGeneratorInstallInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result GetAllLoadGeneratorInstallInfoResult
	var infos []LoadGeneratorInstallInfoResult
	pagedResult, totalRows, err := l.loadRepo.GetPagingLoadGeneratorInstallInfosTx(ctx, param)

	if err != nil {
		utils.LogError("Error fetching paged load generator install infos:", err)
		return result, err
	}

	for _, l := range pagedResult {
		loadGeneratorServerResults := make([]LoadGeneratorServerResult, 0)
		for _, s := range l.LoadGeneratorServers {
			lsr := LoadGeneratorServerResult{
				ID:              s.ID,
				Csp:             s.Csp,
				Region:          s.Region,
				Zone:            s.Zone,
				PublicIp:        s.PublicIp,
				PrivateIp:       s.PrivateIp,
				PublicDns:       s.PublicDns,
				MachineType:     s.MachineType,
				Status:          s.Status,
				SshPort:         s.SshPort,
				Lat:             s.Lat,
				Lon:             s.Lon,
				Username:        s.Username,
				VmId:            s.VmId,
				StartTime:       s.StartTime,
				AdditionalVmKey: s.AdditionalVmKey,
				Label:           s.Label,
				CreatedAt:       s.CreatedAt,
				UpdatedAt:       s.UpdatedAt,
			}
			loadGeneratorServerResults = append(loadGeneratorServerResults, lsr)
		}
		lr := LoadGeneratorInstallInfoResult{
			ID:                   l.ID,
			InstallLocation:      l.InstallLocation,
			InstallType:          l.InstallType,
			InstallPath:          l.InstallPath,
			InstallVersion:       l.InstallVersion,
			Status:               l.Status,
			PublicKeyName:        l.PublicKeyName,
			PrivateKeyName:       l.PrivateKeyName,
			CreatedAt:            l.CreatedAt,
			UpdatedAt:            l.UpdatedAt,
			LoadGeneratorServers: loadGeneratorServerResults,
		}

		infos = append(infos, lr)
	}

	result.LoadGeneratorInstallInfoResults = infos
	result.TotalRows = totalRows
	utils.LogInfof("Fetched %d load generator install info results.", len(infos))

	return result, nil
}

type RunLoadTestParam struct {
	LoadTestKey                string                    `json:"loadTestKey"`
	InstallLoadGenerator       InstallLoadGeneratorParam `json:"installLoadGenerator"`
	LoadGeneratorInstallInfoId uint                      `json:"loadGeneratorInstallInfoId"`
	TestName                   string                    `json:"testName"`
	VirtualUsers               string                    `json:"virtualUsers"`
	Duration                   string                    `json:"duration"`
	RampUpTime                 string                    `json:"rampUpTime"`
	RampUpSteps                string                    `json:"rampUpSteps"`
	Hostname                   string                    `json:"hostname"`
	Port                       string                    `json:"port"`
	AgentInstalled             bool                      `json:"agentInstalled"`
	AgentHostname              string                    `json:"agentHostname"`

	HttpReqs []RunLoadTestHttpParam `json:"httpReqs,omitempty"`
}

type RunLoadTestHttpParam struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`
}

// RunLoadTest initiates the load test and performs necessary initializations.
// Generates a load test key, installs the load generator or retrieves existing installation information,
// saves the load test execution state, and then asynchronously runs the load test.
func (l *LoadService) RunLoadTest(param RunLoadTestParam) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
	param.LoadTestKey = loadTestKey

	utils.LogInfof("Starting load test with key: %s", loadTestKey)

	if param.LoadGeneratorInstallInfoId == uint(0) {
		utils.LogInfo("No LoadGeneratorInstallInfoId provided, installing load generator...")
		result, err := l.InstallLoadGenerator(param.InstallLoadGenerator)
		if err != nil {
			utils.LogErrorf("Error installing load generator: %v", err)
			return "", err
		}

		param.LoadGeneratorInstallInfoId = result.ID
		utils.LogInfof("Load generator installed with ID: %d", result.ID)
	}

	if param.LoadGeneratorInstallInfoId == uint(0) {
		utils.LogErrorf("LoadGeneratorInstallInfoId is still 0 after installation.")
		return "", nil
	}

	utils.LogInfof("Retrieving load generator installation info with ID: %d", param.LoadGeneratorInstallInfoId)
	loadGeneratorInstallInfo, err := l.loadRepo.GetValidLoadGeneratorInstallInfoByIdTx(ctx, param.LoadGeneratorInstallInfoId)
	if err != nil {
		utils.LogErrorf("Error retrieving load generator installation info: %v", err)
		return "", err
	}

	duration, err := strconv.Atoi(param.Duration)
	if err != nil {
		return "", err
	}

	rampUpTime, err := strconv.Atoi(param.RampUpTime)

	if err != nil {
		return "", err
	}

	stateArg := LoadTestExecutionState{
		LoadGeneratorInstallInfoId:  loadGeneratorInstallInfo.ID,
		LoadTestKey:                 loadTestKey,
		ExecutionStatus:             constant.OnPreparing,
		StartAt:                     time.Now(),
		TotalExpectedExcutionSecond: uint64(duration + rampUpTime),
	}

	go l.runLoadTest(param, &loadGeneratorInstallInfo, &stateArg)

	var hs []LoadTestExecutionHttpInfo

	for _, h := range param.HttpReqs {
		hh := LoadTestExecutionHttpInfo{
			Method:   h.Method,
			Protocol: h.Protocol,
			Hostname: h.Hostname,
			Port:     h.Port,
			Path:     h.Path,
			BodyData: h.BodyData,
		}

		hs = append(hs, hh)
	}

	loadArg := LoadTestExecutionInfo{
		LoadTestKey:                loadTestKey,
		TestName:                   param.TestName,
		VirtualUsers:               param.VirtualUsers,
		Duration:                   param.Duration,
		RampUpTime:                 param.RampUpTime,
		RampUpSteps:                param.RampUpSteps,
		Hostname:                   param.Hostname,
		Port:                       param.Port,
		AgentInstalled:             param.AgentInstalled,
		AgentHostname:              param.AgentHostname,
		LoadGeneratorInstallInfoId: loadGeneratorInstallInfo.ID,
		LoadTestExecutionHttpInfos: hs,
	}

	utils.LogInfof("Saving load test execution info for key: %s", loadTestKey)
	err = l.loadRepo.SaveForLoadTestExecutionTx(ctx, &loadArg, &stateArg)
	if err != nil {
		utils.LogErrorf("Error saving load test execution info: %v", err)
		return "", err
	}

	utils.LogInfof("Load test started successfully with key: %s", loadTestKey)

	return loadTestKey, nil

}

// runLoadTest executes the load test.
// Depending on whether the installation location is local or remote, it creates the test plan and runs test commands.
// Fetches and saves test results from the local or remote system.
func (l *LoadService) runLoadTest(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo, loadTestExecutionState *LoadTestExecutionState) {
	defer func() {
		updateErr := l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
		if updateErr != nil {
			utils.LogErrorf("Error updating load test execution state: %v", updateErr)
			return
		}
	}()

	compileDuration, executionDuration, loadTestErr := l.executeLoadTest(param, loadGeneratorInstallInfo)

	loadTestExecutionState.CompileDuration = compileDuration
	loadTestExecutionState.ExecutionDuration = executionDuration

	if loadTestErr != nil {
		loadTestExecutionState.ExecutionStatus = constant.TestFailed
		loadTestExecutionState.FailureMessage = loadTestErr.Error()
		finishAt := time.Now()
		loadTestExecutionState.FinishAt = &finishAt
		return
	}

	updateErr := l.updateLoadTestExecution(loadTestExecutionState)
	if updateErr != nil {
		loadTestExecutionState.ExecutionStatus = constant.UpdateFailed
		loadTestExecutionState.FailureMessage = updateErr.Error()
		finishAt := time.Now()
		loadTestExecutionState.FinishAt = &finishAt
		return
	}

	resultFetchErr := l.fetchResultFile(param, loadGeneratorInstallInfo)

	if resultFetchErr != nil {
		loadTestExecutionState.ExecutionStatus = constant.ResultFailed
		loadTestExecutionState.FailureMessage = resultFetchErr.Error()
		finishAt := time.Now()
		loadTestExecutionState.FinishAt = &finishAt
		return
	}

	loadTestExecutionState.ExecutionStatus = constant.Successed

}

// executeLoadTest
func (l *LoadService) executeLoadTest(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo) (string, string, error) {
	installLocation := loadGeneratorInstallInfo.InstallLocation
	loadTestKey := param.LoadTestKey
	loadGeneratorInstallPath := loadGeneratorInstallInfo.InstallPath
	testPlanName := fmt.Sprintf("%s.jmx", loadTestKey)
	resultFileName := fmt.Sprintf("%s_result.csv", loadTestKey)
	loadGeneratorInstallVersion := loadGeneratorInstallInfo.InstallVersion

	utils.LogInfof("Running load test with key: %s", loadTestKey)
	compileDuration := "0"
	executionDuration := "0"
	start := time.Now()

	if installLocation == constant.Remote {
		utils.LogInfo("Remote execute detected.")
		var buf bytes.Buffer
		err := parseTestPlanStructToString(&buf, param, loadGeneratorInstallInfo)
		if err != nil {
			return compileDuration, executionDuration, err
		}

		testPlan := buf.String()

		createFileCmd := fmt.Sprintf("cat << 'EOF' > %s/test_plan/%s \n%s\nEOF", loadGeneratorInstallPath, testPlanName, testPlan)

		commandReq := tumblebug.SendCommandReq{
			Command: []string{createFileCmd},
		}

		compileDuration = utils.DurationString(start)
		_, err = l.tumblebugClient.CommandToMcisWithContext(context.Background(), antNsId, antMcisId, commandReq)
		if err != nil {
			return compileDuration, executionDuration, err
		}

		jmeterTestCommand := generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName)

		commandReq = tumblebug.SendCommandReq{
			Command: []string{jmeterTestCommand},
		}

		stdout, err := l.tumblebugClient.CommandToMcisWithContext(context.Background(), antNsId, antMcisId, commandReq)
		if err != nil {
			return compileDuration, executionDuration, err
		}
		executionDuration = utils.DurationString(start)

		if strings.Contains(stdout, "exited with status 1") {
			return compileDuration, executionDuration, errors.New("jmeter test stopped unexpectedly")
		}

	} else if installLocation == constant.Local {
		utils.LogInfo("Local execute detected.")

		exist := utils.ExistCheck(loadGeneratorInstallPath)

		if !exist {
			return compileDuration, executionDuration, errors.New("load generator installaion is not validated")
		}

		outputFile, err := os.Create(fmt.Sprintf("%s/test_plan/%s.jmx", loadGeneratorInstallPath, loadTestKey))
		if err != nil {
			return compileDuration, executionDuration, err
		}

		err = parseTestPlanStructToString(outputFile, param, loadGeneratorInstallInfo)

		if err != nil {
			return compileDuration, executionDuration, err
		}

		jmeterTestCommand := generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName)
		compileDuration = utils.DurationString(start)

		err = utils.InlineCmd(jmeterTestCommand)
		executionDuration = utils.DurationString(start)
		if err != nil {
			return compileDuration, executionDuration, fmt.Errorf("jmeter test stopped unexpectedly; %w", err)
		}
	}

	return compileDuration, executionDuration, nil
}

func (l *LoadService) updateLoadTestExecution(loadTestExecutionState *LoadTestExecutionState) error {
	err := l.loadRepo.UpdateLoadTestExecutionInfoDuration(context.Background(), loadTestExecutionState.LoadTestKey, loadTestExecutionState.CompileDuration, loadTestExecutionState.ExecutionDuration)
	if err != nil {
		utils.LogErrorf("Error updating load test execution info: %v", err)
		return err
	}

	loadTestExecutionState.ExecutionStatus = constant.OnFetching
	err = l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
	if err != nil {
		utils.LogErrorf("Error updating load test execution state: %v", err)
		return err
	}
	return nil
}

func (l *LoadService) fetchResultFile(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo) error {
	installLocation := loadGeneratorInstallInfo.InstallLocation
	loadTestKey := param.LoadTestKey
	loadGeneratorInstallPath := loadGeneratorInstallInfo.InstallPath
	utils.LogInfof("Fetching results for load test key: %s", loadTestKey)

	var wg sync.WaitGroup
	resultsPrefix := []string{""}

	if param.AgentInstalled {
		resultsPrefix = append(resultsPrefix, "_cpu", "_disk", "_memory", "_network")
	}

	errorChan := make(chan error, len(resultsPrefix))

	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)

	err := utils.CreateFolderIfNotExist(utils.JoinRootPathWith("/result"))
	if err != nil {
		return err
	}

	err = utils.CreateFolderIfNotExist(resultFolderPath)
	if err != nil {
		return err
	}

	if installLocation == constant.Local {
		for _, p := range resultsPrefix {
			wg.Add(1)
			go func(prefix string) {
				defer wg.Done()
				fileName := fmt.Sprintf("%s%s_result.csv", loadTestKey, prefix)
				fromFilePath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, fileName)
				toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

				if exist := utils.ExistCheck(toFilePath); !exist {
					err := utils.InlineCmd(fmt.Sprintf("cp %s %s", fromFilePath, toFilePath))
					if err != nil {
						errorChan <- err
						return
					}
				}
				errorChan <- nil
			}(p)
		}
	} else if installLocation == constant.Remote {
		var username string
		var publicIp string
		var port string
		for _, s := range loadGeneratorInstallInfo.LoadGeneratorServers {
			if s.IsMaster {
				username = s.Username
				publicIp = s.PublicIp
				port = s.SshPort
			}
		}

		client, err := utils.GetClient(publicIp, port, username, loadGeneratorInstallInfo.PrivateKeyName)

		if err != nil {
			return err
		}
		defer client.Close()

		for _, p := range resultsPrefix {
			wg.Add(1)
			go func(prefix string) {
				defer wg.Done()
				fileName := fmt.Sprintf("%s%s_result.csv", loadTestKey, prefix)
				resultFilePath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, fileName)
				toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

				if exist := utils.ExistCheck(toFilePath); !exist {
					err := utils.DownloadFile(client, toFilePath, resultFilePath)
					if err != nil {
						errorChan <- err
						return
					}
				}
				errorChan <- nil
			}(p)
		}
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// generateJmeterExecutionCmd generates the JMeter execution command.
// Constructs a JMeter command string that includes the test plan path and result file path.
func generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName string) string {
	utils.LogInfof("Generating JMeter execution command for test plan: %s, result file: %s", testPlanName, resultFileName)

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", loadGeneratorInstallPath, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", loadGeneratorInstallPath, loadGeneratorInstallVersion))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	builder.WriteString(fmt.Sprintf(" && sudo rm %s", testPath))
	utils.LogInfof("JMeter execution command generated: %s", builder.String())
	return builder.String()
}

type GetAllLoadTestExecutionStateParam struct {
	Page            int                      `json:"page"`
	Size            int                      `json:"size"`
	LoadTestKey     string                   `json:"loadTestKey"`
	ExecutionStatus constant.ExecutionStatus `json:"executionStatus"`
}

type GetAllLoadTestExecutionStateResult struct {
	LoadTestExecutionStates []LoadTestExecutionStateResult `json:"loadTestExecutionStates,omitempty"`
	TotalRow                int64                          `json:"totalRow,omitempty"`
}

type LoadTestExecutionStateResult struct {
	ID                          uint                           `json:"id"`
	LoadGeneratorInstallInfoId  uint                           `json:"loadGeneratorInstallInfoId,omitempty"`
	LoadGeneratorInstallInfo    LoadGeneratorInstallInfoResult `json:"loadGeneratorInstallInfo,omitempty"`
	LoadTestKey                 string                         `json:"loadTestKey,omitempty"`
	ExecutionStatus             constant.ExecutionStatus       `json:"executionStatus,omitempty"`
	StartAt                     time.Time                      `json:"startAt,omitempty"`
	FinishAt                    *time.Time                     `json:"finishAt,omitempty"`
	TotalExpectedExcutionSecond uint64                         `json:"totalExpectedExecutionSecond,omitempty"`
	FailureMessage              string                         `json:"failureMessage,omitempty"`
	CompileDuration             string                         `json:"compileDuration,omitempty"`
	ExecutionDuration           string                         `json:"executionDuration,omitempty"`
	CreatedAt                   time.Time                      `json:"createdAt,omitempty"`
	UpdatedAt                   time.Time                      `json:"updatedAt,omitempty"`
}

func (l *LoadService) GetAllLoadTestExecutionState(param GetAllLoadTestExecutionStateParam) (GetAllLoadTestExecutionStateResult, error) {
	var res GetAllLoadTestExecutionStateResult
	var states []LoadTestExecutionStateResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetAllLoadExecutionStates called with param: %+v", param)
	result, totalRows, err := l.loadRepo.GetPagingLoadTestExecutionStateTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution state infos: %v", err)
		return res, err
	}

	utils.LogInfof("Fetched %d monitoring agent infos", len(result))

	for _, loadTestExecutionState := range result {
		state := mapLoadTestExecutionStateResult(loadTestExecutionState)
		state.LoadGeneratorInstallInfo = mapLoadGeneratorInstallInfoResult(loadTestExecutionState.LoadGeneratorInstallInfo)
		states = append(states, state)
	}

	res.LoadTestExecutionStates = states
	res.TotalRow = totalRows

	return res, nil
}

type GetLoadTestExecutionStateParam struct {
	LoadTestKey string `json:"loadTestKey"`
}

func (l *LoadService) GetLoadTestExecutionState(param GetLoadTestExecutionStateParam) (LoadTestExecutionStateResult, error) {
	var res LoadTestExecutionStateResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetLoadTestExecutionState called with param: %+v", param)
	state, err := l.loadRepo.GetLoadTestExecutionStateTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution state infos: %v", err)
		return res, err
	}

	res = mapLoadTestExecutionStateResult(state)
	res.LoadGeneratorInstallInfo = mapLoadGeneratorInstallInfoResult(state.LoadGeneratorInstallInfo)
	return res, nil
}

type GetAllLoadTestExecutionInfosParam struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

type GetAllLoadTestExecutionInfosResult struct {
	TotalRow               int64                         `json:"totalRow,omitempty"`
	LoadTestExecutionInfos []LoadTestExecutionInfoResult `json:"loadTestExecutionInfos,omitempty"`
}

type LoadTestExecutionInfoResult struct {
	ID                         uint                              `json:"id"`
	LoadTestKey                string                            `json:"loadTestKey,omitempty"`
	TestName                   string                            `json:"testName,omitempty"`
	VirtualUsers               string                            `json:"virtualUsers,omitempty"`
	Duration                   string                            `json:"duration,omitempty"`
	RampUpTime                 string                            `json:"rampUpTime,omitempty"`
	RampUpSteps                string                            `json:"rampUpSteps,omitempty"`
	Hostname                   string                            `json:"hostname,omitempty"`
	Port                       string                            `json:"port,omitempty"`
	AgentHostname              string                            `json:"agentHostname,omitempty"`
	AgentInstalled             bool                              `json:"agentInstalled,omitempty"`
	CompileDuration            string                            `json:"compileDuration,omitempty"`
	ExecutionDuration          string                            `json:"executionDuration,omitempty"`
	LoadTestExecutionHttpInfos []LoadTestExecutionHttpInfoResult `json:"loadTestExecutionHttpInfos,omitempty"`
	LoadTestExecutionState     LoadTestExecutionStateResult      `json:"loadTestExecutionState,omitempty"`
	LoadGeneratorInstallInfo   LoadGeneratorInstallInfoResult    `json:"loadGeneratorInstallInfo,omitempty"`
}

type LoadTestExecutionHttpInfoResult struct {
	ID       uint   `json:"id"`
	Method   string `json:"method,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	BodyData string `json:"bodyData,omitempty"`
}

func (l *LoadService) GetAllLoadTestExecutionInfos(param GetAllLoadTestExecutionInfosParam) (GetAllLoadTestExecutionInfosResult, error) {
	var res GetAllLoadTestExecutionInfosResult
	var rs []LoadTestExecutionInfoResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetAllLoadTestExecutionInfos called with param: %+v", param)
	result, totalRows, err := l.loadRepo.GetPagingLoadTestExecutionHistoryTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution infos: %v", err)
		return res, err
	}

	utils.LogInfof("Fetched %d load test execution infos:", len(result))

	for _, r := range result {
		rs = append(rs, mapLoadTestExecutionInfoResult(r))
	}

	res.TotalRow = totalRows
	res.LoadTestExecutionInfos = rs

	return res, nil
}

type GetLoadTestExecutionInfoParam struct {
	LoadTestKey string `json:"loadTestKey"`
}

func (l *LoadService) GetLoadTestExecutionInfo(param GetLoadTestExecutionInfoParam) (LoadTestExecutionInfoResult, error) {
	var res LoadTestExecutionInfoResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetLoadTestExecutionInfo called with param: %+v", param)
	executionInfo, err := l.loadRepo.GetLoadTestExecutionInfoTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution state infos: %v", err)
		return res, err
	}

	return mapLoadTestExecutionInfoResult(executionInfo), nil
}

func mapLoadTestExecutionHttpInfoResult(h LoadTestExecutionHttpInfo) LoadTestExecutionHttpInfoResult {
	return LoadTestExecutionHttpInfoResult{
		ID:       h.ID,
		Method:   h.Method,
		Protocol: h.Protocol,
		Hostname: h.Hostname,
		Port:     h.Port,
		Path:     h.Path,
		BodyData: h.BodyData,
	}
}

func mapLoadTestExecutionStateResult(state LoadTestExecutionState) LoadTestExecutionStateResult {
	return LoadTestExecutionStateResult{
		ID:                          state.ID,
		LoadTestKey:                 state.LoadTestKey,
		ExecutionStatus:             state.ExecutionStatus,
		StartAt:                     state.StartAt,
		FinishAt:                    state.FinishAt,
		TotalExpectedExcutionSecond: state.TotalExpectedExcutionSecond,
		FailureMessage:              state.FailureMessage,
		CompileDuration:             state.CompileDuration,
		ExecutionDuration:           state.ExecutionDuration,
		CreatedAt:                   state.CreatedAt,
		UpdatedAt:                   state.UpdatedAt,
	}
}

func mapLoadGeneratorServerResult(s LoadGeneratorServer) LoadGeneratorServerResult {
	return LoadGeneratorServerResult{
		ID:              s.ID,
		Csp:             s.Csp,
		Region:          s.Region,
		Zone:            s.Zone,
		PublicIp:        s.PublicIp,
		PrivateIp:       s.PrivateIp,
		PublicDns:       s.PublicDns,
		MachineType:     s.MachineType,
		Status:          s.Status,
		SshPort:         s.SshPort,
		Lat:             s.Lat,
		Lon:             s.Lon,
		Username:        s.Username,
		VmId:            s.VmId,
		StartTime:       s.StartTime,
		AdditionalVmKey: s.AdditionalVmKey,
		Label:           s.Label,
		CreatedAt:       s.CreatedAt,
	}
}

func mapLoadGeneratorInstallInfoResult(install LoadGeneratorInstallInfo) LoadGeneratorInstallInfoResult {
	var servers []LoadGeneratorServerResult
	for _, s := range install.LoadGeneratorServers {
		servers = append(servers, mapLoadGeneratorServerResult(s))
	}

	return LoadGeneratorInstallInfoResult{
		ID:                   install.ID,
		InstallLocation:      install.InstallLocation,
		InstallType:          install.InstallType,
		InstallPath:          install.InstallPath,
		InstallVersion:       install.InstallVersion,
		Status:               install.Status,
		CreatedAt:            install.CreatedAt,
		UpdatedAt:            install.UpdatedAt,
		PublicKeyName:        install.PublicKeyName,
		PrivateKeyName:       install.PrivateKeyName,
		LoadGeneratorServers: servers,
	}
}

func mapLoadTestExecutionInfoResult(executionInfo LoadTestExecutionInfo) LoadTestExecutionInfoResult {
	var httpResults []LoadTestExecutionHttpInfoResult
	for _, h := range executionInfo.LoadTestExecutionHttpInfos {
		httpResults = append(httpResults, mapLoadTestExecutionHttpInfoResult(h))
	}

	executionState := mapLoadTestExecutionStateResult(executionInfo.LoadTestExecutionState)
	installInfo := mapLoadGeneratorInstallInfoResult(executionInfo.LoadGeneratorInstallInfo)

	return LoadTestExecutionInfoResult{
		ID:                         executionInfo.ID,
		LoadTestKey:                executionInfo.LoadTestKey,
		TestName:                   executionInfo.TestName,
		VirtualUsers:               executionInfo.VirtualUsers,
		Duration:                   executionInfo.Duration,
		RampUpTime:                 executionInfo.RampUpTime,
		RampUpSteps:                executionInfo.RampUpSteps,
		Hostname:                   executionInfo.Hostname,
		Port:                       executionInfo.Port,
		AgentHostname:              executionInfo.AgentHostname,
		AgentInstalled:             executionInfo.AgentInstalled,
		CompileDuration:            executionInfo.CompileDuration,
		ExecutionDuration:          executionInfo.ExecutionDuration,
		LoadTestExecutionHttpInfos: httpResults,
		LoadTestExecutionState:     executionState,
		LoadGeneratorInstallInfo:   installInfo,
	}
}

type StopLoadTestParam struct {
	LoadTestKey string `json:"loadTestKey"`
}

func (l *LoadService) StopLoadTest(param StopLoadTestParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	state, err := l.loadRepo.GetLoadTestExecutionStateTx(ctx, GetLoadTestExecutionStateParam{
		LoadTestKey: param.LoadTestKey,
	})

	if err != nil {
		return err
	}

	if state.ExecutionStatus == constant.Successed {
		return nil
	}

	installInfo := state.LoadGeneratorInstallInfo

	killCmd := killCmdGen(param.LoadTestKey)

	if installInfo.InstallLocation == constant.Remote {

		commandReq := tumblebug.SendCommandReq{
			Command: []string{killCmd},
		}
		_, err := l.tumblebugClient.CommandToMcisWithContext(ctx, antNsId, antMcisId, commandReq)

		if err != nil {
			return err
		}

	} else if installInfo.InstallLocation == constant.Local {
		err := utils.InlineCmd(killCmd)

		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil

}

func killCmdGen(loadTestKey string) string {
	grepRegex := fmt.Sprintf("'\\/bin\\/ApacheJMeter\\.jar.*%s'", loadTestKey)
	utils.LogInfof("Generating kill command for load test key: %s", loadTestKey)
	return fmt.Sprintf("kill -15 $(ps -ef | grep -E %s | awk '{print $2}')", grepRegex)
}
