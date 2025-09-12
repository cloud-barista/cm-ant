package load

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	antNsId            = "ant-default-ns"
	antMciDescription  = "Default MCI for Cloud Migration Verification"
	antInstallMonAgent = "no"
	antLabelKey        = "ant-default-label"
	antMciLabel        = "DynamicMci,AntDefault"
	antMciId           = "ant-default-mci"

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
)

// InstallLoadGenerator installs the load generator either locally or remotely.
// Currently remote request is executing via cb-tumblebug.
func (l *LoadService) InstallLoadGenerator(param InstallLoadGeneratorParam) (LoadGeneratorInstallInfoResult, error) {
	log.Info().Msg("Starting InstallLoadGenerator")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var result LoadGeneratorInstallInfoResult

	loadGeneratorInstallInfo := &LoadGeneratorInstallInfo{
		InstallLocation: param.InstallLocation,
		InstallType:     "jmeter",
		InstallPath:     config.AppConfig.Load.JMeter.Dir,
		InstallVersion:  config.AppConfig.Load.JMeter.Version,
		Status:          "starting",
	}

	err := l.loadRepo.GetOrInsertLoadGeneratorInstallInfoTx(ctx, loadGeneratorInstallInfo)
	if err != nil {
		log.Error().Msgf("Failed to insert LoadGeneratorInstallInfo; %v", err)
		return result, err
	}
	log.Info().Msg("LoadGeneratorInstallInfo fetched successfully")

	defer func() {
		if loadGeneratorInstallInfo.Status == "starting" {
			loadGeneratorInstallInfo.Status = "failed"
			err = l.loadRepo.UpdateLoadGeneratorInstallInfoTx(ctx, loadGeneratorInstallInfo)
			if err != nil {
				log.Error().Msgf("Error updating LoadGeneratorInstallInfo to failed status; %v", err)
			}
		}
	}()

	log.Printf("install info : %+v\n", loadGeneratorInstallInfo)

	// remote && server len == 0
	// local && loadGeneratorInstallInfo.status != "installed"

	installLocation := param.InstallLocation
	installScriptPath := utils.JoinRootPathWith("/script/install-jmeter.sh")

	switch installLocation {
	case constant.Local:
		log.Info().Msgf("Starting local installation of JMeter")

		jmeterPath := config.AppConfig.Load.JMeter.Dir
		jmeterVersion := config.AppConfig.Load.JMeter.Version

		exist := utils.ExistCheck(jmeterPath) && utils.ExistCheck(jmeterPath+"/apache-jmeter-"+jmeterVersion)
		if !exist {
			err := utils.Script(installScriptPath, []string{
				fmt.Sprintf("JMETER_WORK_DIR=%s", jmeterPath),
				fmt.Sprintf("JMETER_VERSION=%s", jmeterVersion),
			})

			if err != nil {
				log.Error().Msgf("Error while installing JMeter locally; %v", err)
				return result, fmt.Errorf("error while installing jmeter; %s", err)
			}
		}

		log.Info().Msg("Local installation of JMeter completed successfully")
	case constant.Remote:
		log.Info().Msg("Starting remote installation of JMeter")
		// get the spec and image information
		recommendVm, err := l.getRecommendVm(ctx, param.Coordinates)
		if err != nil {
			log.Error().Msgf("Failed to get recommended VM; %v", err)
			return result, err
		}
		antVmCommonSpec := recommendVm[0].Name
		recommendVmConnName := recommendVm[0].ConnectionName

		// CB-Tumblebug에서 사용 가능한 이미지 동적 조회
		antVmCommonImage, err := l.getAvailableImage(ctx, recommendVmConnName)
		if err != nil {
			log.Error().Msgf("Failed to get available image; %v", err)
			return result, err
		}

		// 디버깅: MCI 생성에 사용될 값들 로그 출력
		log.Info().Msgf("MCI creation parameters - CommonSpec: '%s', CommonImage: '%s', ConnectionName: '%s'",
			antVmCommonSpec, antVmCommonImage, recommendVmConnName)

		// check namespace is valid or not
		err = l.validDefaultNs(ctx, antNsId)
		if err != nil {
			log.Error().Msgf("Error validating default namespace; %v", err)
			return result, err
		}

		// get the ant default mci
		antMci, err := l.getAndDefaultMci(ctx, antVmCommonSpec, antVmCommonImage, recommendVmConnName)
		if err != nil {
			log.Error().Msgf("Error getting or creating default mci; %v", err)
			return result, err
		}

		// if server is not running state, try to resume and get mci information
		retryCount := config.AppConfig.Load.Retry
		for retryCount > 0 && antMci.StatusCount.CountRunning < 1 {
			log.Info().Msgf("Attempting to resume MCI, retry count: %d", retryCount)

			err = l.tumblebugClient.ControlLifecycleWithContext(ctx, antNsId, antMci.Id, "resume")
			if err != nil {
				log.Error().Msgf("Error resuming MCI; %v", err)
				return result, err
			}
			time.Sleep(defaultDelay)
			antMci, err = l.getAndDefaultMci(ctx, antVmCommonSpec, antVmCommonImage, recommendVmConnName)
			if err != nil {
				log.Error().Msgf("Error getting MCI after resume attempt; %v", err)
				return result, err
			}

			retryCount = retryCount - 1
		}

		if antMci.StatusCount.CountRunning < 1 {
			log.Error().Msg("No running VM on ant default MCI")
			return result, errors.New("there is no running vm on ant default mci")
		}

		addAuthorizedKeyCommand, err := getAddAuthorizedKeyCommand(antPrivKeyName, antPubKeyName)
		if err != nil {
			log.Error().Msgf("Error getting add authorized key command; %v", err)
			return result, err
		}

		installationCommand, err := utils.ReadToString(installScriptPath)
		if err != nil {
			log.Error().Msgf("Error reading installation script; %v", err)
			return result, err
		}

		commandReq := tumblebug.SendCommandReq{
			Command: []string{installationCommand, addAuthorizedKeyCommand},
		}

		_, err = l.tumblebugClient.CommandToMciWithContext(ctx, antNsId, antMci.Id, commandReq)
		if err != nil {
			log.Error().Msgf("Error sending command to MCI; %v", err)
			return result, err
		}

		log.Info().Msg("Commands sent to MCI successfully")

		marking := make(map[string]LoadGeneratorServer)
		deleteChecker := make(map[uint]bool)
		deleteList := make([]uint, 0)

		for _, v := range loadGeneratorInstallInfo.LoadGeneratorServers {
			marking[v.VmUid] = v
			deleteChecker[v.ID] = false
		}

		loadGeneratorServers := make([]LoadGeneratorServer, 0)

		for i, vm := range antMci.Vm {
			var loadGeneratorServer LoadGeneratorServer

			l, ok := marking[vm.Uid]

			if ok {
				deleteChecker[l.ID] = true
				loadGeneratorServer = l
				loadGeneratorServer.VmUid = vm.Uid
				loadGeneratorServer.VmName = vm.Name
				loadGeneratorServer.ImageName = vm.CspImageName
				loadGeneratorServer.Csp = vm.ConnectionConfig.ProviderName
				loadGeneratorServer.Region = vm.Region.Region
				loadGeneratorServer.Zone = vm.Region.Zone
				loadGeneratorServer.PublicIp = vm.PublicIP
				loadGeneratorServer.PrivateIp = vm.PrivateIP
				loadGeneratorServer.PublicDns = vm.PublicDNS
				loadGeneratorServer.MachineType = vm.CspSpecName
				loadGeneratorServer.Status = vm.Status
				loadGeneratorServer.SshPort = vm.SSHPort
				loadGeneratorServer.Lat = fmt.Sprintf("%f", vm.Location.Latitude)
				loadGeneratorServer.Lon = fmt.Sprintf("%f", vm.Location.Longitude)
				loadGeneratorServer.Username = vm.VMUserName
				loadGeneratorServer.VmId = vm.Id
				loadGeneratorServer.StartTime = vm.CreatedTime
				loadGeneratorServer.AdditionalVmKey = vm.CspResourceId
				loadGeneratorServer.Label = "temp-label"
				loadGeneratorServer.IsCluster = false
				loadGeneratorServer.IsMaster = i == 0
				loadGeneratorServer.ClusterSize = uint64(len(antMci.Vm))
			} else {
				loadGeneratorServer = LoadGeneratorServer{
					VmUid:           vm.Uid,
					VmName:          vm.Name,
					ImageName:       vm.CspImageName,
					Csp:             vm.ConnectionConfig.ProviderName,
					Region:          vm.Region.Region,
					Zone:            vm.Region.Zone,
					PublicIp:        vm.PublicIP,
					PrivateIp:       vm.PrivateIP,
					PublicDns:       vm.PublicDNS,
					MachineType:     vm.CspSpecName,
					Status:          vm.Status,
					SshPort:         vm.SSHPort,
					Lat:             fmt.Sprintf("%f", vm.Location.Latitude),
					Lon:             fmt.Sprintf("%f", vm.Location.Longitude),
					Username:        vm.VMUserName,
					VmId:            vm.Id,
					StartTime:       vm.CreatedTime,
					AdditionalVmKey: vm.CspResourceId,
					Label:           "temp-label",
					IsCluster:       false,
					IsMaster:        i == 0,
					ClusterSize:     uint64(len(antMci.Vm)),
				}
			}

			loadGeneratorServers = append(loadGeneratorServers, loadGeneratorServer)
		}

		for id, ok := range deleteChecker {
			if !ok {
				deleteList = append(deleteList, id)
			}
		}

		if len(deleteList) > 0 {
			err = l.loadRepo.DeleteLoadGeneratorServerTx(ctx, deleteList)
			if err != nil {
				log.Error().Msgf("Error delete load generator list; %s", err)
				return result, err
			}
		}

		loadGeneratorInstallInfo.LoadGeneratorServers = loadGeneratorServers
		loadGeneratorInstallInfo.PublicKeyName = antPubKeyName
		loadGeneratorInstallInfo.PrivateKeyName = antPrivKeyName
	}

	loadGeneratorInstallInfo.Status = "installed"
	err = l.loadRepo.UpdateLoadGeneratorInstallInfoTx(ctx, loadGeneratorInstallInfo)
	if err != nil {
		log.Error().Msgf("Error updating LoadGeneratorInstallInfo after installed; %v", err)
		return result, err
	}

	log.Info().Msg("LoadGeneratorInstallInfo updated successfully")

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

	log.Info().Msg("InstallLoadGenerator completed successfully")

	return result, nil
}

