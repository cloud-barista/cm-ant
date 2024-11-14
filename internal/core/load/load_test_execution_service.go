package load

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
)

// RunLoadTest initiates the load test and performs necessary initializations.
// Generates a load test key, installs the load generator or retrieves existing installation information,
// saves the load test execution state, and then asynchronously runs the load test.
func (l *LoadService) RunLoadTest(param RunLoadTestParam) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	loadTestKey := utils.CreateUniqIdBaseOnUnixTime()
	param.LoadTestKey = loadTestKey
	log.Info().Msgf("Starting load test with key: %s", loadTestKey)

	duration, err := strconv.Atoi(param.Duration)
	if err != nil {
		log.Error().Msgf("error while type convert; %s", err.Error())
		return loadTestKey, err
	}

	rampUpTime, err := strconv.Atoi(param.RampUpTime)

	if err != nil {
		return loadTestKey, err
	}

	totalExecutionSecond := uint64(duration + rampUpTime)
	startAt := time.Now()
	e := startAt.Add(time.Duration(totalExecutionSecond) * time.Second)

	stateArg := LoadTestExecutionState{
		LoadTestKey:                 loadTestKey,
		ExecutionStatus:             constant.OnProcessing,
		StartAt:                     startAt,
		ExpectedFinishAt:            e,
		TotalExpectedExcutionSecond: totalExecutionSecond,
		NsId:                        param.NsId,
		MciId:                       param.MciId,
		VmId:                        param.VmId,
		WithMetrics:                 param.CollectAdditionalSystemMetrics,
	}

	err = l.loadRepo.InsertLoadTestExecutionStateTx(ctx, &stateArg)
	if err != nil {
		log.Error().Msgf("Error saving initial load test execution state: %v", err)
		return "", err
	}
	log.Info().Msgf("Initial load test execution state saved for key: %s", loadTestKey)

	go l.processLoadTestAsync(param, &stateArg)

	return loadTestKey, nil

}

