package jmetermanager

import (
	"context"
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/outbound"
	"github.com/melbahja/goph"
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

type JMeterLoadTestManager struct {
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

type processedData struct {
	No         int
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

func makeLocalProcessedData(resultFilePath string) (*map[string][]*processedData, error) {
	csvRows, err := utils.ReadCSV(resultFilePath)
	if err != nil || csvRows == nil {
		return nil, err
	}

	labelGroup := make(map[string][]*processedData)

	// every time is basically millisecond
	for i, row := range (*csvRows)[1:] {
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

		tr := &processedData{
			No:         i,
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
			labelGroup[label] = []*processedData{tr}
		} else {
			labelGroup[label] = append(labelGroup[label], tr)
		}
	}

	return &labelGroup, nil
}

func aggregate(processedData *map[string][]*processedData) *[]model.LoadTestStatistics {
	var statistics []model.LoadTestStatistics

	for label, records := range *processedData {
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

	return &statistics
}

func generateFormat(format string, processedData *map[string][]*processedData) (interface{}, error) {
	if processedData == nil {
		return nil, nil
	}

	switch format {
	case "aggregate":
		return aggregate(processedData), nil
	}

	return processedData, nil
}

func (j *JMeterLoadTestManager) GetResult(loadEnv *model.LoadEnv, testKey, format string) (interface{}, error) {

	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	fileName := fmt.Sprintf("%s_result.csv", testKey)
	resultFilePath := fmt.Sprintf("%s/result/%s", jmeterPath, fileName)

	var processedData *map[string][]*processedData

	if (*loadEnv).InstallLocation == constant.Remote {
		switch (*loadEnv).RemoteConnectionType {
		case constant.BuiltIn:

		case constant.PrivateKey, constant.Password:

			tempFolderPath := configuration.JoinRootPathWith("/temp")

			err := utils.CreateFolderIfNotExist(tempFolderPath)
			if err != nil {
				return nil, err
			}

			copiedFilePath, err := downloadIfNotExist(loadEnv, resultFilePath, tempFolderPath, fileName)

			if err != nil {
				return nil, err
			}
			processedData, err = makeLocalProcessedData(copiedFilePath)
			if err != nil {
				return nil, err
			}

		}

	} else if (*loadEnv).InstallLocation == constant.Local {
		var err error
		processedData, err = makeLocalProcessedData(resultFilePath)
		if err != nil {
			return nil, err
		}

	}

	formattedDate, err := generateFormat(format, processedData)

	if err != nil {
		return nil, err
	}

	return formattedDate, nil
}

func downloadIfNotExist(loadEnv *model.LoadEnv, resultFilePath, tempFolderPath, fileName string) (string, error) {
	copiedFilePath := fmt.Sprintf("%s/%s", tempFolderPath, fileName)

	if exist := utils.ExistCheck(copiedFilePath); !exist {
		var auth goph.Auth
		var err error

		if loadEnv.RemoteConnectionType == constant.PrivateKey {
			auth, err = goph.Key(loadEnv.Cert, "")
			if err != nil {
				return "", err
			}
		} else if loadEnv.RemoteConnectionType == constant.Password {
			auth = goph.Password(loadEnv.Cert)
			if err != nil {
				return "", err
			}
		}

		client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)

		defer client.Close()
		err = client.Download(resultFilePath, copiedFilePath)

		if err != nil {
			return "", err
		}
		return copiedFilePath, nil
	}

	return copiedFilePath, nil
}

func (j *JMeterLoadTestManager) Install(loadEnvReq *api.LoadEnvReq) error {
	installScriptPath := configuration.JoinRootPathWith("/script/install-jmeter.sh")

	if loadEnvReq.InstallLocation == constant.Remote {
		installationCommand, err := readAndParseScript(installScriptPath)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}

		switch loadEnvReq.RemoteConnectionType {
		case constant.BuiltIn:
			tumblebugUrl := outbound.TumblebugHostWithPort()

			commandReq := outbound.SendCommandReq{
				Command:  []string{installationCommand},
				UserName: loadEnvReq.Username,
			}

			stdout, err := outbound.SendCommandTo(tumblebugUrl, loadEnvReq.NsId, loadEnvReq.McisId, commandReq)

			if err != nil {
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnvReq.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnvReq.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnvReq.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnvReq.Cert)
				if err != nil {
					return err
				}
			}

			client, err := goph.New(loadEnvReq.Username, loadEnvReq.PublicIp, auth)
			if err != nil {
				return err
			}

			defer client.Close()

			out, err := client.RunContext(context.Background(), installationCommand)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))
		}

	} else if loadEnvReq.InstallLocation == constant.Local {

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

func (j *JMeterLoadTestManager) Stop(loadTestReq api.LoadExecutionConfigReq) error {

	killCmd := killCmdGen(loadTestReq)

	// TODO code cloud test using tumblebug
	loadEnv := loadTestReq.LoadEnvReq
	if loadEnv.InstallLocation == constant.Remote {

		switch loadEnv.RemoteConnectionType {
		case constant.BuiltIn:
			tumblebugUrl := outbound.TumblebugHostWithPort()

			commandReq := outbound.SendCommandReq{
				Command:  []string{killCmd},
				UserName: loadEnv.Username,
			}

			stdout, err := outbound.SendCommandTo(tumblebugUrl, loadEnv.NsId, loadEnv.McisId, commandReq)

			if err != nil {
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnv.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnv.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnv.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnv.Cert)
				if err != nil {
					return err
				}
			}

			// 1. ssh client connection
			client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)
			if err != nil {
				return err
			}

			defer client.Close()

			out, err := client.RunContext(context.Background(), killCmd)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))
		}

	} else if loadEnv.InstallLocation == constant.Local {
		log.Printf("[%s] stop load test on local", loadTestReq.LoadTestKey)

		err := utils.InlineCmd(killCmd)

		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (j *JMeterLoadTestManager) Run(loadTestReq *api.LoadExecutionConfigReq) error {
	testFolderSetupScript := configuration.JoinRootPathWith("/script/pre-execute-jmeter.sh")
	testPlanName := "test_plan_1.jmx"
	jmeterPath := configuration.Get().Load.JMeter.WorkDir
	jmeterVersion := configuration.Get().Load.JMeter.Version
	loadEnv := loadTestReq.LoadEnvReq

	// TODO code cloud test using tumblebug
	if loadEnv.InstallLocation == constant.Remote {
		preRequirementCmd, err := readAndParseScript(testFolderSetupScript)
		if err != nil {
			log.Println("file doesn't exist on correct path")
			return err
		}
		preRequirementCmd = strings.Replace(preRequirementCmd, "${TEST_PLAN_NAME:=\"test_plan_1.jmx\"}", testPlanName, 1)

		switch loadEnv.RemoteConnectionType {
		case constant.BuiltIn:
			tumblebugUrl := outbound.TumblebugHostWithPort()
			commandReq := outbound.SendCommandReq{
				Command:  []string{preRequirementCmd},
				UserName: loadTestReq.LoadEnvReq.Username,
			}

			// 1. check pre-requisition
			stdout, err := outbound.SendCommandTo(tumblebugUrl, loadTestReq.LoadEnvReq.NsId, loadTestReq.LoadEnvReq.McisId, commandReq)

			if err != nil {
				log.Printf("error occured; %s\n", err)
				log.Println(stdout)
				return err
			}

			log.Println(stdout)

			// 2. execute jmeter test
			jmeterTestCommand := executionCmdGen(loadTestReq, testPlanName, fmt.Sprintf("%s_result.csv", loadTestReq.LoadTestKey))

			commandReq = outbound.SendCommandReq{
				Command:  []string{jmeterTestCommand},
				UserName: loadTestReq.LoadEnvReq.Username,
			}

			stdout, err = outbound.SendCommandTo(tumblebugUrl, loadTestReq.LoadEnvReq.NsId, loadTestReq.LoadEnvReq.McisId, commandReq)

			if err != nil {
				log.Printf("error occured; %s\n", err)
				log.Println(stdout)
				return err
			}

			log.Println(stdout)
		case constant.PrivateKey, constant.Password:
			var auth goph.Auth
			var err error

			if loadEnv.RemoteConnectionType == constant.PrivateKey {
				auth, err = goph.Key(loadEnv.Cert, "")
				if err != nil {
					return err
				}
			} else if loadEnv.RemoteConnectionType == constant.Password {
				auth = goph.Password(loadEnv.Cert)
				if err != nil {
					return err
				}
			}

			// 1. ssh client connection
			client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)
			if err != nil {
				return err
			}

			defer client.Close()

			// 2. check pre-requisition
			out, err := client.RunContext(context.Background(), preRequirementCmd)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))

			// 3. execute jmeter test
			jmeterTestCommand := executionCmdGen(loadTestReq, testPlanName, fmt.Sprintf("%s_result.csv", loadTestReq.LoadTestKey))
			out, err = client.RunContext(context.Background(), jmeterTestCommand)

			if err != nil {
				log.Println(string(out))
				return err
			}

			log.Println(string(out))
		}

	} else if loadEnv.InstallLocation == constant.Local {

		log.Printf("[%s] Do load test on local", loadTestReq.LoadTestKey)

		exist := utils.ExistCheck(jmeterPath)

		if !exist {
			loadInstallReq := api.LoadEnvReq{
				InstallLocation: constant.Local,
			}

			err := j.Install(&loadInstallReq)

			if err != nil {
				log.Printf("error while execute [Run()]; %s\n", err)
				return err
			}
		}

		err := utils.Script(testFolderSetupScript, []string{
			fmt.Sprintf("TEST_PLAN_NAME=%s", testPlanName),
			fmt.Sprintf("JMETER_WORK_DIR=%s", jmeterPath),
			fmt.Sprintf("JMETER_VERSION=%s", jmeterVersion),
		})

		if err != nil {
			log.Println(err)
			return err
		}

		jmeterTestCommand := executionCmdGen(loadTestReq, testPlanName, fmt.Sprintf("%s_result.csv", loadTestReq.LoadTestKey))
		err = utils.InlineCmd(jmeterTestCommand)

		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func executionCmdGen(p *api.LoadExecutionConfigReq, testPlanName, resultFileName string) string {
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
	builder.WriteString(fmt.Sprintf(" -JpropertiesId=%s", p.LoadTestKey))
	builder.WriteString(fmt.Sprintf(" -t=%s", testPath))
	builder.WriteString(fmt.Sprintf(" -l=%s", resultPath))

	return builder.String()
}

func killCmdGen(p api.LoadExecutionConfigReq) string {
	grepRegex := fmt.Sprintf("'\\/bin\\/ApacheJMeter\\.jar.*-JpropertiesId=%s'", p.LoadTestKey)

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