// getAndDefaultMci retrieves or creates the default MCI.
func (l *LoadService) getAndDefaultMci(ctx context.Context, antVmCommonSpec, antVmCommonImage, antVmConnectionName string) (tumblebug.MciRes, error) {
	var antMci tumblebug.MciRes
	var err error
	antMci, err = l.tumblebugClient.GetMciWithContext(ctx, antNsId, antMciId)
	if err != nil {
		if errors.Is(err, tumblebug.ErrNotFound) {
			// CB-Tumblebug will automatically create SSH key, VNet, Security Group, etc.
			log.Info().Msg("MCI not found, creating new MCI with dynamic resource provisioning")

			dynamicMciArg := tumblebug.DynamicMciReq{
				Description:     antMciDescription,
				InstallMonAgent: antInstallMonAgent,
				Label:           map[string]string{antLabelKey: antMciLabel},
				Name:            antMciId,
				SystemLabel:     "",
				SubGroups: []tumblebug.DynamicVmReq{ // v0.11.8: VM -> SubGroups
					{
						ImageId:        antVmCommonImage,
						SpecId:         antVmCommonSpec,
						ConnectionName: antVmConnectionName,
						Description:    antVmDescription,
						Label:          map[string]string{antLabelKey: antVmLabel},
						Name:           antVmName,
						RootDiskSize:   antVmRootDiskSize,
						RootDiskType:   antVmRootDiskType,
						SubGroupSize:   antVmSubGroupSize,
						VMUserPassword: antVmUserPassword,
						// SSH key, VNet, Security Group will be auto-created by CB-Tumblebug
					},
				},
			}
			antMci, err = l.tumblebugClient.DynamicMciWithContext(ctx, antNsId, dynamicMciArg)
			time.Sleep(defaultDelay)
			if err != nil {
				return antMci, err
			}
		} else {
			return antMci, err
		}
	} else if antMci.Vm != nil && len(antMci.Vm) == 0 {
		// CB-Tumblebug will automatically create SSH key, VNet, Security Group, etc.
		log.Info().Msg("MCI exists but no VMs, adding VM with dynamic resource provisioning")

		dynamicVmArg := tumblebug.DynamicVmReq{
			ImageId:        antVmCommonImage,
			SpecId:         antVmCommonSpec,
			ConnectionName: antVmConnectionName,
			Description:    antVmDescription,
			Label:          map[string]string{antLabelKey: antVmLabel},
			Name:           antVmName,
			RootDiskSize:   antVmRootDiskSize,
			RootDiskType:   antVmRootDiskType,
			SubGroupSize:   antVmSubGroupSize,
			VMUserPassword: antVmUserPassword,
			// SSH key, VNet, Security Group will be auto-created by CB-Tumblebug
		}

		antMci, err = l.tumblebugClient.DynamicVmWithContext(ctx, antNsId, antMciId, dynamicVmArg)
		time.Sleep(defaultDelay)
		if err != nil {
			return antMci, err
		}
	}
	return antMci, nil
}