func (l *LoadService) processLoadTestAsync(param RunLoadTestParam, loadTestExecutionState *LoadTestExecutionState) {

	var globalErr error

	defer func() {
		if globalErr != nil {
			_ = l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
		}
	}()

	failed := func(logMsg string, occuredError error) {
		log.Error().Msg(logMsg)
		loadTestExecutionState.ExecutionStatus = constant.TestFailed
		loadTestExecutionState.FailureMessage = logMsg
		globalErr = occuredError
		finishAt := time.Now()
		loadTestExecutionState.FinishAt = &finishAt
	}

	if param.LoadGeneratorInstallInfoId == uint(0) {
		log.Info().Msgf("No LoadGeneratorInstallInfoId provided, installing load generator...")

		result, err := l.InstallLoadGenerator(param.InstallLoadGenerator)
		if err != nil {
			failed(fmt.Sprintf("Error installing load generator: %v", err), err)
			return
		}

		param.LoadGeneratorInstallInfoId = result.ID
		log.Info().Msgf("Load generator installed with ID: %d", result.ID)
	}

	loadGeneratorInstallInfo, err := l.loadRepo.GetValidLoadGeneratorInstallInfoByIdTx(context.Background(), param.LoadGeneratorInstallInfoId)
	if err != nil {
		failed(fmt.Sprintf("Error installing load generator: %v", err), err)
		return
	}

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

	loadTestExecutionInfoParam := LoadTestExecutionInfo{
		LoadTestKey:  param.LoadTestKey,
		TestName:     param.TestName,
		VirtualUsers: param.VirtualUsers,
		Duration:     param.Duration,
		RampUpTime:   param.RampUpTime,
		RampUpSteps:  param.RampUpSteps,

		NsId:  param.NsId,
		MciId: param.MciId,
		VmId:  param.VmId,

		AgentInstalled: param.CollectAdditionalSystemMetrics,
		AgentHostname:  param.AgentHostname,

		LoadGeneratorInstallInfoId: loadGeneratorInstallInfo.ID,
		LoadTestExecutionHttpInfos: hs,

		LoadTestExecutionStateId: loadTestExecutionState.ID,
	}

	err = l.loadRepo.SaveForLoadTestExecutionTx(context.Background(), &loadTestExecutionInfoParam)
	if err != nil {
		failed(fmt.Sprintf("Error saving load test execution info: %v", err), err)
		return
	}

	loadTestExecutionState.GeneratorInstallInfoId = loadGeneratorInstallInfo.ID
	loadTestExecutionState.TestExecutionInfoId = loadTestExecutionInfoParam.ID

	err = l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)

	if err != nil {
		failed(fmt.Sprintf("Error while update load test execution state to save load test execution info id and load generator install info id; %v", err), err)
	}

	if param.CollectAdditionalSystemMetrics {
		if strings.TrimSpace(param.AgentHostname) == "" {
			mci, err := l.tumblebugClient.GetMciWithContext(context.Background(), param.NsId, param.MciId)

			if err != nil {
				failed(fmt.Sprintf("unexpected error occurred while fetching mci for install metrics agent; %s", err), err)
				return
			}

			if len(mci.Vm) == 0 {
				failed("mci vm's length is zero", errors.New("mci vm's length is zero"))
				return
			}

			if len(mci.Vm) == 1 {
				param.AgentHostname = mci.Vm[0].PublicIP
			} else {
				for _, v := range mci.Vm {
					if v.Id == param.VmId {
						param.AgentHostname = v.PublicIP
					}
				}
			}

			if param.AgentHostname == "" {
				err := errors.New("agent host name afeter get mci from tumblebug must be set to not nil")
				failed(fmt.Sprintf("invalid agent hostname for test %s; %v", param.LoadTestKey, err), err)
				return
			}
		}

		arg := MonitoringAgentInstallationParams{
			NsId:  param.NsId,
			MciId: param.MciId,
			VmIds: []string{param.VmId},
		}

		// install and run the agent for collect metrics
		_, err := l.InstallMonitoringAgent(arg)
		if err != nil {
			failed(fmt.Sprintf("unexpected error occurred while installing monitoring agent; %s", err), err)
			return
		}

		log.Info().Msgf("metrics agent installed successfully for load test; %s %s %s", arg.NsId, arg.MciId, arg.VmIds)
	}

	loadTestDone := make(chan bool)

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

	home, err := os.UserHomeDir()
	if err != nil {
		failed(fmt.Sprintf("user home dir is not valid; %s", err), err)
		return
	}

	dataParam := &fetchDataParam{
		LoadTestDone:                   loadTestDone,
		LoadTestKey:                    param.LoadTestKey,
		InstallLocation:                loadGeneratorInstallInfo.InstallLocation,
		InstallPath:                    loadGeneratorInstallInfo.InstallPath,
		PublicKeyName:                  loadGeneratorInstallInfo.PublicKeyName,
		PrivateKeyName:                 loadGeneratorInstallInfo.PrivateKeyName,
		Username:                       username,
		PublicIp:                       publicIp,
		Port:                           port,
		CollectAdditionalSystemMetrics: param.CollectAdditionalSystemMetrics,
		Home:                           home,
	}

	go l.fetchData(dataParam)

	defer func() {
		loadTestDone <- true
		close(loadTestDone)
		updateErr := l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
		if updateErr != nil {
			failed(fmt.Sprintf("Error updating load test execution state: %v", updateErr), updateErr)
			return
		}

		log.Info().Msgf("successfully done load test for %s", dataParam.LoadTestKey)
	}()

	compileDuration, executionDuration, loadTestErr := l.executeLoadTest(param, &loadGeneratorInstallInfo)

	loadTestExecutionState.CompileDuration = compileDuration
	loadTestExecutionState.ExecutionDuration = executionDuration

	if loadTestErr != nil {
		failed(fmt.Sprintf("Error while load testing: %v", loadTestErr), loadTestErr)
		return
	}

	updateErr := l.updateLoadTestExecution(loadTestExecutionState)
	if updateErr != nil {
		failed(fmt.Sprintf("Error while updating load test execution: %v", updateErr), updateErr)
		return
	}

	loadTestExecutionState.ExecutionStatus = constant.Successed
}

