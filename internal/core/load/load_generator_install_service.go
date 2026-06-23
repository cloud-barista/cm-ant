package load

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	antMciDescription  = "Default MCI for Cloud Migration Verification"
	antInstallMonAgent = "no"
	antLabelKey        = "ant-test-label"
	antMciLabel        = "DynamicMci,AntDefault"

	antVmDescription  = "Default VM for Cloud Migration Verification"
	antVmLabel        = "DynamicVm,AntDefault"
	antVmRootDiskSize = 0
	antVmRootDiskType = "default"
	antVmSubGroupSize = 1
	antVmUserPassword = ""

	antPubKeyName  = "id_rsa_ant.pub"
	antPrivKeyName = "id_rsa_ant"

	defaultDelay = 20 * time.Second
)

// InstallLoadGenerator installs the load generator either locally or remotely.
// Currently remote request is executing via cb-tumblebug.
func (l *LoadService) InstallLoadGenerator(param InstallLoadGeneratorParam) (LoadGeneratorInstallInfoResult, error) {
	log.Info().Msg("Starting InstallLoadGenerator")

	// config.yaml의 commandExecution 타임아웃 설정 사용
	timeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.CommandExecution)
	if err != nil {
		log.Warn().Msgf("Failed to parse commandExecution timeout, using default 5 minutes: %v", err)
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info().Msgf("Using command execution timeout: %v", timeout)

	var result LoadGeneratorInstallInfoResult

	loadGeneratorInstallInfo := &LoadGeneratorInstallInfo{
		InstallLocation: param.InstallLocation,
		InstallType:     "jmeter",
		InstallPath:     config.AppConfig.Load.JMeter.Dir,
		InstallVersion:  config.AppConfig.Load.JMeter.Version,
		Status:          "starting",
	}

	err = l.loadRepo.GetOrInsertLoadGeneratorInstallInfoTx(ctx, loadGeneratorInstallInfo)
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

		// ✅ 1. 기존 VM의 CSP 정보 조회
		var existingProvider, existingRegion, existingConnectionName string

		if param.NsId != "" && param.InfraId != "" && param.NodeId != "" {
			log.Info().Msgf("Getting CSP information from existing VM: nsId=%s, mciId=%s, vmId=%s",
				param.NsId, param.InfraId, param.NodeId)

			vmInfo, err := l.tumblebugClient.GetVmWithContext(ctx, param.NsId, param.InfraId, param.NodeId)
			if err != nil {
				log.Error().Msgf("Failed to get VM info; %v", err)
				return result, fmt.Errorf("failed to get VM info: %w", err)
			}

			existingProvider = vmInfo.ConnectionConfig.ProviderName
			existingRegion = vmInfo.ConnectionConfig.RegionDetail.RegionName
			existingConnectionName = vmInfo.ConnectionName

			log.Info().Msgf("Extracted CSP information - Provider: %s, Region: %s, ConnectionName: %s",
				existingProvider, existingRegion, existingConnectionName)
		} else {
			// 폴백: 기존 방식 사용 (하위 호환성)
			log.Warn().Msg("VM information not provided, using fallback method with default provider")
			existingProvider = "aws"          // 기본값
			existingRegion = "ap-northeast-2" // 기본값
			existingConnectionName = fmt.Sprintf("%s-%s", existingProvider, existingRegion)
		}

		// ✅ 2. 동일한 CSP로 VM 추천 요청
		recommendVm, err := l.getRecommendVm(ctx, []string{existingRegion}, existingProvider)
		if err != nil {
			log.Error().Msgf("Failed to get recommended VM for provider %s; %v", existingProvider, err)
			return result, err
		}

		antVmCommonSpec := recommendVm[0].Name

		// ✅ 기존 VM의 리전과 연결명을 사용하여 이미지 조회 (동일한 리전에 설치)
		antVmCommonImage, err := l.getAvailableImage(ctx, existingConnectionName)
		if err != nil {
			log.Error().Msgf("Failed to get available image; %v", err)
			return result, err
		}

		// 디버깅: MCI 생성에 사용될 값들 로그 출력
		log.Info().Msgf("MCI creation parameters - CommonSpec: '%s', CommonImage: '%s', ConnectionName: '%s'",
			antVmCommonSpec, antVmCommonImage, existingConnectionName)

		// check namespace is valid or not
		nsId, _, _, _ := getResourceNames()
		err = l.validDefaultNs(ctx, nsId)
		if err != nil {
			log.Error().Msgf("Error validating default namespace; %v", err)
			return result, err
		}

		// get the ant default mci
		antMci, err := l.getAndDefaultMci(ctx, antVmCommonSpec, antVmCommonImage, existingConnectionName)
		if err != nil {
			log.Error().Msgf("Error getting or creating default mci; %v", err)
			return result, err
		}

		// if server is not running state, try to resume and get mci information
		retryCount := config.AppConfig.Load.Retry
		for retryCount > 0 && antMci.StatusCount.CountRunning < 1 {
			log.Info().Msgf("Attempting to resume MCI, retry count: %d", retryCount)

			err = l.tumblebugClient.ControlLifecycleWithContext(ctx, nsId, antMci.Id, "resume")
			if err != nil {
				log.Error().Msgf("Error resuming MCI; %v", err)
				return result, err
			}
			time.Sleep(defaultDelay)
			antMci, err = l.getAndDefaultMci(ctx, antVmCommonSpec, antVmCommonImage, existingConnectionName)
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

		_, err = l.tumblebugClient.CommandToMciWithContext(ctx, nsId, antMci.Id, commandReq)
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
				// cb-tumblebug v0.12.7~ BREAKING: NodeInfo.sshPort은 integer — DTO는 string 보존
				loadGeneratorServer.SshPort = strconv.Itoa(vm.SSHPort)
				loadGeneratorServer.Lat = fmt.Sprintf("%f", vm.Location.Latitude)
				loadGeneratorServer.Lon = fmt.Sprintf("%f", vm.Location.Longitude)
				loadGeneratorServer.Username = vm.VMUserName
				loadGeneratorServer.NodeId = vm.Id
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
					SshPort:         strconv.Itoa(vm.SSHPort),
					Lat:             fmt.Sprintf("%f", vm.Location.Latitude),
					Lon:             fmt.Sprintf("%f", vm.Location.Longitude),
					Username:        vm.VMUserName,
					NodeId:            vm.Id,
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
			NodeId:            l.NodeId,
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
	nsId, mciId, vmName, _ := getResourceNames()
	antMci, err = l.tumblebugClient.GetMciWithContext(ctx, nsId, mciId)
	if err != nil {
		if errors.Is(err, tumblebug.ErrNotFound) {
			// CB-Tumblebug will automatically create SSH key, VNet, Security Group, etc.
			log.Info().Msg("MCI not found, creating new MCI with dynamic resource provisioning")

			dynamicMciArg := tumblebug.DynamicMciReq{
				Description:     antMciDescription,
				InstallMonAgent: antInstallMonAgent,
				Label:           map[string]string{antLabelKey: antMciLabel},
				Name:            mciId,
				SystemLabel:     "",
				SubGroups: []tumblebug.DynamicVmReq{ // v0.11.8: VM -> SubGroups
					{
						ImageId:        antVmCommonImage,
						SpecId:         antVmCommonSpec,
						ConnectionName: antVmConnectionName,
						Description:    antVmDescription,
						Label:          map[string]string{antLabelKey: antVmLabel},
						Name:           vmName,
						RootDiskSize:   antVmRootDiskSize,
						RootDiskType:   antVmRootDiskType,
						SubGroupSize:   antVmSubGroupSize,
						VMUserPassword: antVmUserPassword,
						// SSH key, VNet, Security Group will be auto-created by CB-Tumblebug
					},
				},
			}
			antMci, err = l.tumblebugClient.DynamicMciWithContext(ctx, nsId, dynamicMciArg)
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
			Name:           vmName,
			RootDiskSize:   antVmRootDiskSize,
			RootDiskType:   antVmRootDiskType,
			SubGroupSize:   antVmSubGroupSize,
			VMUserPassword: antVmUserPassword,
			// SSH key, VNet, Security Group will be auto-created by CB-Tumblebug
		}

		antMci, err = l.tumblebugClient.DynamicVmWithContext(ctx, nsId, mciId, dynamicVmArg)
		time.Sleep(defaultDelay)
		if err != nil {
			return antMci, err
		}
	}
	return antMci, nil
}

// getRecommendVm retrieves recommendVm to specify the location of provisioning.
func (l *LoadService) getRecommendVm(ctx context.Context, coordinates []string, provider string) (tumblebug.SpecInfoList, error) {
	// config.yaml에서 스펙 요구사항 동적 로드
	specConfig := config.AppConfig.Load.Spec

	recommendVmArg := tumblebug.RecommendVmReq{
		Filter: tumblebug.FilterInfo{
			Policy: []tumblebug.FilterCondition{
				{
					Metric: "vCPU",
					Condition: []tumblebug.Operation{
						{
							Operator: ">=",
							Operand:  fmt.Sprintf("%d", specConfig.MinVcpu),
						},
						{
							Operator: "<=",
							Operand:  fmt.Sprintf("%d", specConfig.MaxVcpu),
						},
					},
				},
				{
					Metric: "memoryGiB",
					Condition: []tumblebug.Operation{
						{
							Operator: ">=",
							Operand:  fmt.Sprintf("%d", specConfig.MinMemory),
						},
						{
							Operator: "<=",
							Operand:  fmt.Sprintf("%d", specConfig.MaxMemory),
						},
					},
				},
				{
					Metric: "providerName",
					Condition: []tumblebug.Operation{
						{
							Operator: "==",
							Operand:  provider, // ✅ 동적으로 추출한 CSP 사용
						},
					},
				},
				{
					Metric: "regionName",
					Condition: []tumblebug.Operation{
						{
							Operator: "==",
							Operand:  coordinates[0], // ✅ 리전명 사용
						},
					},
				},
				{
					Metric: "architecture",
					Condition: []tumblebug.Operation{
						{
							Operator: "==",
							Operand:  specConfig.Architecture,
						},
					},
				},
			},
		},
		// cb-tumblebug v0.12.x BREAKING: RecommendSpecReq.limit string -> integer
		Limit: 1,
		Priority: tumblebug.PriorityInfo{
			Policy: []tumblebug.PriorityCondition{
				{
					Metric: "costPerHour",
					Parameter: []tumblebug.Parameter{
						{
							Key: "order",
							Val: []string{"asc"},
						},
					},
				},
			},
		},
	}

	// 디버깅: 스펙 요구사항 로그 출력
	log.Info().Msgf("VM spec requirements - vCPU: %d-%d, Memory: %d-%d GB, Provider: %s, Region: %s, Architecture: %s",
		specConfig.MinVcpu, specConfig.MaxVcpu, specConfig.MinMemory, specConfig.MaxMemory, provider, coordinates[0], specConfig.Architecture)

	// 디버깅: CB-Tumblebug API 요청 정보를 curl 형태로 출력
	requestBody, _ := json.MarshalIndent(recommendVmArg, "", "  ")
	log.Info().Msgf("CB-Tumblebug recommendSpec API request (curl format):")
	log.Info().Msgf("curl -X POST http://localhost:1323/tumblebug/recommendSpec \\")
	log.Info().Msgf("  -H 'Content-Type: application/json' \\")
	log.Info().Msgf("  -d '%s'", string(requestBody))

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
	imageConfig := config.AppConfig.Load.Image

	// 1. useFallbackImagesOnly가 true인 경우, 하드코딩된 이미지만 사용
	if imageConfig.UseFallbackImagesOnly {
		log.Info().Msg("Using fallback images only mode - skipping image search")
		return l.getFallbackImage(connectionName)
	}

	// 2. 스마트 매칭 기능이 활성화된 경우 우선 사용
	if imageConfig.UseSmartMatching {
		imageId, err := l.getAvailableImageWithSmartMatching(ctx, connectionName)
		if err == nil {
			return imageId, nil
		}
		log.Warn().Msgf("Smart matching failed, falling back to traditional method: %v", err)
	}

	// 3. 기존 방식으로 폴백
	return l.getAvailableImageTraditional(ctx, connectionName)
}

// getAvailableImageWithSmartMatching uses CB-Tumblebug v0.11.8+ smart matching
func (l *LoadService) getAvailableImageWithSmartMatching(ctx context.Context, connectionName string) (string, error) {
	// 연결명에서 프로바이더와 리전 추출 (예: "aws-ap-northeast-2" -> "aws", "ap-northeast-2")
	provider, region := l.extractProviderAndRegionFromConnection(connectionName)
	if provider == "" || region == "" {
		return "", fmt.Errorf("cannot extract provider and region from connection name: %s", connectionName)
	}

	imageConfig := config.AppConfig.Load.Image
	specConfig := config.AppConfig.Load.Spec
	searchOptions := imageConfig.SearchOptions

	// 1. matchedSpecId를 사용한 스마트 매칭 시도 (권장 방법)
	matchedSpecId := fmt.Sprintf("%s+%s+t3.small", provider, region) // 기본 스펙 사용
	log.Info().Msgf("Trying matchedSpecId approach: %s", matchedSpecId)

	searchReq := tumblebug.SearchImageRequest{
		MatchedSpecId:          matchedSpecId,
		OSType:                 imageConfig.PreferredOs,
		IncludeBasicImageOnly:  &searchOptions.IncludeBasicImageOnly,
		IsRegisteredByAsset:    &searchOptions.IsRegisteredByAsset,
		IncludeDeprecatedImage: &searchOptions.IncludeDeprecatedImage,
		MaxResults:             searchOptions.MaxResults,
	}

	images, err := l.tumblebugClient.SearchImagesWithContext(ctx, "system", searchReq)
	if err != nil {
		log.Warn().Msgf("matchedSpecId approach failed: %v", err)
	} else if len(images) > 0 {
		log.Info().Msgf("matchedSpecId found %d images for '%s', using: %s (ID: %s)",
			len(images), imageConfig.PreferredOs, images[0].Name, images[0].Id)
		return images[0].Name, nil
	}

	// 2. 기본 방법으로 선호하는 OS 검색 (includeBasicImageOnly=true)
	log.Info().Msgf("Trying basic image search for preferred OS: %s", imageConfig.PreferredOs)
	searchReq = tumblebug.SearchImageRequest{
		ProviderName:           provider,
		RegionName:             region,
		OSType:                 imageConfig.PreferredOs,
		OSArchitecture:         specConfig.Architecture,
		IncludeBasicImageOnly:  &[]bool{true}[0],
		IsRegisteredByAsset:    &searchOptions.IsRegisteredByAsset,
		IncludeDeprecatedImage: &searchOptions.IncludeDeprecatedImage,
		MaxResults:             searchOptions.MaxResults,
	}

	images, err = l.tumblebugClient.SearchImagesWithContext(ctx, "system", searchReq)
	if err != nil {
		log.Warn().Msgf("Basic image search failed for preferred OS: %v", err)
	} else if len(images) > 0 {
		log.Info().Msgf("Basic image search found %d images for preferred OS '%s', using: %s (ID: %s)",
			len(images), imageConfig.PreferredOs, images[0].Name, images[0].Id)
		return images[0].Name, nil
	}

	// 3. includeBasicImageOnly=false로 모든 이미지 검색
	log.Info().Msgf("Trying all images search for preferred OS: %s", imageConfig.PreferredOs)
	searchReq.IncludeBasicImageOnly = &[]bool{false}[0]

	images, err = l.tumblebugClient.SearchImagesWithContext(ctx, "system", searchReq)
	if err != nil {
		log.Warn().Msgf("All images search failed for preferred OS: %v", err)
	} else if len(images) > 0 {
		// 최적의 이미지 선택 로직
		selectedImage := l.selectBestImage(images, imageConfig.PreferredOs)
		log.Info().Msgf("Selected best image for preferred OS '%s': %s (ID: %s, isBasic: %v)",
			imageConfig.PreferredOs, selectedImage.Name, selectedImage.Id, selectedImage.IsBasicImage)
		return selectedImage.Name, nil
	}

	// 4. 대체 OS로 시도
	log.Info().Msgf("Trying fallback OS: %s", imageConfig.FallbackOs)
	searchReq.OSType = imageConfig.FallbackOs
	searchReq.IncludeBasicImageOnly = &[]bool{true}[0]

	images, err = l.tumblebugClient.SearchImagesWithContext(ctx, "system", searchReq)
	if err != nil {
		log.Warn().Msgf("Basic image search failed for fallback OS: %v", err)
	} else if len(images) > 0 {
		log.Info().Msgf("Found %d images for fallback OS '%s', using: %s (ID: %s)",
			len(images), imageConfig.FallbackOs, images[0].Name, images[0].Id)
		return images[0].Name, nil
	}

	return "", fmt.Errorf("no images found with smart matching for connection: %s", connectionName)
}

// getFallbackImage returns a hardcoded image from config.yaml fallbackImages
func (l *LoadService) getFallbackImage(connectionName string) (string, error) {
	// 연결명에서 프로바이더와 리전 추출 (예: "aws-ap-northeast-2" -> "aws", "ap-northeast-2")
	provider, region := l.extractProviderAndRegionFromConnection(connectionName)
	if provider == "" || region == "" {
		return "", fmt.Errorf("cannot extract provider and region from connection name: %s", connectionName)
	}

	imageConfig := config.AppConfig.Load.Image
	fallbackImages := imageConfig.FallbackImages

	// 프로바이더별 이미지 확인
	if providerImages, exists := fallbackImages[provider]; exists {
		// 리전별 이미지 확인
		if imageId, exists := providerImages[region]; exists {
			log.Info().Msgf("Using fallback image for %s-%s: %s", provider, region, imageId)
			return imageId, nil
		}
		log.Warn().Msgf("No fallback image found for region %s in provider %s", region, provider)
	} else {
		log.Warn().Msgf("No fallback images configured for provider %s", provider)
	}

	// 프로바이더별 기본 이미지 찾기 (첫 번째 리전의 이미지 사용)
	if providerImages, exists := fallbackImages[provider]; exists {
		for regionKey, imageId := range providerImages {
			log.Info().Msgf("Using fallback image from different region %s for %s-%s: %s",
				regionKey, provider, region, imageId)
			return imageId, nil
		}
	}

	return "", fmt.Errorf("no fallback images available for connection: %s", connectionName)
}

// selectBestImage selects the most appropriate image from the search results
// v0.12.1: OSType 필드를 활용하여 더 정확한 이미지 선택
func (l *LoadService) selectBestImage(images []tumblebug.ImageInfo, preferredOS string) tumblebug.ImageInfo {
	// 우선순위별로 이미지 분류
	var basicImages []tumblebug.ImageInfo
	var osMatchedImages []tumblebug.ImageInfo
	var serverImages []tumblebug.ImageInfo
	var otherImages []tumblebug.ImageInfo

	for _, img := range images {
		// v0.12.1: OSType 필드를 사용하여 더 정확한 매칭
		osType := img.GetOSType() // v0.11.19 및 v0.12.1 호환

		// isBasicImage=true인 이미지 우선
		if img.IsBasicImage {
			basicImages = append(basicImages, img)
		} else if osType != "" && strings.Contains(strings.ToLower(osType), strings.ToLower(preferredOS)) {
			// v0.12.1: OSType이 선호 OS와 일치하는 경우
			osMatchedImages = append(osMatchedImages, img)
		} else if strings.Contains(strings.ToLower(img.Name), "server") ||
			strings.Contains(strings.ToLower(img.Name), "jammy") {
			// server 또는 jammy가 포함된 이미지
			serverImages = append(serverImages, img)
		} else {
			// 기타 이미지
			otherImages = append(otherImages, img)
		}
	}

	// 1순위: isBasicImage=true인 이미지
	if len(basicImages) > 0 {
		// 기본 이미지 중에서도 OSType이 선호 OS와 일치하는 것 우선
		for _, img := range basicImages {
			osType := img.GetOSType()
			if osType != "" && strings.Contains(strings.ToLower(osType), strings.ToLower(preferredOS)) {
				log.Info().Msgf("Found basic image with matching OSType '%s': %s", osType, img.Name)
				return img
			}
		}
		log.Info().Msgf("Found %d basic images, selecting first one: %s", len(basicImages), basicImages[0].Name)
		return basicImages[0]
	}

	// 2순위: v0.12.1 OSType이 선호 OS와 일치하는 이미지
	if len(osMatchedImages) > 0 {
		for _, img := range osMatchedImages {
			// daily, pro, minimal, fips, k8s, deep-learning 등 제외
			if l.isSuitableImage(img) {
				log.Info().Msgf("Found OSType matched image: %s (OSType: %s)", img.Name, img.GetOSType())
				return img
			}
		}
	}

	// 3순위: server/jammy 이미지 (daily, pro 제외)
	for _, img := range serverImages {
		if l.isSuitableImage(img) {
			log.Info().Msgf("Found suitable server image: %s", img.Name)
			return img
		}
	}

	// 4순위: 첫 번째 이미지 (폴백)
	if len(images) > 0 {
		log.Warn().Msgf("No optimal image found, using first available: %s", images[0].Name)
		return images[0]
	}

	// 이론적으로는 도달하지 않음
	return tumblebug.ImageInfo{}
}

// isSuitableImage checks if an image is suitable (excludes daily, pro, minimal, fips, k8s, etc.)
func (l *LoadService) isSuitableImage(img tumblebug.ImageInfo) bool {
	name := strings.ToLower(img.Name)
	osDistribution := strings.ToLower(img.OSDistribution)

	// 제외할 패턴 목록
	excludePatterns := []string{
		"daily", "pro", "minimal", "fips", "k8s", "kubernetes",
		"container", "deep", "learning", "neuron", "parallelcluster",
		"sql", "ecs", "eks", "bottlerocket",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(name, pattern) || strings.Contains(osDistribution, pattern) {
			return false
		}
	}

	// v0.12.1: Kubernetes 이미지 제외
	if img.IsKubernetesImage {
		return false
	}

	// v0.12.1: Deprecated 이미지 제외
	if img.ImageStatus == tumblebug.ImageDeprecated {
		return false
	}

	return true
}

// getAvailableImageTraditional uses the traditional image selection method
func (l *LoadService) getAvailableImageTraditional(ctx context.Context, connectionName string) (string, error) {
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

	// CB-Tumblebug에 이미지가 없는 경우, config.yaml의 폴백 이미지 사용
	log.Info().Msgf("No images found in CB-Tumblebug, using configured fallback images for connection: %s", connectionName)

	// 연결명에서 CSP와 리전 추출 (예: "aws-ap-northeast-2" -> "aws", "ap-northeast-2")
	provider, region := l.extractProviderAndRegionFromConnection(connectionName)
	if provider == "" || region == "" {
		return "", fmt.Errorf("cannot extract provider and region from connection name: %s", connectionName)
	}

	// 폴백 이미지 목록에서 해당 CSP와 리전의 이미지 찾기
	fallbackImages := config.AppConfig.Load.Image.FallbackImages
	if providerImages, exists := fallbackImages[provider]; exists {
		if imageId, exists := providerImages[region]; exists {
			log.Info().Msgf("Using configured fallback image for provider %s, region %s: %s", provider, region, imageId)
			return imageId, nil
		}
	}

	// 폴백 이미지도 없는 경우 에러 반환
	return "", fmt.Errorf("no available image found for provider %s, region %s. Please check CB-Tumblebug image availability or configure fallback images in config.yaml", provider, region)
}

// extractProviderAndRegionFromConnection extracts provider and region from connection name
func (l *LoadService) extractProviderAndRegionFromConnection(connectionName string) (string, string) {
	// "aws-ap-northeast-2" -> "aws", "ap-northeast-2"
	// "azure-koreacentral" -> "azure", "koreacentral"
	// "gcp-asia-northeast3" -> "gcp", "asia-northeast3"
	// "ncp-KR" -> "ncp", "KR"

	parts := strings.Split(connectionName, "-")
	if len(parts) < 2 {
		return "", ""
	}

	provider := parts[0]
	region := strings.Join(parts[1:], "-")

	return provider, region
}

// validDefaultNs checks if the default namespace exists, and creates it if not.
func (l *LoadService) validDefaultNs(ctx context.Context, nsId string) error {
	_, err := l.tumblebugClient.GetNsWithContext(ctx, nsId)
	if err != nil && errors.Is(err, tumblebug.ErrNotFound) {

		arg := tumblebug.CreateNsReq{
			Name:        nsId,
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

		nsId, mciId, _, _ := getResourceNames()
		_, err = l.tumblebugClient.CommandToMciWithContext(ctx, nsId, mciId, commandReq)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error().Msg("VM is not in running state. Cannot connect to the VMs.")
				return errors.New("vm is not running state. cannot connect to the vms")
			}
			log.Error().Msgf("Error sending uninstall command to MCI; %v", err)
			return err
		}

		// err = l.tumblebugClient.ControlLifecycleWithContext(ctx, nsId, mciId, "suspend")
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
				NodeId:            s.NodeId,
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
