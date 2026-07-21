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
	constant.StepPrecheck:         1,
	constant.StepGeneratorInstall: 2,
	constant.StepAgentInstall:     3,
	constant.StepJmxPrepare:       4,
	constant.StepJmeterRun:        5,
	constant.StepResultFetch:      6,
}

// subStepSeq orders the sub-steps within their phase. They are seeded alongside the phases so
// the console can draw the whole tree from the first poll rather than watching rows appear.
var subStepSeq = map[constant.ExecutionStep]int{
	constant.SubTargetExists:    1,
	constant.SubTargetRunning:   2,
	constant.SubTargetReachable: 3,
	constant.SubMetricPortOpen:  4,
	constant.SubRemoteCommand:   5,

	constant.SubGeneratorReachable: 1,

	constant.SubAgentInstall: 1,
	constant.SubAgentProcess: 2,
	constant.SubAgentPort:    3,
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
	if seq, ok := stepSeq[step.Name]; ok {
		step.Seq = seq
	} else {
		step.Seq = subStepSeq[step.Name]
	}
	if err := s.l.loadRepo.UpsertLoadTestExecutionStepTx(context.Background(), step); err != nil {
		log.Warn().Msgf("failed to record execution step %s; %v", step.Name, err)
	}
}

// seed creates the full pipeline as pending so the web can render every step upfront.
func (s *stepRecorder) seed() {
	for name := range stepSeq {
		s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepPending})
	}
	for name := range subStepSeq {
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

// skip marks a step as deliberately not run. It stamps a finish time like any other ending:
// without one the step keeps reporting a growing elapsed figure, which reads as "still busy"
// when it is in fact long done.
func (s *stepRecorder) skip(name constant.ExecutionStep, message string) {
	now := time.Now()
	s.upsert(&LoadTestExecutionStep{Name: name, Status: constant.StepSkipped, FinishAt: &now, Message: message})
}

// generatorRunMu serializes load test runs. The load generator is a shared singleton
// (one ant-default MCI reused across runs), so concurrent runs would contend on the same
// VM (JMeter CPU contention, install/reset races). Runs are therefore serialized and a
// concurrent request is rejected rather than corrupting the shared generator (BAR-1414).
var generatorRunMu sync.Mutex

// RunLoadTest initiates the load test and performs necessary initializations.
// Generates a load test key, installs the load generator or retrieves existing installation information,
// saves the load test execution state, and then asynchronously runs the load test.
// resolveNodeUid asks cb-tumblebug which VM the given node name currently refers to.
//
// Returns "" when the lookup fails. A load test is worth running even if we cannot label it,
// and an empty uid is understood downstream as "unknown" rather than as a mismatch - the
// same treatment rows written before this column existed get.
func (l *LoadService) resolveNodeUid(ctx context.Context, nsId, infraId, nodeId string) string {
	if nsId == "" || infraId == "" || nodeId == "" {
		return ""
	}

	vm, err := l.tumblebugClient.GetVmWithContext(ctx, nsId, infraId, nodeId)
	if err != nil {
		log.Warn().Msgf("could not resolve node uid for %s/%s/%s: %v", nsId, infraId, nodeId, err)
		return ""
	}
	return vm.Uid
}

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

	// Record which VM this run belongs to, not just what it was called.
	//
	// Node ids are names that get reused: cb-tumblebug builds them as "{group}-{index}", so
	// deleting a VM and recreating it under the same name gives an identical
	// ns/infra/node triple. Without the uid a later lookup cannot tell whether the run it
	// found belongs to the VM in front of the user or to its predecessor.
	//
	// Resolved here rather than taken from the request: the client would have to send it
	// correctly for the guarantee to hold, and cb-tumblebug already answers the question.
	nodeUid := l.resolveNodeUid(ctx, param.NsId, param.InfraId, param.NodeId)

	stateArg := LoadTestExecutionState{
		LoadTestKey:                 loadTestKey,
		ExecutionStatus:             constant.OnProcessing,
		StartAt:                     startAt,
		ExpectedFinishAt:            e,
		TotalExpectedExcutionSecond: totalExecutionSecond,
		NsId:                        param.NsId,
		InfraId:                     param.InfraId,
		NodeId:                      param.NodeId,
		NodeUid:                     nodeUid,
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
		l.applyGeneratorIdlePolicy(param) // BAR-1413: keep / suspend / terminate the generator when idle
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

	// Persist the requested configuration up front, before anything can fail. A run that dies in
	// pre-check or generator install used to leave no info row at all, because the save happened
	// only after both succeeded — so GetLoadTestExecutionInfo (which Re-run reads to pre-fill the
	// form) had nothing to return and answered 500. The parameters are known from the request, so
	// record them now; the generator id is not known yet and is linked in after install.
	var hs []LoadTestExecutionHttpInfo
	for _, h := range param.HttpReqs {
		hs = append(hs, LoadTestExecutionHttpInfo{
			Method:   h.Method,
			Protocol: h.Protocol,
			Hostname: h.Hostname,
			Port:     h.Port,
			Path:     h.Path,
			BodyData: h.BodyData,
		})
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

		LoadTestExecutionHttpInfos: hs,
		LoadTestExecutionStateId:   loadTestExecutionState.ID,
	}
	if err := l.loadRepo.SaveForLoadTestExecutionTx(context.Background(), &loadTestExecutionInfoParam); err != nil {
		failed(fmt.Sprintf("Error saving load test execution info: %v", err), err)
		return
	}
	loadTestExecutionState.TestExecutionInfoId = loadTestExecutionInfoParam.ID

	// Check the environment before building anything (BAR-1553). A missing target or a closed
	// port is answered in seconds here, instead of surfacing minutes into a run.
	precheckCtx, cancelPrecheck := context.WithTimeout(context.Background(), 2*time.Minute)
	err := l.runPrecheck(precheckCtx, param, rec)
	cancelPrecheck()
	if err != nil {
		failed(fmt.Sprintf("Precheck failed: %v", err), err)
		return
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

	// A reused generator can be recorded as installed and still be unreachable, because the
	// key that reaches it lives in this container rather than anywhere durable: replacing the
	// container leaves every existing generator with an authorized key whose private half is
	// gone. Nothing before the very end of the run notices - the load runs on the generator
	// itself and only fetching the results needs ssh - so the run would spend its whole length
	// to fail at collection.
	//
	// cb-tumblebug still holds the VM's own key, so the way back in is through it: regenerate
	// the pair and have tumblebug put the new public half in place. That is also why the key
	// is not worth persisting here - keeping a copy would mean owning its lifetime and its
	// exposure, to save a step tumblebug can do on demand.
	if err := l.ensureGeneratorReachable(loadGeneratorInstallInfo, rec); err != nil {
		rec.fail(constant.StepGeneratorInstall, "Load generator unreachable", err.Error())
		failed(fmt.Sprintf("load generator is not reachable and could not be recovered: %v", err), err)
		return
	}

	// The generator exists now; link the info row saved earlier to it. TestExecutionInfoId was
	// already set from the early save above.
	if err = l.loadRepo.UpdateLoadTestExecutionInfoGeneratorTx(context.Background(), param.LoadTestKey, loadGeneratorInstallInfo.ID); err != nil {
		failed(fmt.Sprintf("Error linking load test execution info to the generator: %v", err), err)
		return
	}
	loadTestExecutionInfoParam.LoadGeneratorInstallInfoId = loadGeneratorInstallInfo.ID
	loadTestExecutionState.GeneratorInstallInfoId = loadGeneratorInstallInfo.ID

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

		rec.ok(constant.SubAgentInstall, "Monitoring agent files installed")

		// Installed is not the same as running, and running is not the same as answering.
		// The install call reports success as soon as the remote command returns, and the
		// script behind it starts the agent with nohup and prints success either way — so a
		// target with no agent process at all was reported as installed, and the run then
		// spent 27 minutes waiting on a port that was never going to answer (BAR-1552).
		if err := l.verifyMetricAgent(param, rec); err != nil {
			// Losing metrics is not worth failing the run for. Say what happened, drop the
			// metrics and carry on with the load figures.
			param.CollectAdditionalSystemMetrics = false
			loadTestExecutionState.WithMetrics = false
			rec.skip(constant.StepAgentInstall, "Continuing without metrics - the agent is not answering")
			log.Warn().Msgf("metric agent unavailable, continuing without metrics; %v", err)
		} else {
			rec.ok(constant.StepAgentInstall, "Monitoring agent installed")
			log.Info().Msgf("metrics agent installed successfully for load test; %s %s %s", arg.NsId, arg.InfraId, arg.NodeIds)
		}
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

	dataParam.Finished = make(chan struct{})
	go l.fetchData(dataParam)

	defer func() {
		loadTestDone <- true
		close(loadTestDone)
		if dataParam.Finished != nil {
			// Wait for the final result rsync to complete before finalizing state and applying
			// the generator idle policy, so suspend/terminate can't interrupt it (BAR-1413).
			<-dataParam.Finished
		}
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

		// The remote path above records each step; this one recorded none, so a local run
		// showed "jmx_prepare" and "jmeter_run" as pending from start to finish and gave no
		// clue where it was. Both paths now report the same way.
		rec.begin(constant.StepJmxPrepare, "Preparing test plan")

		exist := utils.ExistCheck(loadGeneratorInstallPath)

		if !exist {
			rec.fail(constant.StepJmxPrepare, "Load generator missing", fmt.Sprintf("nothing installed at %s", loadGeneratorInstallPath))
			return compileDuration, executionDuration, errors.New("load generator installaion is not validated")
		}

		outputFile, err := os.Create(fmt.Sprintf("%s/test_plan/%s.jmx", loadGeneratorInstallPath, loadTestKey))
		if err != nil {
			rec.fail(constant.StepJmxPrepare, "Test plan write failed", err.Error())
			return compileDuration, executionDuration, err
		}

		err = parseTestPlanStructToString(outputFile, param, loadGeneratorInstallInfo)

		if err != nil {
			rec.fail(constant.StepJmxPrepare, "Test plan generation failed", err.Error())
			return compileDuration, executionDuration, err
		}
		rec.ok(constant.StepJmxPrepare, "Test plan ready")

		rec.begin(constant.StepJmeterRun, "Running load test")
		jmeterTestCommand := generateJmeterExecutionCmd(loadGeneratorInstallPath, loadGeneratorInstallVersion, testPlanName, resultFileName)
		compileDuration = utils.DurationString(start)

		err = utils.InlineCmd(jmeterTestCommand)
		executionDuration = utils.DurationString(start)
		if err != nil {
			rec.fail(constant.StepJmeterRun, "Load test failed", fmt.Sprintf("jmeter stopped unexpectedly: %v", err))
			return compileDuration, executionDuration, fmt.Errorf("jmeter test stopped unexpectedly; %w", err)
		}
		rec.ok(constant.StepJmeterRun, "Load test finished")
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

// belongsToDifferentNode reports whether a recorded run demonstrably belongs to a VM other
// than the one being asked about.
//
// Only a disagreement between two known uids counts. An unknown uid on either side - a run
// recorded before the column existed, or cb-tumblebug being unreachable - is not evidence of
// anything, and treating it as a mismatch would block stopping runs that are perfectly
// legitimate.
func belongsToDifferentNode(recordedUid, currentUid string) bool {
	if recordedUid == "" || currentUid == "" {
		return false
	}
	return recordedUid != currentUid
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

	// Guard the key-less lookup, which resolves to "the most recent run for this
	// ns/infra/node" - names that get reused, so the run may belong to a VM that has since
	// been replaced. Stopping that one would kill a test nobody asked about.
	//
	// Not reachable over HTTP today: the handler rejects a stop request that omits the load
	// test key, so callers always arrive with one and take the branch above. The guard is
	// here because the service accepts the key-less shape and nothing but this handler check
	// stands between it and the wrong test.
	//
	// A stored uid that disagrees with the live one is that case. An empty uid on either side
	// means "unknown" (rows predating the column, or cb-tumblebug unreachable) and is left
	// alone rather than guessed at.
	if param.LoadTestKey == "" {
		currentUid := l.resolveNodeUid(ctx, param.NsId, param.InfraId, param.NodeId)
		if belongsToDifferentNode(state.NodeUid, currentUid) {
			return fmt.Errorf(
				"the most recent load test on %s/%s/%s belongs to a different VM (recorded uid %s, current uid %s); pass its load test key to stop it",
				param.NsId, param.InfraId, param.NodeId, state.NodeUid, currentUid,
			)
		}
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

// applyGeneratorIdlePolicy suspends or tears down the shared load generator after a run,
// per load.generator.idle ("keep" | "suspend" | "terminate"). Default "keep" preserves the
// previous behavior (the generator stays running for fast reuse). It runs only after the
// final result rsync has completed (guaranteed by the caller) and only for remote generators.
// BAR-1413 / FR-MA2-PERF-007-01.
func (l *LoadService) applyGeneratorIdlePolicy(param RunLoadTestParam) {
	idle := strings.ToLower(strings.TrimSpace(config.AppConfig.Load.Generator.Idle))
	if idle == "" || idle == "keep" {
		return
	}

	timeout, err := time.ParseDuration(config.AppConfig.Load.Timeout.CommandExecution)
	if err != nil {
		timeout = 50 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Resolve the generator's actual location. When a run reuses a generator by ID, the
	// request's InstallLoadGenerator is blanked out, so consult the install record instead.
	location := param.InstallLoadGenerator.InstallLocation
	if param.LoadGeneratorInstallInfoId != 0 {
		if info, e := l.loadRepo.GetValidLoadGeneratorInstallInfoByIdTx(ctx, param.LoadGeneratorInstallInfoId); e == nil {
			location = info.InstallLocation
		}
	}
	if location != constant.Remote {
		return // local generators have no VM lifecycle to manage
	}

	nsId, mciId, _, _ := getResourceNames()
	switch idle {
	case "suspend":
		log.Info().Msgf("idle policy: suspending load generator %s/%s", nsId, mciId)
		if err := l.tumblebugClient.ControlLifecycleWithContext(ctx, nsId, mciId, "suspend"); err != nil {
			log.Warn().Msgf("idle policy: failed to suspend generator %s/%s: %v", nsId, mciId, err)
		}
	case "terminate":
		log.Info().Msgf("idle policy: terminating load generator %s/%s", nsId, mciId)
		l.forceResetLoadGenerator(ctx, nsId, mciId)
	default:
		log.Warn().Msgf("idle policy: unknown value %q (expected keep|suspend|terminate); leaving generator as-is", idle)
	}
}
