package load

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
)

// getResourceNames returns resource names from config
func getResourceNames() (nsId, mciId, vmName, sshKeyBase string) {
	nsId = config.AppConfig.Load.DefaultResourceName.Namespace
	mciId = config.AppConfig.Load.DefaultResourceName.Mci
	vmName = config.AppConfig.Load.DefaultResourceName.Vm
	sshKeyBase = config.AppConfig.Load.DefaultResourceName.SshKey
	return
}

// stepSeq fixes the display order of the execution steps (FR-MA2-PERF-007-08).
var stepSeq = map[constant.ExecutionStep]int{
	constant.StepGeneratorInstall: 1,
	constant.StepAgentInstall:     2,
	constant.StepJmxPrepare:       3,
	constant.StepJmeterRun:        4,
	constant.StepResultFetch:      5,
}

// stepRecorder persists per-step progress of a running load test so the web console can show
// detailed, real-time status with timing and error messages (FR-MA2-PERF-007-08). A nil
// recorder (or one without a persisted state) is a no-op, so callers that have no execution
// state can pass nil safely.
type stepRecorder struct {
	l     *LoadService
	state *LoadTestExecutionState
}

func (l *LoadService) newStepRecorder(state *LoadTestExecutionState) *stepRecorder {
	return &stepRecorder{l: l, state: state}
}

func (s *stepRecorder) upsert(step *LoadTestExecutionStep) {
	if s == nil || s.state == nil || s.state.ID == 0 {
		return
	}
	step.LoadTestExecutionStateId = s.state.ID
	step.LoadTestKey = s.state.LoadTestKey
	step.Seq = stepSeq[step.Name]
	if err := s.l.loadRepo.UpsertLoadTestExecutionStepTx(context.Background(), step); err != nil {
		log.Warn().Msgf("failed to record execution step %s; %v", step.Name, err)
	}
}

// seed creates the full pipeline as pending so the web can render every step upfront.
func (s *stepRecorder) seed() {
	for name := range stepSeq {
		s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepPending})
	}
}

func (s *stepRecorder) begin(name constant.ExecutionStep, message string) {
	now := time.Now()
	s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepRunning, StartAt: &now, Message: message})
}

// progress updates a running step with a retry count and diagnostic detail
// (e.g. "Installing JMeter (retry 1)").
func (s *stepRecorder) progress(name constant.ExecutionStep, attempt int, message, detail string) {
	s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepRunning, Attempt: attempt, Message: message, Detail: detail})
}

func (s *stepRecorder) ok(name constant.ExecutionStep, message string) {
	now := time.Now()
	s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepOk, FinishAt: &now, Message: message})
}

func (s *stepRecorder) fail(name constant.ExecutionStep, message, detail string) {
	now := time.Now()
	s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepFailed, FinishAt: &now, Message: message, Detail: detail})
}

func (s *stepRecorder) skip(name constant.ExecutionStep, message string) {
	s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepSkipped, Message: message})
}

// generatorRunMu serializes load test runs. The load generator is a shared singleton
// (one ant-default MCI reused across runs), so concurrent runs would contend on the same
// VM (JMeter CPU contention, install/reset races). Runs are therefore serialized and a
// concurrent request is rejected rather than corrupting the shared generator (BAR-1414).
var generatorRunMu sync.Mutex

