package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/outbound/tumblebug"
	"github.com/melbahja/goph"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/repository"
	"github.com/cloud-barista/cm-ant/pkg/load/managers"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

// Ubuntu Server 22.04 LTS
var regionImageMap = map[string]string{
	"ap-south-2":     "",
	"ap-south-1":     "ami-05e00961530ae1b55",
	"eu-south-1":     "",
	"eu-south-2":     "",
	"me-central-1":   "",
	"il-central-1":   "",
	"ca-central-1":   "ami-0083d3f8b2a6c7a81",
	"eu-central-1":   "ami-026c3177c9bd54288",
	"eu-central-2":   "",
	"us-west-1":      "ami-036cafe742923b3d9",
	"us-west-2":      "ami-03c983f9003cb9cd1",
	"af-south-1":     "ami-0f256846cac23da94",
	"eu-north-1":     "ami-011e54f70c1c91e17",
	"eu-west-3":      "ami-0326f9264af7e51e2",
	"eu-west-2":      "ami-09627c82937ccdd6d",
	"eu-west-1":      "ami-0607a9783dd204cae",
	"ap-northeast-3": "ami-0c1531991482a24e1",
	"ap-northeast-2": "ami-01ed8ade75d4eee2f",
	"me-south-1":     "",
	"ap-northeast-1": "ami-0595d6e81396a9efb",
	"sa-east-1":      "ami-0cdc2f24b2f67ea17",
	"ap-east-1":      "",
	"ca-west-1":      "",
	"ap-southeast-1": "ami-0be48b687295f8bd6",
	"ap-southeast-2": "ami-076fe60835f136dc9",
	"ap-southeast-3": "",
	"ap-southeast-4": "",
	"us-east-1":      "ami-0e001c9271cf7f3b9",
	"us-east-2":      "ami-0f30a9c3a48f3fa79",
}

