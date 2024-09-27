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
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
)

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

	go l.processLoadTest(param, &loadGeneratorInstallInfo, &stateArg)

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

// processLoadTest executes the load test.
// Depending on whether the installation location is local or remote, it creates the test plan and runs test commands.
// Fetches and saves test results from the local or remote system.
func (l *LoadService) processLoadTest(param RunLoadTestParam, loadGeneratorInstallInfo *LoadGeneratorInstallInfo, loadTestExecutionState *LoadTestExecutionState) {

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
		LoadTestDone:    loadTestDone,
		LoadTestKey:     param.LoadTestKey,
		InstallLocation: loadGeneratorInstallInfo.InstallLocation,
		InstallPath:     loadGeneratorInstallInfo.InstallPath,
		PublicKeyName:   loadGeneratorInstallInfo.PublicKeyName,
		PrivateKeyName:  loadGeneratorInstallInfo.PrivateKeyName,
		Username:        username,
		PublicIp:        publicIp,
		Port:            port,
		AgentInstalled:  param.AgentInstalled,
		Home:            home,
	}

	go l.fetchData(dataParam)

	defer func() {
		loadTestDone <- true
		close(loadTestDone)
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

	loadTestExecutionState.ExecutionStatus = constant.Successed
}

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
		_, err := l.tumblebugClient.CommandToMciWithContext(ctx, antNsId, antMciId, commandReq)

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