// RunLoadTest initiates the load test and performs necessary initializations.
// Generates a load test key, installs the load generator or retrieves existing installation information,
// saves the load test execution state, and then asynchronously runs the load test.
func (l *LoadService) RunLoadTest(param RunLoadTestParam) (string, error) {
	timeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.CommandExecution)
	if err != nil {
		log.Warn().Msgf("Failed to parse commandExecution timeout, using default 50 minutes: %v", err)
		timeout = 50 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// BAR-1414: only one load test may run at a time on the shared generator.
	if !generatorRunMu.TryLock() {
		return "", errors.New("a load test is already running; the shared load generator supports one run at a time")
	}
	locked := true
	defer func() {
		if locked {
			generatorRunMu.Unlock()
		}
	}()

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
		InfraId:                     param.InfraId,
		NodeId:                      param.NodeId,
		WithMetrics:                 param.CollectAdditionalSystemMetrics,
	}

	err = l.loadRepo.InsertLoadTestExecutionStateTx(ctx, &stateArg)
	if err != nil {
		log.Error().Msgf("Error saving initial load test execution state: %v", err)
		return "", err
	}
	log.Info().Msgf("Initial load test execution state saved for key: %s", loadTestKey)

	locked = false // ownership transferred to the async run; it releases the lock when done
	go func() {
		defer generatorRunMu.Unlock()
		l.processLoadTestAsync(param, &stateArg)
	}()

	return loadTestKey, nil

}