func InstallLoadTester(antTargetServerReq *api.AntTargetServerReq) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	antTargetServerMcis, err := tumblebug.GetMcisObjectWithContext(ctx, antTargetServerReq.NsId, antTargetServerReq.McisId)

	if err != nil {
		return 0, err
	}

	if len(antTargetServerMcis.VMs) == 0 {
		return 0, errors.New("cannot find any vm in target mcis")
	}

	var antTargetServerVm tumblebug.VmRes

	for _, v := range antTargetServerMcis.VMs {
		if strings.EqualFold(v.ID, antTargetServerReq.VmId) {
			antTargetServerVm = v
		}
	}

	if antTargetServerVm.ID == "" {
		return 0, errors.New(fmt.Sprintf("%s does not exist", antTargetServerReq.VmId))
	}

	connectionId := antTargetServerVm.ConnectionName
	antNsId := antTargetServerReq.NsId
	antVNetId := antTargetServerVm.VNetID
	antSubnetId := antTargetServerVm.SubnetID
	antCspRegion := antTargetServerVm.Region.Region
	antCspImageId, ok := regionImageMap[antCspRegion]
	antUsername := "cb-user"
	antSgId := "ant-load-test-sg"
	antSshId := "ant-load-test-ssh"
	antImageId := "ant-load-test-image"
	antSpecId := "ant-load-test-spec"
	antCspSpecName := "t3.small"
	antVmId := "ant-load-test-vm"
	antMcisId := "ant-load-test-mcis"

	log.Println(connectionId, antNsId, antVNetId, antSubnetId, antUsername, antCspRegion, antCspImageId, antSgId, antSshId, antImageId, antSpecId, antCspSpecName, antVmId, antMcisId)

	if !ok {
		return 0, errors.New("region base ubuntu 22.04 lts image doesn't exist")
	}

	var wg sync.WaitGroup
	goroutine := 4
	errChan := make(chan error, goroutine)

	wg.Add(1)
	go func() {
		defer wg.Done()

		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()

		err2 := tumblebug.GetSecurityGroupWithContext(tc, antNsId, antSgId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}

		sg := tumblebug.SecurityGroupReq{
			Name:           antSgId,
			ConnectionName: connectionId,
			VNetID:         antVNetId,
			Description:    "Default Security Group for Ant load test",
			FirewallRules: []tumblebug.FirewallRuleReq{
				{FromPort: "22", ToPort: "22", IPProtocol: "tcp", Direction: "inbound", CIDR: "0.0.0.0/0"},
			},
		}
		_, err2 = tumblebug.CreateSecurityGroupWithContext(tc, antNsId, sg)
		if err2 != nil {
			errChan <- err2
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()

		_, err2 := tumblebug.GetSecureShellWithContext(tc, antNsId, antSshId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}

		ssh := tumblebug.SecureShellReq{
			ConnectionName: connectionId,
			Name:           antSshId,
			Username:       antUsername,
			Description:    "Default secure shell key for Ant load test",
		}
		sshResult, err2 := tumblebug.CreateSecureShellWithContext(ctx, antNsId, ssh)
		if err2 != nil {
			errChan <- err2
			return
		}
		home, err2 := os.UserHomeDir()
		if err2 != nil {
			errChan <- err2
			return
		}
		pemFilePath := fmt.Sprintf("%s/.ssh/%s.pem", home, sshResult.Id)

		err2 = os.WriteFile(pemFilePath, []byte(sshResult.PrivateKey), 0600)
		if err2 != nil {
			errChan <- err2
			return
		}

		log.Printf("%s.pem ssh private key file save to default ssh path", sshResult.Id)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()

		err2 := tumblebug.GetImageWithContext(tc, antNsId, antImageId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}
		// TODO add dynamic spec integration in advanced version
		image := tumblebug.ImageReq{
			ConnectionName: connectionId,
			Name:           antImageId,
			CspImageId:     antCspImageId,
			Description:    "Default machine image for Ant load test",
		}
		_, err2 = tumblebug.CreateImageWithContext(ctx, antNsId, image)
		if err2 != nil {
			errChan <- err2
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()

		err2 := tumblebug.GetSpecWithContext(tc, antNsId, antSpecId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}

		// TODO add dynamic spec integration in advanced version
		spec := tumblebug.SpecReq{
			ConnectionName: connectionId,
			Name:           antSpecId,
			CspSpecName:    antCspSpecName,
			Description:    "Default spec for Ant load test",
		}
		_, err2 = tumblebug.CreateSpecWithContext(ctx, antNsId, spec)
		if err2 != nil {
			errChan <- err2
			return
		}
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) != 0 {
		err = <-errChan
		return 0, err
	}

	antLoadGenerateServerMcis, err := tumblebug.GetMcisWithContext(ctx, antNsId, antMcisId)

	if err != nil && !errors.Is(err, tumblebug.ResourcesNotFound) {
		return 0, err
	} else if err == nil {
		if !antLoadGenerateServerMcis.IsRunning(antVmId) {
			// TODO - need to change mcis or antTargetServerVm status execution
			return 0, errors.New("vm is not running condition")
		}
	} else {
		antLoadGenerateServerReq := tumblebug.McisReq{
			Name:            antMcisId,
			Description:     "Default mcis for Ant load test",
			InstallMonAgent: "no",
			Label:           "ANT",
			SystemLabel:     "ANT",
			Vm: []tumblebug.VmReq{
				{
					SubGroupSize:     "1",
					Name:             antVmId,
					ImageId:          antImageId,
					VmUserAccount:    antUsername,
					ConnectionName:   connectionId,
					SshKeyId:         antSshId,
					SpecId:           antSpecId,
					SecurityGroupIds: []string{antSgId},
					VNetId:           antVNetId,
					SubnetId:         antSubnetId,
					Description:      "Default vm for Ant load test",
					VmUserPassword:   "",
					RootDiskType:     "default",
					RootDiskSize:     "default",
				},
			},
		}

		antLoadGenerateServerMcis, err = tumblebug.CreateMcisWithContext(ctx, antNsId, antLoadGenerateServerReq)

		if err != nil {
			return 0, err
		}

		log.Println("******************created*******************\n", antLoadGenerateServerMcis)
		time.Sleep(3 * time.Second)
	}

	log.Println("ant load generate server mcis is ready with running condition")

	ssh, err := tumblebug.GetSecureShellWithContext(ctx, antNsId, antSshId)
	if err != nil {
		return 0, err
	}

	vm := antLoadGenerateServerMcis.VMs[0]
	if err != nil {
		return 0, err
	}

	req := api.LoadEnvReq{
		InstallLocation: constant.Remote,
		NsId:            antNsId,
		McisId:          antLoadGenerateServerMcis.ID,
		VmId:            antLoadGenerateServerMcis.VmId(),
		Username:        antUsername,
		PublicIp:        vm.PublicIP,
		PemKeyPath:      filepath.Join(os.Getenv("HOME"), ".ssh", ssh.Id+".pem"),
	}

	manager := managers.NewLoadTestManager()

	err = manager.Install(&req)
	if err != nil {
		return 0, err
	}

	err = utils.AddToKnownHost(req.PemKeyPath, req.PublicIp, req.Username)
	if err != nil {
		return 0, err
	}

	loadEnv := model.LoadEnv{
		InstallLocation: req.InstallLocation,
		NsId:            req.NsId,
		McisId:          req.McisId,
		VmId:            req.VmId,
		Username:        req.Username,
		PublicIp:        req.PublicIp,
		PemKeyPath:      req.PemKeyPath,
	}

	createdEnvId, err := repository.SaveLoadTestInstallEnv(&loadEnv)
	if err != nil {
		return 0, fmt.Errorf("failed to save load test installation environment: %w", err)
	}
	log.Printf("Environment ID %d for load test is successfully created", createdEnvId)

	return createdEnvId, nil
}

func UninstallLoadTester(envId string) error {
	manager := managers.NewLoadTestManager()

	var loadEnvReq api.LoadEnvReq

	loadEnv, err := repository.GetEnvironment(envId)
	if err != nil {
		return err
	}

	loadEnvReq.InstallLocation = (*loadEnv).InstallLocation
	loadEnvReq.Username = (*loadEnv).Username
	loadEnvReq.NsId = (*loadEnv).NsId
	loadEnvReq.McisId = (*loadEnv).McisId
	loadEnvReq.VmId = (*loadEnv).VmId
	loadEnvReq.PublicIp = (*loadEnv).PublicIp
	loadEnvReq.PemKeyPath = (*loadEnv).PemKeyPath

	if err := manager.Uninstall(&loadEnvReq); err != nil {
		return fmt.Errorf("failed to uninstall load tester: %w", err)
	}

	err = repository.DeleteLoadTestInstallEnv(envId)
	if err != nil {
		return fmt.Errorf("failed to delete load test installation environment: %w", err)
	}

	return nil
}

func ExecuteLoadTest(loadTestReq *api.LoadExecutionConfigReq) (string, error) {
	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
	loadTestReq.LoadTestKey = loadTestKey

	if loadTestReq.EnvId == "" {
		envId, err := InstallLoadTester(&loadTestReq.AntTargetServerReq)
		if err != nil {
			return "", err
		}

		loadTestReq.EnvId = fmt.Sprintf("%d", envId)
	}

	// check env
	if err := prepareEnvironment(loadTestReq); err != nil {
		return "", err
	}

	loadTestManager := managers.NewLoadTestManager()

	go runLoadTest(loadTestManager, loadTestReq, loadTestKey)

	_, err := repository.SaveLoadTestExecution(loadTestReq)
	if err != nil {
		return "", err
	}

	return loadTestKey, nil
}

func GetLoadTestResult(testKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(testKey)
	if err != nil {
		return nil, err
	}

	if loadExecutionState.ExecutionStatus == constant.Processing {
		return nil, errors.New("load test is processing")
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)

	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return nil, err
	}

	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetResult(loadEnv, testKey, format)
	if err != nil {
		return nil, fmt.Errorf("error on [GetLoadTestResult()]; %s", err)
	}
	return result, nil
}

