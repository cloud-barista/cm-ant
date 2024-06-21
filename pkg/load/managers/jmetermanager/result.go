package jmetermanager

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/melbahja/goph"
)

type resultRawData struct {
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

type metricsRawData struct {
	Value     string
	Unit      string
	IsError   bool
	Timestamp time.Time
}
type metricsUnits struct {
	Multiple float64
	Unit     string
}

var tags = map[string]metricsUnits{
	"cpu_all_combined": {
		Multiple: 0.001,
		Unit:     "%",
	},
	"cpu_all_idle": {
		Multiple: 0.001,
		Unit:     "%",
	},
	"memory_all_used": {
		Multiple: 0.001,
		Unit:     "%",
	},
	"memory_all_free": {
		Multiple: 0.001,
		Unit:     "%",
	},
	"memory_all_used_kb": {
		Multiple: 0.000001,
		Unit:     "mb",
	},
	"memory_all_free_kb": {
		Multiple: 0.000001,
		Unit:     "mb",
	},
	"disk_read_kb": {
		Multiple: 0.001,
		Unit:     "kb",
	},
	"disk_write_kb": {
		Multiple: 0.001,
		Unit:     "kb",
	},
	"disk_use": {
		Multiple: 0.001,
		Unit:     "%",
	},
	"disk_total": {
		Multiple: 0.000001,
		Unit:     "mb",
	},
	"network_recv_kb": {
		Multiple: 0.000001,
		Unit:     "kb",
	},
	"network_sent_kb": {
		Multiple: 0.001,
		Unit:     "kb",
	},
}

func (j *JMeterLoadTestManager) GetResult(loadEnv *model.LoadEnv, loadTestKey, format string) (interface{}, error) {

	jmeterPath := config.AppConfig.Load.JMeter.Dir
	fileName := fmt.Sprintf("%s_result.csv", loadTestKey)
	resultFilePath := fmt.Sprintf("%s/result/%s", jmeterPath, fileName)
	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)
	toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)
	var resultRawData = make(map[string][]*resultRawData)

	err := utils.CreateFolderIfNotExist(utils.JoinRootPathWith("/result"))
	if err != nil {
		return nil, err
	}

	err = utils.CreateFolderIfNotExist(resultFolderPath)
	if err != nil {
		return nil, err
	}

	if (*loadEnv).InstallLocation == constant.Remote {

		if !utils.ExistCheck(toFilePath) {

			auth, err := goph.Key(loadEnv.PemKeyPath, "")
			if err != nil {
				return nil, err
			}

			client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)

			if err != nil {
				return nil, err
			}

			defer client.Close()

			err = client.Download(resultFilePath, toFilePath)

			if err != nil {
				return nil, err
			}
		}

	} else if (*loadEnv).InstallLocation == constant.Local {

		if !utils.ExistCheck(toFilePath) {
			err := utils.InlineCmd(fmt.Sprintf("cp %s %s", resultFilePath, toFilePath))

			if err != nil {
				return nil, err
			}
		}

	}

	resultRawData, err = appendResultRawData(resultRawData, toFilePath)
	if err != nil {
		return nil, err
	}

	formattedDate, err := resultFormat(format, resultRawData)

	if err != nil {
		return nil, err
	}

	return formattedDate, nil
}