// processLoadTest executes the load test.
// Depending on whether the installation location is local or remote, it creates the test plan and runs test commands.
// Fetches and saves test results from the local or remote system.
func (l *LoadService) processLoadTest(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo, loadTestExecutionState *LoadTestExecutionState) {

	// check the installation of agent for additional metrics
	if param.CollectAdditionalSystemMetrics {
		if strings.TrimSpace(param.AgentHostname) == "" {
			mci, err := l.tumblebugClient.GetMciWithContext(context.Background(), param.NsId, param.MciId)

			if err != nil {
				log.Error().Msgf("unexpected error occurred while fetching mci for install metrics agent")
				return
			}

			if len(mci.Vm) == 0 {
				log.Error().Msgf("unexpected error occurred while fetching mci for install metrics agent")
				return
			}

			if len(mci.Vm) == 1 {
				param.AgentHostname = mci.Vm[0].PublicIP
			} else {
				for _, v := range mci.Vm {
					if v.Id == param.VmId {
						param.AgentHostname = v.PublicIP
					}
				}
			}

			if param.AgentHostname == "" {
				log.Error().Msgf("invalid agent hostname for test %s", param.LoadTestKey)
				return
			}
		}

		arg := MonitoringAgentInstallationParams{
			NsId:  param.NsId,
			MciId: param.MciId,
			VmIds: []string{param.VmId},
		}

		// install and run the agent for collect metrics
		_, err := l.InstallMonitoringAgent(arg)
		if err != nil {
			log.Error().Msgf("unexpected error occurred while fetching mci for ")
			return
		}

		log.Info().Msgf("metrics agent installed successfully for load test; %s %s %s", arg.NsId, arg.MciId, arg.VmIds)
	}

	loadTestDone := make(chan bool)

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

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	dataParam := &fetchDataParam{
		LoadTestDone:                   loadTestDone,
		LoadTestKey:                    param.LoadTestKey,
		InstallLocation:                loadGeneratorInstallInfo.InstallLocation,
		InstallPath:                    loadGeneratorInstallInfo.InstallPath,
		PublicKeyName:                  loadGeneratorInstallInfo.PublicKeyName,
		PrivateKeyName:                 loadGeneratorInstallInfo.PrivateKeyName,
		Username:                       username,
		PublicIp:                       publicIp,
		Port:                           port,
		CollectAdditionalSystemMetrics: param.CollectAdditionalSystemMetrics,
		Home:                           home,
	}

	go l.fetchData(dataParam)

	defer func() {
		loadTestDone <- true
		close(loadTestDone)
		updateErr := l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
		if updateErr != nil {
			log.Error().Msgf("Error updating load test execution state: %v", updateErr)
			return
		}

		log.Info().Msgf("successfully done load test for %s", dataParam.LoadTestKey)
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
		loadTestExecutionState.ExecutionStatus = constant.TestFailed
		loadTestExecutionState.FailureMessage = updateErr.Error()
		finishAt := time.Now()
		loadTestExecutionState.FinishAt = &finishAt
		return
	}

	loadTestExecutionState.ExecutionStatus = constant.Successed
}