func prepareEnvironment(loadTestReq *api.LoadExecutionConfigReq) error {
	if loadTestReq.EnvId == "" {
		return nil
	}

	loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	if loadEnv != nil && loadTestReq.LoadEnvReq.InstallLocation == "" {
		loadTestReq.LoadEnvReq = convertToLoadEnvReq(loadEnv)
	}

	return nil
}

func convertToLoadEnvReq(loadEnv *model.LoadEnv) api.LoadEnvReq {
	return api.LoadEnvReq{
		InstallLocation: loadEnv.InstallLocation,
		Username:        loadEnv.Username,
		PublicIp:        loadEnv.PublicIp,
		PemKeyPath:      loadEnv.PemKeyPath,
		NsId:            loadEnv.NsId,
		McisId:          loadEnv.McisId,
		VmId:            loadEnv.VmId,
	}
}

func runLoadTest(loadTestManager managers.LoadTestManager, loadTestReq *api.LoadExecutionConfigReq, loadTestKey string) {
	log.Printf("[%s] start load test", loadTestKey)

	loadTestErr := loadTestManager.Run(loadTestReq)
	if updateErr := repository.UpdateLoadExecutionState(loadTestKey, constant.Fetching); updateErr != nil {
		log.Println(updateErr)
	}

	log.Println("load test result fetching..")

	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	resultsPrefix := []string{"", "_cpu", "_disk", "_memory", "_network"}
	resultFolderPath := configuration.JoinRootPathWith("/result/" + loadTestKey)
	err := utils.CreateFolderIfNotExist(resultFolderPath)
	if err != nil {
		log.Println(err)
		return
	}

	loadEnv := loadTestReq.LoadEnvReq
	auth, err := goph.Key(loadEnv.PemKeyPath, "")
	if err != nil {
		log.Println(err)
		return
	}

	client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)

	if err != nil {
		log.Println(err)
		return
	}

	defer client.Close()
	var wg sync.WaitGroup
	errCh := make(chan error)

	for _, p := range resultsPrefix {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()
			fileName := fmt.Sprintf("%s%s_result.csv", loadTestKey, prefix)
			resultFilePath := fmt.Sprintf("%s/result/%s", jmeterPath, fileName)
			toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

			if exist := utils.ExistCheck(toFilePath); !exist {
				err2 := client.Download(resultFilePath, toFilePath)
				if err2 != nil {
					errCh <- err2
				}
			}
		}(p)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) != 0 {
		err = <-errCh
		log.Println(err)
	}

	if loadTestErr != nil {
		log.Printf("Error during load test: %v", loadTestErr)
		if updateErr := repository.UpdateLoadExecutionStateWithNoTime(loadTestKey, constant.Failed); updateErr != nil {
			log.Println(updateErr)
		}
	} else {
		log.Printf("load test complete!")

		if updateErr := repository.UpdateLoadExecutionStateWithNoTime(loadTestKey, constant.Success); updateErr != nil {
			log.Println(updateErr)
		}
	}
}