func (j *JMeterLoadTestManager) GetMetrics(loadEnv *model.LoadEnv, loadTestKey, format string) (interface{}, error) {

	jmeterPath := config.AppConfig.Load.JMeter.Dir
	metricsPrePath := fmt.Sprintf("%s/result", jmeterPath)
	metrics := []string{"cpu", "disk", "memory", "network"}
	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)
	var metricsRawData = make(map[string][]*metricsRawData)

	err := utils.CreateFolderIfNotExist(utils.JoinRootPathWith("/result"))
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	errCh := make(chan error)

	if (*loadEnv).InstallLocation == constant.Remote {

		err := utils.CreateFolderIfNotExist(resultFolderPath)
		if err != nil {
			return nil, err
		}
		auth, err := goph.Key(loadEnv.PemKeyPath, "")
		if err != nil {
			return nil, err
		}
		client, err := goph.New(loadEnv.Username, loadEnv.PublicIp, auth)
		if err != nil {
			return nil, err
		}

		defer client.Close()

		for _, v := range metrics {
			wg.Add(1)
			go func(prefix string) {
				defer wg.Done()
				fileName := fmt.Sprintf("%s_%s_result.csv", loadTestKey, prefix)
				toPath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)
				fromPath := fmt.Sprintf("%s/%s", metricsPrePath, fileName)

				if !utils.ExistCheck(toPath) {
					err = client.Download(fromPath, toPath)
					if err != nil {
						log.Println(err)
						errCh <- err
					}
				}
			}(v)
		}

	} else if (*loadEnv).InstallLocation == constant.Local {

		for _, v := range metrics {
			wg.Add(1)

			go func(prefix string) {
				defer wg.Done()
				fileName := fmt.Sprintf("%s_%s_result.csv", loadTestKey, prefix)
				toPath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)
				fromPath := fmt.Sprintf("%s/%s", metricsPrePath, fileName)
				var err error

				if !utils.ExistCheck(toPath) {
					err = utils.InlineCmd(fmt.Sprintf("cp %s %s", fromPath, toPath))
					if err != nil {
						errCh <- err
					}
				}
			}(v)

		}
	}

	wg.Wait()
	close(errCh)

	if len(errCh) != 0 {
		err := <-errCh
		log.Println(err)
		return nil, err
	}

	for _, v := range metrics {

		fileName := fmt.Sprintf("%s_%s_result.csv", loadTestKey, v)
		toPath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

		metricsRawData, err = appendMetricsRawData(metricsRawData, toPath)
		if err != nil {
			return nil, err
		}

	}

	formattedDate, err := metricFormat(format, metricsRawData)

	if err != nil {
		return nil, err
	}

	return formattedDate, nil
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

func appendResultRawData(resultRawDataMap map[string][]*resultRawData, filePath string) (map[string][]*resultRawData, error) {
	csvRows, err := utils.ReadCSV(filePath)
	if err != nil || csvRows == nil {
		return nil, err
	}

	if len(*csvRows) <= 1 {
		return nil, errors.New("result data file is empty")
	}

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
			log.Printf("[%d] time has error %s\n", i, err)
			continue
		}

		isError := row[7] == "false"
		url := row[13]
		t := time.UnixMilli(unixMilliseconds)

		tr := &resultRawData{
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
		if _, ok := resultRawDataMap[label]; !ok {
			resultRawDataMap[label] = []*resultRawData{tr}
		} else {
			resultRawDataMap[label] = append(resultRawDataMap[label], tr)
		}
	}

	return resultRawDataMap, nil
}

func appendMetricsRawData(resultRawDataMap map[string][]*metricsRawData, filePath string) (map[string][]*metricsRawData, error) {
	csvRows, err := utils.ReadCSV(filePath)
	if err != nil || csvRows == nil {
		return nil, err
	}

	if len(*csvRows) <= 1 {
		return nil, errors.New("metrics data file is empty")
	}

	// every time is basically millisecond
	for i, row := range (*csvRows)[1:] {
		isError := row[7] == "false"
		intValue, err := strconv.Atoi(row[1])
		if err != nil {
			log.Printf("[%d] value has error %s\n", i, err)
			continue
		}

		var label string
		var value string
		var u string

		if isError {
			label = row[2]
		} else {
			words := strings.Split(row[2], " ")
			label = words[len(words)-1]

			unit, ok := tags[label]
			if !ok {
				continue
			}

			floatValue := float64(intValue) * unit.Multiple
			value = strconv.FormatFloat(floatValue, 'f', 3, 64)
			u = unit.Unit
		}

		unixMilliseconds, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			log.Printf("[%d] time has error %s\n", i, err)
			continue
		}

		t := time.UnixMilli(unixMilliseconds)

		rd := &metricsRawData{
			Value:     value,
			Unit:      u,
			IsError:   isError,
			Timestamp: t,
		}

		if _, ok := resultRawDataMap[label]; !ok {
			resultRawDataMap[label] = []*metricsRawData{rd}
		} else {
			resultRawDataMap[label] = append(resultRawDataMap[label], rd)
		}
	}

	return resultRawDataMap, nil
}

func aggregate(resultRawDatas map[string][]*resultRawData) *[]model.LoadTestStatistics {
	var statistics []model.LoadTestStatistics

	for label, records := range resultRawDatas {
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

func resultFormat(format string, resultRawDatas map[string][]*resultRawData) (interface{}, error) {
	if resultRawDatas == nil {
		return nil, nil
	}

	switch format {
	case "aggregate":
		return aggregate(resultRawDatas), nil
	}

	return resultRawDatas, nil
}

func metricFormat(format string, metricsRawDatas map[string][]*metricsRawData) (interface{}, error) {
	if metricsRawDatas == nil {
		return nil, nil
	}

	switch format {
	case "aggregate":
		return nil, nil
	}

	return metricsRawDatas, nil
}
