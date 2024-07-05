package load

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var res []MonitoringAgentInstallationResult

	scriptPath := utils.JoinRootPathWith("/script/install-server-agent.sh")
	utils.LogInfof("Reading installation script from %s", scriptPath)
	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Printf("[ERROR] Failed to read installation script: %v", err)
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
	MonitoringAgentInfos []MonitoringAgentInstallationResult `json:"monitoringAgentInfos"`
	TotalRow             int64                               `json:"totalRow"`
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
	ID              uint
	Csp             string
	Region          string
	Zone            string
	PublicIp        string
	PrivateIp       string
	PublicDns       string
	MachineType     string
	Status          string
	SshPort         string
	Lat             string
	Lon             string
	Username        string
	VmId            string
	StartTime       time.Time
	AdditionalVmKey string
	Label           string
	CreatedAt       time.Time
}

type LoadGeneratorInstallInfoResult struct {
	ID              uint
	InstallLocation constant.InstallLocation
	InstallType     string
	InstallPath     string
	InstallVersion  string
	Status          string
	CreatedAt       time.Time

	PublicKeyName        string
	PrivateKeyName       string
	LoadGeneratorServers []LoadGeneratorServerResult
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
		imageOs := "ubuntu22.04"
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
			time.Sleep(10 * time.Second)
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

		for _, vm := range antMcis.VMs {
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
			time.Sleep(10 * time.Second)
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
		time.Sleep(10 * time.Second)
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
	LoadGeneratorInstallInfoId string
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

	log.Println(loadGeneratorInstallInfo)

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
	LoadGeneratorInstallInfoResults []LoadGeneratorInstallInfoResult
	TotalRows                       int64
}

func (l *LoadService) GetAllLoadGeneratorInstallInfo(param GetAllLoadGeneratorInstallInfoParam) (GetAllLoadGeneratorInstallInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
			LoadGeneratorServers: loadGeneratorServerResults,
		}

		infos = append(infos, lr)
	}

	result.LoadGeneratorInstallInfoResults = infos
	result.TotalRows = totalRows
	utils.LogInfof("Fetched %d load generator install info results.", len(infos))

	return result, nil
}
