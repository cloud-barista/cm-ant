package managers

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/outbound"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/utils"
)

type LoadTestManager struct {
}

func NewLoadTestManager() LoadTestManager {
	return LoadTestManager{}
}

func calculatePercentile(elapsedList []int, percentile float64) float64 {
	index := int(math.Ceil(float64(len(elapsedList))*percentile)) - 1

	return float64(elapsedList[index])
}

func calculateMedian(data []int) float64 {
	n := len(data)
	if n%2 == 0 {
		return float64(data[n/2-1]+data[n/2]) / 2
	}
	return float64(data[n/2])
}

func findMin(elapsedList []int) float64 {
	if len(elapsedList) == 0 {
		return 0
	}

	return float64(elapsedList[0])
}

func findMax(elapsedList []int) float64 {
	if len(elapsedList) == 0 {
		return 0
	}

	return float64(elapsedList[len(elapsedList)-1])
}

func calculateErrorPercent(errorCount, requestCount int) float64 {
	if requestCount == 0 {
		return 0
	}
	errorPercent := float64(errorCount) / float64(requestCount) * 100
	return errorPercent
}

func calculateThroughput(totalRequests int, totalMillTime int) float64 {
	return float64(totalRequests) / (float64(totalMillTime)) * 1000
}

func calculateReceivedKBPerSec(totalBytes int, totalMillTime int) float64 {
	return (float64(totalBytes) / 1024) / (float64(totalMillTime)) * 1000
}

func calculateSentKBPerSec(totalBytes int, totalMillTime int) float64 {
	return (float64(totalBytes) / 1024) / (float64(totalMillTime)) * 1000
}

type tempRecord struct {
	Elapsed    int // time to last byte
	Bytes      int
	SentBytes  int
	URL        string
	Latency    int // time to first byte
	IdleTime   int // time not spent sampling in jmeter (milliseconds) (generally 0)
	Connection int // time to establish connection
	IsError    bool
	Timestamp  time.Time
}

func (l *LoadTestManager) GetResult(testId string) (interface{}, error) {

	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	resultFilePath := fmt.Sprintf("%s/result/%s_result.csv", jmeterPath, testId)

	// TODO: testId 가 어떤 구성으로(LoadEnvReq) 로 테스트를 실행했는지 확인 필요 - 현재 로컬만 커버
	// TODO: db 설정 후 간단 통계 계산 데이터를 저장하는 방식이 나쁘지 않아 보임

	csvRows := utils.ReadCSV(resultFilePath)

	labelGroup := make(map[string][]*tempRecord)

	// every time is basically millisecond
	for i, row := range csvRows[1:] {
		label := row[2]

		elapsed, err := strconv.Atoi(row[1])
		if err != nil {
			log.Printf("[%d] elapsed has error %s\n", i, err)
			continue
		}
		bytes, err := strconv.Atoi(row[9])
		if err != nil {
			log.Printf("[%d] bytes has error %s\n", i, err)
			continue
		}
		sentBytes, err := strconv.Atoi(row[10])
		if err != nil {
			log.Printf("[%d] sentBytes has error %s\n", i, err)
			continue
		}
		latency, err := strconv.Atoi(row[14])
		if err != nil {
			log.Printf("[%d] latency has error %s\n", i, err)
			continue
		}
		idleTime, err := strconv.Atoi(row[15])
		if err != nil {
			log.Printf("[%d] idleTime has error %s\n", i, err)
			continue
		}
		connection, err := strconv.Atoi(row[16])
		if err != nil {
			log.Printf("[%d] connection has error %s\n", i, err)
			continue
		}
		unixMilliseconds, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			log.Printf("[%d] connection has error %s\n", i, err)
			continue
		}

		isError := row[7] == "false"
		url := row[13]
		t := time.UnixMilli(unixMilliseconds)

		tr := &tempRecord{
			Elapsed:    elapsed,
			Bytes:      bytes,
			SentBytes:  sentBytes,
			IsError:    isError,
			URL:        url,
			Latency:    latency,
			IdleTime:   idleTime,
			Connection: connection,
			Timestamp:  t,
		}
		if _, ok := labelGroup[label]; !ok {
			labelGroup[label] = []*tempRecord{tr}
		} else {
			labelGroup[label] = append(labelGroup[label], tr)
		}
	}

	var statistics []model.LoadTestStatistics

	for label, records := range labelGroup {
		var requestCount, totalElapsed, totalBytes, totalSentBytes, errorCount int
		var elapsedList []int
		if len(records) < 1 {
			continue
		}

		startTime := records[0].Timestamp
		endTime := records[0].Timestamp
		for _, record := range records {
			requestCount++
			if !record.IsError {
				totalElapsed += record.Elapsed
			}

			totalBytes += record.Bytes
			totalSentBytes += record.SentBytes

			if record.IsError {
				errorCount++
			}

			if record.Timestamp.Before(startTime) {
				startTime = record.Timestamp
			}
			if record.Timestamp.After(endTime) {
				endTime = record.Timestamp
			}

			elapsedList = append(elapsedList, record.Elapsed)
		}

		// total Elapsed time and running time is different
		runningTime := endTime.Sub(startTime).Milliseconds()

		// for percentile calculation purpose
		sort.Ints(elapsedList)

		average := float64(totalElapsed) / float64(requestCount)
		median := calculateMedian(elapsedList)
		ninetyPercent := calculatePercentile(elapsedList, 0.9)
		ninetyFive := calculatePercentile(elapsedList, 0.95)
		ninetyNine := calculatePercentile(elapsedList, 0.99)
		calcMin := findMin(elapsedList)
		calcMax := findMax(elapsedList)
		errorPercent := calculateErrorPercent(errorCount, requestCount)
		throughput := calculateThroughput(requestCount, int(runningTime))
		receivedKB := calculateReceivedKBPerSec(totalBytes, int(runningTime))
		sentKB := calculateSentKBPerSec(totalSentBytes, int(runningTime))

		labelStat := model.LoadTestStatistics{
			Label:         label,
			RequestCount:  requestCount,
			Average:       average,
			Median:        median,
			NinetyPercent: ninetyPercent,
			NinetyFive:    ninetyFive,
			NinetyNine:    ninetyNine,
			MinTime:       calcMin,
			MaxTime:       calcMax,
			ErrorPercent:  errorPercent,
			Throughput:    throughput,
			ReceivedKB:    receivedKB,
			SentKB:        sentKB,
		}

		statistics = append(statistics, labelStat)
	}

	return statistics, nil
}