func (l *LoadService) processLoadTestAsync(param RunLoadTestParam, loadTestExecutionState *LoadTestExecutionState) {

	var globalErr error

	defer func() {
		if globalErr != nil {
			_ = l.loadRepo.UpdateLoadTestExecutionStateTx(context.Background(), loadTestExecutionState)
		}
	}()

	// FR-MA2-PERF-007-08: record per-step progress for the web console.
	rec := l.newStepRecorder(loadTestExecutionState)
	rec.seed()

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
		rec.begin(constant.StepGeneratorInstall, "Installing load generator")

		// ✅ VM 정보를 부하 발생기 설치 파라미터에 추가
		installParam := param.InstallLoadGenerator
		installParam.NsId = param.NsId
		installParam.InfraId = param.InfraId
		installParam.NodeId = param.NodeId

		result, err := l.InstallLoadGenerator(installParam, func(attempt int, message, detail string) {
			rec.progress(constant.StepGeneratorInstall, attempt, message, detail)
		})
		if err != nil {
			rec.fail(constant.StepGeneratorInstall, "Load generator install failed", err.Error())
			failed(fmt.Sprintf("Error installing load generator: %v", err), err)
			return
		}

		rec.ok(constant.StepGeneratorInstall, "Load generator ready")
		param.LoadGeneratorInstallInfoId = result.ID
		log.Info().Msgf("Load generator installed with ID: %d", result.ID)
	} else {
		rec.ok(constant.StepGeneratorInstall, "Reusing existing generator")
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

		NsId:    param.NsId,
		InfraId: param.InfraId,
		NodeId:  param.NodeId,

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
		rec.begin(constant.StepAgentInstall, "Installing monitoring agent")
		if strings.TrimSpace(param.AgentHostname) == "" {
			mci, err := l.tumblebugClient.GetMciWithContext(context.Background(), param.NsId, param.InfraId)

			if err != nil {
				rec.fail(constant.StepAgentInstall, "Monitoring agent install failed", fmt.Sprintf("failed to look up target MCI: %v", err))
				failed(fmt.Sprintf("unexpected error occurred while fetching mci for install metrics agent; %s", err), err)
				return
			}

			if len(mci.Vm) == 0 {
				rec.fail(constant.StepAgentInstall, "Monitoring agent install failed", "target MCI has no VM")
				failed("mci vm's length is zero", errors.New("mci vm's length is zero"))
				return
			}

			if len(mci.Vm) == 1 {
				param.AgentHostname = mci.Vm[0].PublicIP
			} else {
				for _, v := range mci.Vm {
					if v.Id == param.NodeId {
						param.AgentHostname = v.PublicIP
					}
				}
			}

			if param.AgentHostname == "" {
				err := errors.New("agent host name afeter get mci from tumblebug must be set to not nil")
				rec.fail(constant.StepAgentInstall, "Monitoring agent install failed", "could not resolve the target node hostname")
				failed(fmt.Sprintf("invalid agent hostname for test %s; %v", param.LoadTestKey, err), err)
				return
			}
		}

		arg := MonitoringAgentInstallationParams{
			NsId:    param.NsId,
			InfraId: param.InfraId,
			NodeIds: []string{param.NodeId},
		}

		// install and run the agent for collect metrics
		_, err := l.InstallMonitoringAgent(arg)
		if err != nil {
			rec.fail(constant.StepAgentInstall, "Monitoring agent install failed", err.Error())
			failed(fmt.Sprintf("unexpected error occurred while installing monitoring agent; %s", err), err)
			return
		}

		rec.ok(constant.StepAgentInstall, "Monitoring agent installed")
		log.Info().Msgf("metrics agent installed successfully for load test; %s %s %s", arg.NsId, arg.InfraId, arg.NodeIds)
	} else {
		rec.skip(constant.StepAgentInstall, "Additional metrics not collected")
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
		StepRec:                        rec,
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

	compileDuration, executionDuration, loadTestErr := l.executeLoadTest(param, &loadGeneratorInstallInfo, rec)

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
			mci, err := l.tumblebugClient.GetMciWithContext(context.Background(), param.NsId, param.InfraId)

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
					if v.Id == param.NodeId {
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
			NsId:    param.NsId,
			InfraId: param.InfraId,
			NodeIds: []string{param.NodeId},
		}

		// install and run the agent for collect metrics
		_, err := l.InstallMonitoringAgent(arg)
		if err != nil {
			log.Error().Msgf("unexpected error occurred while fetching mci for ")
			return
		}

		log.Info().Msgf("metrics agent installed successfully for load test; %s %s %s", arg.NsId, arg.InfraId, arg.NodeIds)
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

	compileDuration, executionDuration, loadTestErr := l.executeLoadTest(param, loadGeneratorInstallInfo, nil)

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

func (l *LoadService) executeLoadTest(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo, rec *stepRecorder) (string, string, error) {
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
		rec.begin(constant.StepJmxPrepare, "Preparing and sending test plan")
		var buf bytes.Buffer
		err := parseTestPlanStructToString(&buf, param, loadGeneratorInstallInfo)
		if err != nil {
			rec.fail(constant.StepJmxPrepare, "Test plan generation failed", err.Error())
			return compileDuration, executionDuration, err
		}

		testPlan := buf.String()

		createFileCmd := fmt.Sprintf("cat << 'EOF' > %s/test_plan/%s \n%s\nEOF", loadGeneratorInstallPath, testPlanName, testPlan)

		commandReq := tumblebug.SendCommandReq{
			Command: []string{createFileCmd},
		}

		compileDuration = utils.DurationString(start)
		nsId, mciId, _, _ := getResourceNames()
		_, err = l.tumblebugClient.CommandToMciWithContext(context.Background(), nsId, mciId, commandReq)
		if err != nil {
			rec.fail(constant.StepJmxPrepare, "Test plan transfer failed", err.Error())
			return compileDuration, executionDuration, err
		}
		rec.ok(constant.StepJmxPrepare, "Test plan ready")

		rec.begin(constant.StepJmeterRun, "Running load test")
		jmeterTestCommand := generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName)

		commandReq = tumblebug.SendCommandReq{
			Command: []string{jmeterTestCommand},
		}

		stdout, err := l.tumblebugClient.CommandToMciWithContext(context.Background(), nsId, mciId, commandReq)
		if err != nil {
			rec.fail(constant.StepJmeterRun, "Load test failed", err.Error())
			return compileDuration, executionDuration, err
		}
		executionDuration = utils.DurationString(start)

		if strings.Contains(stdout, "exited with status 1") {
			rec.fail(constant.StepJmeterRun, "Load test failed", "jmeter test stopped unexpectedly")
			return compileDuration, executionDuration, errors.New("jmeter test stopped unexpectedly")
		}
		rec.ok(constant.StepJmeterRun, "Load test finished")

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
			NsId:    param.NsId,
			InfraId: param.InfraId,
			NodeId:  param.NodeId,
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
		nsId, mciId, _, _ := getResourceNames()
		_, err := l.tumblebugClient.CommandToMciWithContext(ctx, nsId, mciId, commandReq)

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