// getRecommendVm retrieves recommendVm to specify the location of provisioning.
func (l *LoadService) getRecommendVm(ctx context.Context, coordinates []string) (tumblebug.SpecInfoList, error) {
	// config.yaml에서 스펙 요구사항 동적 로드
	specConfig := config.AppConfig.Load.Spec

	recommendVmArg := tumblebug.RecommendVmReq{
		Filter: tumblebug.Filter{
			Policy: []tumblebug.FilterPolicy{
				{
					Condition: []tumblebug.Condition{
						{
							Operand:  fmt.Sprintf("%d", specConfig.MinVcpu),
							Operator: ">=",
						},
						{
							Operand:  fmt.Sprintf("%d", specConfig.MaxVcpu),
							Operator: "<=",
						},
					},
					Metric: "vCPU",
				},
				{
					Condition: []tumblebug.Condition{
						{
							Operand:  fmt.Sprintf("%d", specConfig.MinMemory),
							Operator: ">=",
						},
						{
							Operand:  fmt.Sprintf("%d", specConfig.MaxMemory),
							Operator: "<=",
						},
					},
					Metric: "memoryGiB",
				},
				{
					Condition: []tumblebug.Condition{
						{
							Operand: specConfig.Provider,
						},
					},
					Metric: "providerName",
				},
				{
					Condition: []tumblebug.Condition{
						{
							Operand: specConfig.Architecture,
						},
					},
					Metric: "architecture",
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

	// 디버깅: 스펙 요구사항 로그 출력
	log.Info().Msgf("VM spec requirements - vCPU: %d-%d, Memory: %d-%d GB, Provider: %s, Architecture: %s",
		specConfig.MinVcpu, specConfig.MaxVcpu, specConfig.MinMemory, specConfig.MaxMemory, specConfig.Provider, specConfig.Architecture)

	recommendRes, err := l.tumblebugClient.GetRecommendVmWithContext(ctx, recommendVmArg)

	if err != nil {
		return nil, err
	}

	if len(recommendRes) == 0 {
		return nil, errors.New("there is no recommended vm list")
	}

	// 디버깅: 추천된 VM 정보 로그 출력
	log.Info().Msgf("Recommended VM count: %d", len(recommendRes))
	if len(recommendRes) > 0 {
		log.Info().Msgf("First recommended VM - ID: %s, Name: %s, CspSpecName: %s, ProviderName: %s, RegionName: %s",
			recommendRes[0].Id, recommendRes[0].Name, recommendRes[0].CspSpecName, recommendRes[0].ProviderName, recommendRes[0].RegionName)
	}

	return recommendRes, nil
}

// getAvailableImage retrieves an available image for the specified connection
func (l *LoadService) getAvailableImage(ctx context.Context, connectionName string) (string, error) {
	// CB-Tumblebug에서 사용 가능한 이미지 목록 조회 시도
	images, err := l.tumblebugClient.GetAvailableImagesWithContext(ctx, connectionName)
	if err != nil {
		log.Warn().Msgf("Failed to get available images from CB-Tumblebug; %v", err)
	} else {
		log.Info().Msgf("Found %d available images for connection: %s", len(images), connectionName)

		// 선호하는 OS 찾기
		preferredOs := config.AppConfig.Load.Image.PreferredOs
		fallbackOs := config.AppConfig.Load.Image.FallbackOs

		// 선호하는 OS로 이미지 찾기
		for _, image := range images {
			if strings.Contains(strings.ToLower(image.Name), strings.ToLower(preferredOs)) {
				log.Info().Msgf("Found preferred image: %s (ID: %s)", image.Name, image.Id)
				return image.Name, nil
			}
		}

		// 대체 OS로 이미지 찾기
		for _, image := range images {
			if strings.Contains(strings.ToLower(image.Name), strings.ToLower(fallbackOs)) {
				log.Info().Msgf("Found fallback image: %s (ID: %s)", image.Name, image.Id)
				return image.Name, nil
			}
		}

		// Ubuntu 계열 이미지 찾기
		for _, image := range images {
			if strings.Contains(strings.ToLower(image.Name), "ubuntu") {
				log.Info().Msgf("Found Ubuntu image: %s (ID: %s)", image.Name, image.Id)
				return image.Name, nil
			}
		}

		// 첫 번째 이미지 사용
		if len(images) > 0 {
			log.Info().Msgf("Using first available image: %s (ID: %s)", images[0].Name, images[0].Id)
			return images[0].Name, nil
		}
	}

	// CB-Tumblebug에 이미지가 없는 경우, config.yaml의 기본 이미지 사용
	log.Info().Msgf("No images found in CB-Tumblebug, using configured default images for connection: %s", connectionName)

	// 연결명에서 리전 추출 (예: "aws-ap-northeast-2" -> "ap-northeast-2")
	region := l.extractRegionFromConnection(connectionName)
	if region == "" {
		return "", fmt.Errorf("cannot extract region from connection name: %s", connectionName)
	}

	// AWS 이미지 목록에서 해당 리전의 이미지 찾기
	awsImages := config.AppConfig.Load.Image.AwsImages
	if imageId, exists := awsImages[region]; exists {
		log.Info().Msgf("Using configured AWS image for region %s: %s", region, imageId)
		return imageId, nil
	}

	// 기본 이미지 사용
	defaultImage := "ami-0c76973fbe0ee100c" // Ubuntu 22.04 LTS
	log.Info().Msgf("Using default AWS image for region %s: %s", region, defaultImage)
	return defaultImage, nil
}

// extractRegionFromConnection extracts region from connection name
func (l *LoadService) extractRegionFromConnection(connectionName string) string {
	// "aws-ap-northeast-2" -> "ap-northeast-2"
	if strings.HasPrefix(connectionName, "aws-") {
		return strings.TrimPrefix(connectionName, "aws-")
	}
	return ""
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

func (l *LoadService) UninstallLoadGenerator(param UninstallLoadGeneratorParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	loadGeneratorInstallInfo, err := l.loadRepo.GetValidLoadGeneratorInstallInfoByIdTx(ctx, param.LoadGeneratorInstallInfoId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Msgf("Cannot find valid load generator install info; %v", err)
			return errors.New("cannot find valid load generator install info")
		}
		log.Error().Msgf("Error retrieving load generator install info; %v", err)
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
			log.Error().Msgf("Error while uninstalling load generator; %s", err)
			return fmt.Errorf("error while uninstalling load generator: %s", err)
		}
	case constant.Remote:

		uninstallCommand, err := utils.ReadToString(uninstallScriptPath)
		if err != nil {
			log.Error().Msgf("Error reading uninstall script; %v", err)
			return err
		}

		commandReq := tumblebug.SendCommandReq{
			Command: []string{uninstallCommand},
		}

		_, err = l.tumblebugClient.CommandToMciWithContext(ctx, antNsId, antMciId, commandReq)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error().Msg("VM is not in running state. Cannot connect to the VMs.")
				return errors.New("vm is not running state. cannot connect to the vms")
			}
			log.Error().Msgf("Error sending uninstall command to MCI; %v", err)
			return err
		}

		// err = l.tumblebugClient.ControlLifecycleWithContext(ctx, antNsId, antMciId, "suspend")
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
		log.Error().Msgf("Error updating load generator install info; %v", err)
		return err
	}

	log.Info().Msg("Successfully uninstalled load generator.")
	return nil
}

func (l *LoadService) GetAllLoadGeneratorInstallInfo(param GetAllLoadGeneratorInstallInfoParam) (GetAllLoadGeneratorInstallInfoResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result GetAllLoadGeneratorInstallInfoResult
	var infos []LoadGeneratorInstallInfoResult
	pagedResult, totalRows, err := l.loadRepo.GetPagingLoadGeneratorInstallInfosTx(ctx, param)

	if err != nil {
		log.Error().Msgf("Error fetching paged load generator install infos; %v", err)
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
	log.Info().Msgf("Fetched %d load generator install info results.", len(infos))

	return result, nil
}