func (l *LoadTestManager) Install(installReq api.LoadEnvReq) error {
	installScriptPath := configuration.JoinRootPathWith("/script/install-jmeter.sh")

	if installReq.Type == constant.Remote {
		tumblebugUrl := outbound.TumblebugHostWithPort()
		multiLineCommand, err := readAndParseScript(installScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		commandReq := outbound.SendCommandReq{
			Command:  []string{multiLineCommand},
			UserName: installReq.Username,
		}

		stdout, err := outbound.SendCommandTo(tumblebugUrl, installReq.NsId, installReq.McisId, commandReq)

		if err != nil {
			log.Println(stdout)
			return err
		}

		log.Println(stdout)

	} else if installReq.Type == constant.Local {

		err := utils.Script(installScriptPath, []string{
			fmt.Sprintf("JMETER_WORK_DIR=%s", configuration.Get().Load.JMeter.WorkDir),
			fmt.Sprintf("JMETER_VERSION=%s", configuration.Get().Load.JMeter.Version),
		})
		if err != nil {
			return fmt.Errorf("error while installing jmeter; %s", err)
		}

	}

	return nil
}

func (l *LoadTestManager) Stop(property api.LoadTestPropertyReq) error {

	// TODO code cloud test using tumblebug
	if property.LoadEnvReq.Type == constant.Remote {
		tumblebugUrl := outbound.TumblebugHostWithPort()

		killCmd := killCmdGen(property)

		commandReq := outbound.SendCommandReq{
			Command:  []string{killCmd},
			UserName: property.LoadEnvReq.Username,
		}

		stdout, err := outbound.SendCommandTo(tumblebugUrl, property.LoadEnvReq.NsId, property.LoadEnvReq.McisId, commandReq)

		if err != nil {
			log.Println(stdout)
			return err
		}

	} else if property.LoadEnvReq.Type == constant.Local {

		log.Printf("[%s] stop load test on local", property.PropertiesId)
		killCmd := killCmdGen(property)

		err := utils.InlineCmd(killCmd)

		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (l *LoadTestManager) Run(property api.LoadTestPropertyReq) (string, error) {
	var testId string
	testFolderSetupScript := configuration.JoinRootPathWith("/script/pre-execute-jmeter.sh")
	testPlanName := "test_plan_1.jmx"
	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	jmeterVersion := configuration.Get().Load.JMeter.Version

	// TODO code cloud test using tumblebug
	if property.LoadEnvReq.Type == constant.Remote {
		tumblebugUrl := outbound.TumblebugHostWithPort()

		// 1. Installation check
		err := l.Install(property.LoadEnvReq)

		if err != nil {
			return "", err
		}

		// 2. pre-requirement check

		multiLineCommand, err := readAndParseScript(testFolderSetupScript)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return "", err
		}

		multiLineCommand = strings.Replace(multiLineCommand, "${TEST_PLAN_NAME:=\"test_plan_1.jmx\"}", testPlanName, 1)

		commandReq := outbound.SendCommandReq{
			Command:  []string{multiLineCommand},
			UserName: property.LoadEnvReq.Username,
		}

		stdout, err := outbound.SendCommandTo(tumblebugUrl, property.LoadEnvReq.NsId, property.LoadEnvReq.McisId, commandReq)

		if err != nil {
			log.Printf("error occured; %s\n", err)
			log.Println(stdout)
			return "", err
		}

		log.Println(stdout)

		// 3. execute jmeter test
		jmeterTestCommand := executionCmdGen(property, testPlanName, fmt.Sprintf("%s_result.csv", property.PropertiesId))

		commandReq = outbound.SendCommandReq{
			Command:  []string{jmeterTestCommand},
			UserName: property.LoadEnvReq.Username,
		}

		stdout, err = outbound.SendCommandTo(tumblebugUrl, property.LoadEnvReq.NsId, property.LoadEnvReq.McisId, commandReq)

		if err != nil {
			log.Printf("error occured; %s\n", err)
			log.Println(stdout)
			return "", err
		}

		log.Println(stdout)

	} else if property.LoadEnvReq.Type == constant.Local {

		log.Printf("[%s] Do load test on local", property.PropertiesId)

		exist := utils.ExistCheck(jmeterPath)

		if !exist {
			loadInstallReq := api.LoadEnvReq{
				Type: constant.Local,
			}

			err := l.Install(loadInstallReq)

			if err != nil {
				log.Printf("error while execute [Run()]; %s\n", err)
				return "", err
			}
		}

		err := utils.Script(testFolderSetupScript, []string{
			fmt.Sprintf("TEST_PLAN_NAME=%s", testPlanName),
			fmt.Sprintf("JMETER_WORK_DIR=%s", jmeterPath),
			fmt.Sprintf("JMETER_VERSION=%s", jmeterVersion),
		})

		if err != nil {
			log.Println(err)
			return "", err
		}

		jmeterTestCommand := executionCmdGen(property, testPlanName, fmt.Sprintf("%s_result.csv", property.PropertiesId))
		err = utils.InlineCmd(jmeterTestCommand)

		if err != nil {
			log.Println(err)
			return "", err
		}

		// 3. save test configuration
		testId = property.PropertiesId
	}

	return testId, nil
}

func executionCmdGen(p api.LoadTestPropertyReq, testPlanName, resultFileName string) string {
	jmeterConf := configuration.Get().Load.JMeter

	var builder strings.Builder
	testPath := fmt.Sprintf("%s/test_plan/%s", jmeterConf.WorkDir, testPlanName)
	resultPath := fmt.Sprintf("%s/result/%s", jmeterConf.WorkDir, resultFileName)

	builder.WriteString(fmt.Sprintf("%s/apache-jmeter-%s/bin/jmeter.sh", jmeterConf.WorkDir, jmeterConf.Version))
	builder.WriteString(" -n -f")
	builder.WriteString(fmt.Sprintf(" -Jthreads=%s", p.Threads))
	builder.WriteString(fmt.Sprintf(" -JrampTime=%s", p.RampTime))
	builder.WriteString(fmt.Sprintf(" -JloopCount=%s", p.LoopCount))
	builder.WriteString(fmt.Sprintf(" -Jprotocol=%s", p.HttpReqs.Protocol))
	builder.WriteString(fmt.Sprintf(" -Jhostname=%s", p.HttpReqs.Hostname))
	builder.WriteString(fmt.Sprintf(" -Jport=%s", p.HttpReqs.Port))
	builder.WriteString(fmt.Sprintf(" -Jpath=%s", p.HttpReqs.Path))
	builder.WriteString(fmt.Sprintf(" -JbodyData=%s", p.HttpReqs.BodyData))
	builder.WriteString(fmt.Sprintf(" -JbodyData=%s", p.LoopCount))
	builder.WriteString(fmt.Sprintf(" -JpropertiesId=%s", p.PropertiesId))
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	return builder.String()
}

func killCmdGen(p api.LoadTestPropertyReq) string {
	grepRegex := fmt.Sprintf("'\\/bin\\/ApacheJMeter\\.jar.*-JpropertiesId=%s'", p.PropertiesId)

	return fmt.Sprintf("kill -9 $(ps -ef | grep -E %s | awk '{print $2}')", grepRegex)
}

func readAndParseScript(scriptPath string) (string, error) {
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Println("file doesn't exist on correct path")
		return "", err
	}

	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	jmeterVersion := configuration.Get().Load.JMeter.Version

	multiLineCommand := strings.Replace(string(data), "#!/bin/bash", "", 1)
	multiLineCommand = strings.Replace(multiLineCommand, "${JMETER_WORK_DIR:=\"${HOME}/jmeter\"}", jmeterPath, 1)
	multiLineCommand = strings.Replace(multiLineCommand, "${JMETER_VERSION:=\"5.3\"}", jmeterVersion, 1)

	return multiLineCommand, nil
}
