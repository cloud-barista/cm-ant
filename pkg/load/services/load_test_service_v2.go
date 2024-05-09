package services

import (
	"context"
	"errors"
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/outbound/tumblebug"
	"log"
	"os"
	"sync"
	"time"
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

func InstallLoadTesterV2(loadTesterReq *api.LoadTesterReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	vm, err := tumblebug.GetVmWithContext(ctx, loadTesterReq.NsId, loadTesterReq.McisId, loadTesterReq.VmId)

	if err != nil {
		return err
	}

	connectionId := vm.ConnectionName
	nsId := loadTesterReq.NsId
	vNetId := vm.VNetID
	subnetId := vm.SubnetID
	username := vm.VMUserAccount
	cspRegion := vm.Region.Region
	cspImageId, ok := regionImageMap[cspRegion]
	sgId := "ant-load-test-sg"
	sshId := "ant-load-test-ssh"
	imageId := "ant-load-test-image"
	specId := "ant-load-test-spec"
	cspSpecName := "t3.small"
	vmId := "ant-load-test-vm"
	mcisId := "ant-load-test-mcis"

	log.Println(connectionId, nsId, vNetId, subnetId, username, cspRegion, cspImageId, sgId, sshId, imageId, specId, cspSpecName, vmId, mcisId)

	if !ok {
		return errors.New("region base ubuntu 22.04 lts image doesn't exist")
	}

	var wg sync.WaitGroup
	goroutine := 4
	errChan := make(chan error, goroutine)

	wg.Add(1)
	go func() {
		defer wg.Done()

		tc, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()

		err2 := tumblebug.GetSecurityGroupWithContext(tc, nsId, sgId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}

		sg := tumblebug.SecurityGroupReq{
			Name:           sgId,
			ConnectionName: connectionId,
			VNetID:         vNetId,
			Description:    "Default Security Group for Ant load test",
			FirewallRules: []tumblebug.FirewallRuleReq{
				{FromPort: "22", ToPort: "22", IPProtocol: "tcp", Direction: "inbound", CIDR: "0.0.0.0/0"},
			},
		}
		_, err2 = tumblebug.CreateSecurityGroupWithContext(tc, nsId, sg)
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

		err2 := tumblebug.GetSecureShellWithContext(tc, nsId, sshId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}

		ssh := tumblebug.SecureShellReq{
			ConnectionName: connectionId,
			Name:           sshId,
			Username:       username,
			Description:    "Default secure shell key for Ant load test",
		}
		_, err2 = tumblebug.CreateSecureShellWithContext(ctx, nsId, ssh)
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

		err2 := tumblebug.GetImageWithContext(tc, nsId, imageId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}
		// TODO add dynamic spec integration in advanced version
		image := tumblebug.ImageReq{
			ConnectionName: connectionId,
			Name:           imageId,
			CspImageId:     cspImageId,
			Description:    "Default machine image for Ant load test",
		}
		_, err2 = tumblebug.CreateImageWithContext(ctx, nsId, image)
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

		err2 := tumblebug.GetSpecWithContext(tc, nsId, specId)
		if err2 != nil && !errors.Is(err2, tumblebug.ResourcesNotFound) {
			errChan <- err2
			return
		} else if err2 == nil {
			return
		}

		// TODO add dynamic spec integration in advanced version
		spec := tumblebug.SpecReq{
			ConnectionName: connectionId,
			Name:           specId,
			CspSpecName:    cspSpecName,
			Description:    "Default spec for Ant load test",
		}
		_, err2 = tumblebug.CreateSpecWithContext(ctx, nsId, spec)
		if err2 != nil {
			errChan <- err2
			return
		}
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) != 0 {
		err = <-errChan
		return err
	}

	// TODO check mcis and condition
	mcis, err := tumblebug.GetMcisWithContext(ctx, nsId, mcisId)
	if err != nil && !errors.Is(err, tumblebug.ResourcesNotFound) {
		return err
	} else if err == nil {
		// mcis 상태 확인 후 jmeter 설치
		if mcis.IsRunning(vmId) {
			scriptPath := configuration.JoinRootPathWith("/script/install-jmeter.sh")

			installScript, err := os.ReadFile(scriptPath)
			if err != nil {
				return err
			}

			commandReq := tumblebug.SendCommandReq{
				Command:  []string{string(installScript)},
				UserName: username,
			}

			stdout, err := tumblebug.CommandToVmWithContext(ctx, nsId, mcisId, mcis.VmId(), commandReq)

			if err != nil {
				log.Println(stdout)
				return err
			}
			log.Println(stdout)

			return nil
		}

		return errors.New("mcis condition is not correct")

	}

	mcisReq := tumblebug.McisReq{
		Name:            mcisId,
		Description:     "Default mcis for Ant load test",
		InstallMonAgent: "no",
		Label:           "ANT",
		SystemLabel:     "ANT",
		Vm: []tumblebug.VmReq{
			{
				SubGroupSize:     "1",
				Name:             vmId,
				ImageId:          imageId,
				VmUserAccount:    "cb-user",
				ConnectionName:   connectionId,
				SshKeyId:         sshId,
				SpecId:           specId,
				SecurityGroupIds: []string{sgId},
				VNetId:           vNetId,
				SubnetId:         subnetId,
				Description:      "Default vm for Ant load test",
				VmUserPassword:   "",
				RootDiskType:     "default",
				RootDiskSize:     "default",
			},
		},
	}

	createdMcis, err := tumblebug.CreateMcisWithContext(ctx, nsId, mcisReq)

	if err != nil {
		return err
	}

	log.Println("******************created*******************\n", createdMcis)

	time.Sleep(5 * time.Second)

	scriptPath := configuration.JoinRootPathWith("/script/install-jmeter.sh")

	installScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	commandReq := tumblebug.SendCommandReq{
		Command:  []string{string(installScript)},
		UserName: username,
	}

	stdout, err := tumblebug.CommandToVmWithContext(ctx, nsId, mcisId, createdMcis.VmId(), commandReq)

	if err != nil {
		log.Println(stdout)
		return err
	}
	log.Println(stdout)

	return nil
}

//func UninstallLoadTester(loadEnvId string) error {
//	loadTestManager := managers.NewLoadTestManager()
//
//	var loadEnvReq api.LoadEnvReq
//	if loadEnvId != "" {
//		loadEnv, err := repository.GetEnvironment(loadEnvId)
//		if err != nil {
//			return err
//		}
//
//		loadEnvReq.InstallLocation = (*loadEnv).InstallLocation
//		loadEnvReq.RemoteConnectionType = (*loadEnv).RemoteConnectionType
//		loadEnvReq.Username = (*loadEnv).Username
//		loadEnvReq.PublicIp = (*loadEnv).PublicIp
//		loadEnvReq.Cert = (*loadEnv).Cert
//		loadEnvReq.NsId = (*loadEnv).NsId
//		loadEnvReq.McisId = (*loadEnv).McisId
//	}
//
//	if err := loadTestManager.Uninstall(&loadEnvReq); err != nil {
//		return fmt.Errorf("failed to uninstall load tester: %w", err)
//	}
//
//	//err := repository.DeleteLoadTestInstallEnv(loadEnvId)
//	//if err != nil {
//	//	return fmt.Errorf("failed to delete load test installation environment: %w", err)
//	//}
//	log.Println("load test environment is successfully deleted")
//
//	return nil
//}
//
//func prepareEnvironment(loadTestReq *api.LoadExecutionConfigReq) error {
//	if loadTestReq.EnvId == "" {
//		return nil
//	}
//
//	loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
//	if err != nil {
//		return fmt.Errorf("failed to get environment: %w", err)
//	}
//
//	if loadEnv != nil && loadTestReq.LoadEnvReq.InstallLocation == "" {
//		loadTestReq.LoadEnvReq = convertToLoadEnvReq(loadEnv)
//	}
//
//	return nil
//}
//
//func convertToLoadEnvReq(loadEnv *model.LoadEnv) api.LoadEnvReq {
//	return api.LoadEnvReq{
//		InstallLocation:      loadEnv.InstallLocation,
//		RemoteConnectionType: loadEnv.RemoteConnectionType,
//		Username:             loadEnv.Username,
//		PublicIp:             loadEnv.PublicIp,
//		Cert:                 loadEnv.Cert,
//		NsId:                 loadEnv.NsId,
//		McisId:               loadEnv.McisId,
//	}
//}
//
//func runLoadTest(loadTestManager managers.LoadTestManager, loadTestReq *api.LoadExecutionConfigReq, loadTestKey string) {
//	log.Printf("[%s] start load test", loadTestKey)
//	if err := loadTestManager.Run(loadTestReq); err != nil {
//		log.Printf("Error during load test: %v", err)
//		if updateErr := repository.UpdateLoadExecutionState(loadTestKey, constant.Failed); updateErr != nil {
//			log.Println(updateErr)
//		}
//	} else {
//		log.Printf("load test complete!")
//
//		if updateErr := repository.UpdateLoadExecutionState(loadTestKey, constant.Success); updateErr != nil {
//			log.Println(updateErr)
//		}
//	}
//}
//
//func ExecuteLoadTest(loadTestReq *api.LoadExecutionConfigReq) (string, error) {
//	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
//	loadTestReq.LoadTestKey = loadTestKey
//
//	// check env
//	if err := prepareEnvironment(loadTestReq); err != nil {
//		return "", err
//	}
//
//	// installation jmeter
//	envId, err := InstallLoadTester(&loadTestReq.LoadEnvReq)
//	if err != nil {
//		return "", err
//	}
//
//	loadTestReq.EnvId = fmt.Sprintf("%d", envId)
//
//	log.Printf("[%s] start load test", loadTestKey)
//	loadTestManager := managers.NewLoadTestManager()
//
//	go runLoadTest(loadTestManager, loadTestReq, loadTestKey)
//
//	_, err = repository.SaveLoadTestExecution(loadTestReq)
//	if err != nil {
//		return "", err
//	}
//
//	return loadTestKey, nil
//}
//
//func StopLoadTest(loadTestKeyReq api.LoadTestKeyReq) error {
//	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKeyReq.LoadTestKey)
//
//	if err != nil {
//		return err
//	}
//
//	if loadExecutionState.IsFinished() {
//		return fmt.Errorf("load test is already finished")
//	}
//
//	loadTestReq := api.LoadExecutionConfigReq{
//		LoadTestKey: loadTestKeyReq.LoadTestKey,
//		EnvId:       fmt.Sprintf("%d", loadExecutionState.LoadEnvID),
//	}
//
//	var env api.LoadEnvReq
//	if loadTestReq.EnvId != "" {
//		loadEnv, err := repository.GetEnvironment(loadTestReq.EnvId)
//		if err != nil {
//			return err
//		}
//
//		env.InstallLocation = (*loadEnv).InstallLocation
//		env.RemoteConnectionType = (*loadEnv).RemoteConnectionType
//		env.Username = (*loadEnv).Username
//		env.PublicIp = (*loadEnv).PublicIp
//		env.Cert = (*loadEnv).Cert
//		env.NsId = (*loadEnv).NsId
//		env.McisId = (*loadEnv).McisId
//
//		loadTestReq.LoadEnvReq = env
//	}
//
//	log.Printf("[%s] stop load test. %+v\n", loadTestKeyReq.LoadTestKey, loadTestReq)
//	loadTestManager := managers.NewLoadTestManager()
//
//	err = loadTestManager.Stop(loadTestReq)
//
//	if err != nil {
//		log.Printf("Error while execute load test; %v\n", err)
//		return fmt.Errorf("service - execute load test error; %w", err)
//	}
//
//	return nil
//}
//
//func GetLoadTestResult(testKey, format string) (interface{}, error) {
//	loadExecutionState, err := repository.GetLoadExecutionState(testKey)
//	if err != nil {
//		return nil, err
//	}
//
//	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)
//
//	loadEnv, err := repository.GetEnvironment(loadEnvId)
//	if err != nil {
//		return nil, err
//	}
//
//	loadTestManager := managers.NewLoadTestManager()
//
//	result, err := loadTestManager.GetResult(loadEnv, testKey, format)
//	if err != nil {
//		return nil, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
//	}
//	return result, nil
//}
//
//func GetLoadTestMetrics(testKey, format string) (interface{}, error) {
//	loadExecutionState, err := repository.GetLoadExecutionState(testKey)
//	if err != nil {
//		return nil, err
//	}
//
//	loadEnvId := fmt.Sprintf("%d", loadExecutionState.LoadEnvID)
//
//	loadEnv, err := repository.GetEnvironment(loadEnvId)
//	if err != nil {
//		return nil, err
//	}
//
//	loadTestManager := managers.NewLoadTestManager()
//
//	result, err := loadTestManager.GetMetrics(loadEnv, testKey, format)
//	if err != nil {
//		return nil, fmt.Errorf("error on [InstallLoadGenerator()]; %s", err)
//	}
//	return result, nil
//}
//
//func GetAllLoadExecutionConfig() ([]api.LoadExecutionRes, error) {
//	loadExecutionConfigs, err := repository.GetAllLoadExecutionConfig()
//	if err != nil {
//		return nil, err
//	}
//
//	var loadExecutionConfigResponses []api.LoadExecutionRes
//	for _, v := range loadExecutionConfigs {
//		loadEnvId := fmt.Sprintf("%d", v.LoadEnvID)
//		loadEnv, err := repository.GetEnvironment(loadEnvId)
//		if err != nil {
//			return nil, err
//		}
//		state, err := repository.GetLoadExecutionState(v.LoadTestKey)
//		if err != nil {
//			return nil, err
//		}
//		var load api.LoadEnvRes
//		load.LoadEnvId = loadEnv.ID
//		load.InstallLocation = loadEnv.InstallLocation
//		load.RemoteConnectionType = loadEnv.RemoteConnectionType
//		load.Username = loadEnv.Username
//		load.PublicIp = loadEnv.PublicIp
//		load.Cert = loadEnv.Cert
//		load.NsId = loadEnv.NsId
//		load.McisId = loadEnv.McisId
//
//		loadExecutionHttps := make([]api.LoadExecutionHttpRes, 0)
//
//		for _, v := range v.LoadExecutionHttps {
//			loadHttp := api.LoadExecutionHttpRes{
//				LoadExecutionHttpId: v.LoadExecutionConfigID,
//				Method:              v.Method,
//				Protocol:            v.Protocol,
//				Hostname:            v.Hostname,
//				Port:                v.Port,
//				Path:                v.Path,
//				BodyData:            v.BodyData,
//			}
//			loadExecutionHttps = append(loadExecutionHttps, loadHttp)
//		}
//		res := api.LoadExecutionRes{
//			LoadExecutionConfigId: v.ID,
//			LoadTestKey:           v.LoadTestKey,
//			VirtualUsers:          v.VirtualUsers,
//			Duration:              v.Duration,
//			RampUpTime:            v.RampUpTime,
//			RampUpSteps:           v.RampUpSteps,
//			LoadEnv:               load,
//			LoadExecutionHttp:     loadExecutionHttps,
//			TestName:              v.TestName,
//			LoadExecutionState: api.LoadExecutionStateRes{
//				LoadExecutionStateId: state.ID,
//				LoadTestKey:          state.LoadTestKey,
//				ExecutionStatus:      state.ExecutionStatus,
//				StartAt:              state.StartAt,
//				EndAt:                state.EndAt,
//			},
//		}
//
//		loadExecutionConfigResponses = append(loadExecutionConfigResponses, res)
//	}
//
//	return loadExecutionConfigResponses, nil
//}
//
//func GetLoadExecutionConfig(loadTestKey string) (api.LoadExecutionRes, error) {
//	loadExecutionConfig, err := repository.GetLoadExecutionConfig(loadTestKey)
//	if err != nil {
//		return api.LoadExecutionRes{}, err
//	}
//
//	loadEnvId := fmt.Sprintf("%d", loadExecutionConfig.LoadEnvID)
//	loadEnv, err := repository.GetEnvironment(loadEnvId)
//	if err != nil {
//		return api.LoadExecutionRes{}, err
//	}
//	state, err := repository.GetLoadExecutionState(loadTestKey)
//	if err != nil {
//		return api.LoadExecutionRes{}, err
//	}
//	var load api.LoadEnvRes
//	load.LoadEnvId = loadEnv.ID
//	load.InstallLocation = loadEnv.InstallLocation
//	load.RemoteConnectionType = loadEnv.RemoteConnectionType
//	load.Username = loadEnv.Username
//	load.PublicIp = loadEnv.PublicIp
//	load.Cert = loadEnv.Cert
//	load.NsId = loadEnv.NsId
//	load.McisId = loadEnv.McisId
//
//	loadExecutionHttps := make([]api.LoadExecutionHttpRes, 0)
//
//	for _, v := range loadExecutionConfig.LoadExecutionHttps {
//		loadHttp := api.LoadExecutionHttpRes{
//			LoadExecutionHttpId: v.LoadExecutionConfigID,
//			Method:              v.Method,
//			Protocol:            v.Protocol,
//			Hostname:            v.Hostname,
//			Port:                v.Port,
//			Path:                v.Path,
//			BodyData:            v.BodyData,
//		}
//		loadExecutionHttps = append(loadExecutionHttps, loadHttp)
//	}
//
//	res := api.LoadExecutionRes{
//		LoadExecutionConfigId: loadExecutionConfig.ID,
//		LoadTestKey:           loadExecutionConfig.LoadTestKey,
//		VirtualUsers:          loadExecutionConfig.VirtualUsers,
//		Duration:              loadExecutionConfig.Duration,
//		RampUpTime:            loadExecutionConfig.RampUpTime,
//		RampUpSteps:           loadExecutionConfig.RampUpSteps,
//		TestName:              loadExecutionConfig.TestName,
//		LoadEnv:               load,
//		LoadExecutionHttp:     loadExecutionHttps,
//		LoadExecutionState: api.LoadExecutionStateRes{
//			LoadExecutionStateId: state.ID,
//			LoadTestKey:          state.LoadTestKey,
//			ExecutionStatus:      state.ExecutionStatus,
//			StartAt:              state.StartAt,
//			EndAt:                state.EndAt,
//		},
//	}
//
//	return res, nil
//}
//
//func GetAllLoadExecutionState() (interface{}, error) {
//	loadExecutionStates, err := repository.GetAllLoadExecutionState()
//
//	if err != nil {
//		return nil, err
//	}
//
//	responseStates := make([]api.LoadExecutionStateRes, 0)
//	for _, v := range loadExecutionStates {
//		loadExecutionStateRes := api.LoadExecutionStateRes{
//			LoadExecutionStateId: v.ID,
//			LoadTestKey:          v.LoadTestKey,
//			ExecutionStatus:      v.ExecutionStatus,
//			StartAt:              v.StartAt,
//			EndAt:                v.EndAt,
//		}
//
//		responseStates = append(responseStates, loadExecutionStateRes)
//	}
//
//	return responseStates, nil
//}
//
//func GetLoadExecutionState(loadTestKey string) (interface{}, error) {
//	loadExecutionState, err := repository.GetLoadExecutionState(loadTestKey)
//
//	if err != nil {
//		return nil, err
//	}
//
//	loadExecutionStateRes := api.LoadExecutionStateRes{
//		LoadExecutionStateId: loadExecutionState.ID,
//		LoadTestKey:          loadExecutionState.LoadTestKey,
//		ExecutionStatus:      loadExecutionState.ExecutionStatus,
//		StartAt:              loadExecutionState.StartAt,
//		EndAt:                loadExecutionState.EndAt,
//	}
//
//	return loadExecutionStateRes, nil
//}