func StopLoadTest(loadTestKeyReq api.LoadTestKeyReq) error {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKeyReq.LoadTestKey)

	if err != nil {
		return err
	}

	if loadExecutionState.IsFinished() {
		return fmt.Errorf("load test is already finished")
	}

	loadTestReq := api.LoadExecutionConfigReq{
		LoadTestKey: loadTestKeyReq.LoadTestKey,
		EnvId:       fmt.Sprintf("%d", loadExecutionState.LoadEnvID),
	}

	var env api.LoadEnvReq
	if loadTestReq.EnvId != "" {
		loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
		if err != nil {
			return err
		}

		env.InstallLocation = (*loadEnv).InstallLocation
		env.Username = (*loadEnv).Username
		env.PublicIp = (*loadEnv).PublicIp
		env.PemKeyPath = (*loadEnv).PemKeyPath
		env.NsId = (*loadEnv).NsId
		env.McisId = (*loadEnv).McisId
		env.VmId = (*loadEnv).VmId

		loadTestReq.LoadEnvReq = env
	}

	log.Printf("[%s] stop load test. %+v\n", loadTestKeyReq.LoadTestKey, loadTestReq)
	loadTestManager := managers.NewLoadTestManager()

	err = loadTestManager.Stop(loadTestReq)

	if err != nil {
		log.Printf("Error while execute load test; %v\n", err)
		return fmt.Errorf("service - execute load test error; %w", err)
	}

	return nil
}