func (l *LoadService) executeLoadTest(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo) (string, string, error) {
	installLocation := loadGeneratorInstallInfo.InstallLocation
	loadTestKey := param.LoadTestKey
	loadGeneratorInstallPath := loadGeneratorInstallInfo.InstallPath
	testPlanName := fmt.Sprintf("%s.jmx", loadTestKey)
	resultFileName := fmt.Sprintf("%s_result.csv", loadTestKey)
	loadGeneratorInstallVersion := loadGeneratorInstallInfo.InstallVersion

	log.Info().Msgf("Running load test with key: %s", loadTestKey)
	compileDuration := "0"
	executionDuration := "0"
	start := time.Now()

	if installLocation == constant.Remote {
		log.Info().Msg("Remote execute detected.")
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
		_, err = l.tumblebugClient.CommandToMciWithContext(context.Background(), antNsId, antMciId, commandReq)
		if err != nil {
			return compileDuration, executionDuration, err
		}

		jmeterTestCommand := generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName)

		commandReq = tumblebug.SendCommandReq{
			Command: []string{jmeterTestCommand},
		}

		stdout, err := l.tumblebugClient.CommandToMciWithContext(context.Background(), antNsId, antMciId, commandReq)
		if err != nil {
			return compileDuration, executionDuration, err
		}
		executionDuration = utils.DurationString(start)

		if strings.Contains(stdout, "exited with status 1") {
			return compileDuration, executionDuration, errors.New("jmeter test stopped unexpectedly")
		}

	} else if installLocation == constant.Local {
		log.Info().Msg("Local execute detected.")

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
		log.Error().Msgf("Error updating load test execution info; %v", err)
		return err
	}

	loadTestExecutionState.ExecutionStatus = constant.OnFetching
	err = l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
	if err != nil {
		log.Error().Msgf("Error updating load test execution state; %v", err)
		return err
	}
	return nil
}

// generateJmeterExecutionCmd generates the JMeter execution command.
// Constructs a JMeter command string that includes the test plan path and result file path.
func generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName string) string {
	log.Info().Msgf("Generating JMeter execution command for test plan: %s, result file: %s", testPlanName, resultFileName)

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", loadGeneratorInstallPath, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", loadGeneratorInstallPath, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", loadGeneratorInstallPath, loadGeneratorInstallVersion))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	builder.WriteString(fmt.Sprintf(" && sudo rm %s", testPath))
	log.Info().Msgf("JMeter execution command generated: %s", builder.String())
	return builder.String()
}

func (l *LoadService) StopLoadTest(param StopLoadTestParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var arg GetLoadTestExecutionStateParam

	if param.LoadTestKey == "" {
		arg = GetLoadTestExecutionStateParam{
			NsId:  param.NsId,
			MciId: param.MciId,
			VmId:  param.VmId,
		}

	} else {
		arg = GetLoadTestExecutionStateParam{
			LoadTestKey: param.LoadTestKey,
		}
	}

	state, err := l.loadRepo.GetLoadTestExecutionStateTx(ctx, arg)

	if err != nil {
		return fmt.Errorf("error occurred while retrieve load test execution state: %w", err)
	}

	if state.ExecutionStatus == constant.Successed {
		return errors.New("load test is already completed")
	}

	installInfo, err := l.loadRepo.GetValidLoadGeneratorInstallInfoByIdTx(ctx, state.GeneratorInstallInfoId)

	if err != nil {
		return fmt.Errorf("error occurred while retrieve load install info: %w", err)
	}

	killCmd := killCmdGen(param.LoadTestKey)

	if installInfo.InstallLocation == constant.Remote {

		commandReq := tumblebug.SendCommandReq{
			Command: []string{killCmd},
		}
		_, err := l.tumblebugClient.CommandToMciWithContext(ctx, antNsId, antMciId, commandReq)

		if err != nil {
			return err
		}

	} else if installInfo.InstallLocation == constant.Local {
		err := utils.InlineCmd(killCmd)

		if err != nil {
			log.Error().Msg(err.Error())
			return err
		}
	}

	return nil

}

func killCmdGen(loadTestKey string) string {
	grepRegex := fmt.Sprintf("'\\/bin\\/ApacheJMeter\\.jar.*%s'", loadTestKey)
	log.Info().Msgf("Generating kill command for load test key: %s", loadTestKey)
	return fmt.Sprintf("kill -15 $(ps -ef | grep -E %s | awk '{print $2}')", grepRegex)
}