func GetLoadTestMetrics(loadTestKey, format string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKey)
	if err != nil {
		return nil, err
	}

	if loadExecutionState.ExecutionStatus == constant.Processing {
		return nil, errors.New("load test is processing")
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)

	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return nil, err
	}

	loadTestManager := managers.NewLoadTestManager()

	result, err := loadTestManager.GetMetrics(loadEnv, loadTestKey, format)
	if err != nil {
		return nil, fmt.Errorf("error on [GetLoadTestMetrics()]; %s", err)
	}
	return result, nil
}

func GetAllLoadExecutionConfig() ([]api.LoadExecutionRes, error) {
	loadExecutionConfigs, err := repository.GetAllLoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	var loadExecutionConfigResponses []api.LoadExecutionRes
	for _, v := range loadExecutionConfigs {
		loadEnvId := fmt.Sprintf("%d", v.LoadEnvID)
		loadEnv, err := repository.GetEnvironment(loadEnvId)
		if err != nil {
			return nil, err
		}
		state, err := repository.GetLoadExecutionState(v.LoadTestKey)
		if err != nil {
			return nil, err
		}
		var load api.LoadEnvRes
		load.LoadEnvId = loadEnv.ID
		load.InstallLocation = loadEnv.InstallLocation
		load.Username = loadEnv.Username
		load.PublicIp = loadEnv.PublicIp
		load.PemKeyPath = loadEnv.PemKeyPath
		load.NsId = loadEnv.NsId
		load.McisId = loadEnv.McisId
		load.VmId = loadEnv.VmId

		loadExecutionHttps := make([]api.LoadExecutionHttpRes, 0)

		for _, v := range v.LoadExecutionHttps {
			loadHttp := api.LoadExecutionHttpRes{
				LoadExecutionHttpId: v.LoadExecutionConfigID,
				Method:              v.Method,
				Protocol:            v.Protocol,
				Hostname:            v.Hostname,
				Port:                v.Port,
				Path:                v.Path,
				BodyData:            v.BodyData,
			}
			loadExecutionHttps = append(loadExecutionHttps, loadHttp)
		}
		res := api.LoadExecutionRes{
			LoadExecutionConfigId: v.ID,
			LoadTestKey:           v.LoadTestKey,
			VirtualUsers:          v.VirtualUsers,
			Duration:              v.Duration,
			RampUpTime:            v.RampUpTime,
			RampUpSteps:           v.RampUpSteps,
			LoadEnv:               load,
			LoadExecutionHttp:     loadExecutionHttps,
			TestName:              v.TestName,
			LoadExecutionState: api.LoadExecutionStateRes{
				LoadExecutionStateId: state.ID,
				LoadTestKey:          state.LoadTestKey,
				ExecutionStatus:      state.ExecutionStatus,
				StartAt:              state.StartAt,
				EndAt:                state.EndAt,
				TotalSec:             state.TotalSec,
			},
		}

		loadExecutionConfigResponses = append(loadExecutionConfigResponses, res)
	}

	return loadExecutionConfigResponses, nil
}

func GetLoadExecutionConfig(loadTestKey string) (api.LoadExecutionRes, error) {
	loadExecutionConfig, err := repository.GetLoadExecutionConfig(loadTestKey)
	if err != nil {
		return api.LoadExecutionRes{}, err
	}

	loadEnvId := fmt.Sprintf("%d", loadExecutionConfig.LoadEnvID)
	loadEnv, err := repository.GetEnvironment(loadEnvId)
	if err != nil {
		return api.LoadExecutionRes{}, err
	}
	state, err := repository.GetLoadExecutionState(loadTestKey)
	if err != nil {
		return api.LoadExecutionRes{}, err
	}
	var load api.LoadEnvRes
	load.LoadEnvId = loadEnv.ID
	load.InstallLocation = loadEnv.InstallLocation
	load.Username = loadEnv.Username
	load.PublicIp = loadEnv.PublicIp
	load.PemKeyPath = loadEnv.PemKeyPath
	load.NsId = loadEnv.NsId
	load.McisId = loadEnv.McisId
	load.VmId = loadEnv.VmId

	loadExecutionHttps := make([]api.LoadExecutionHttpRes, 0)

	for _, v := range loadExecutionConfig.LoadExecutionHttps {
		loadHttp := api.LoadExecutionHttpRes{
			LoadExecutionHttpId: v.LoadExecutionConfigID,
			Method:              v.Method,
			Protocol:            v.Protocol,
			Hostname:            v.Hostname,
			Port:                v.Port,
			Path:                v.Path,
			BodyData:            v.BodyData,
		}
		loadExecutionHttps = append(loadExecutionHttps, loadHttp)
	}

	res := api.LoadExecutionRes{
		LoadExecutionConfigId: loadExecutionConfig.ID,
		LoadTestKey:           loadExecutionConfig.LoadTestKey,
		VirtualUsers:          loadExecutionConfig.VirtualUsers,
		Duration:              loadExecutionConfig.Duration,
		RampUpTime:            loadExecutionConfig.RampUpTime,
		RampUpSteps:           loadExecutionConfig.RampUpSteps,
		TestName:              loadExecutionConfig.TestName,
		LoadEnv:               load,
		LoadExecutionHttp:     loadExecutionHttps,
		LoadExecutionState: api.LoadExecutionStateRes{
			LoadExecutionStateId: state.ID,
			LoadTestKey:          state.LoadTestKey,
			ExecutionStatus:      state.ExecutionStatus,
			StartAt:              state.StartAt,
			EndAt:                state.EndAt,
		},
	}

	return res, nil
}

func GetAllLoadExecutionState() (interface{}, error) {
	loadExecutionStates, err := repository.GetAllLoadExecutionState()

	if err != nil {
		return nil, err
	}

	responseStates := make([]api.LoadExecutionStateRes, 0)
	for _, v := range loadExecutionStates {
		loadExecutionStateRes := api.LoadExecutionStateRes{
			LoadExecutionStateId: v.ID,
			LoadTestKey:          v.LoadTestKey,
			ExecutionStatus:      v.ExecutionStatus,
			StartAt:              v.StartAt,
			EndAt:                v.EndAt,
			TotalSec:             v.TotalSec,
		}

		responseStates = append(responseStates, loadExecutionStateRes)
	}

	return responseStates, nil
}

func GetLoadExecutionState(loadTestKey string) (interface{}, error) {
	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKey)

	if err != nil {
		return nil, err
	}

	loadExecutionStateRes := api.LoadExecutionStateRes{
		LoadExecutionStateId: loadExecutionState.ID,
		LoadTestKey:          loadExecutionState.LoadTestKey,
		ExecutionStatus:      loadExecutionState.ExecutionStatus,
		StartAt:              loadExecutionState.StartAt,
		EndAt:                loadExecutionState.EndAt,
		TotalSec:             loadExecutionState.TotalSec,
	}

	return loadExecutionStateRes, nil
}
