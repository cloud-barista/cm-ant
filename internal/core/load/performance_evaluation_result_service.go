package load

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
)

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

func (l *LoadService) GetLoadTestResult(param GetLoadTestResultParam) (interface{}, error) {
	loadTestKey := param.LoadTestKey
	fileName := fmt.Sprintf("%s_result.csv", loadTestKey)
	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)
	toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)
	resultMap, err := appendResultRawData(toFilePath)
	if err != nil {
		return nil, err
	}

	var resultSummaries []ResultSummary

	for label, results := range resultMap {
		resultSummaries = append(resultSummaries, ResultSummary{
			Label:   label,
			Results: results,
		})
	}

	formattedDate, err := resultFormat(param.Format, resultSummaries)

	if err != nil {
		return nil, err
	}

	return formattedDate, nil
}

func (l *LoadService) GetLoadTestMetrics(param GetLoadTestResultParam) ([]MetricsSummary, error) {
	loadTestKey := param.LoadTestKey
	metrics := []string{"cpu", "disk", "memory", "network"}
	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)

	metricsMap := make(map[string][]*MetricsRawData)
	var err error
	for _, v := range metrics {

		fileName := fmt.Sprintf("%s_%s_result.csv", loadTestKey, v)
		toPath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

		metricsMap, err = appendMetricsRawData(metricsMap, toPath)
		if err != nil {
			return nil, err
		}

	}

	var metricsSummaries []MetricsSummary

	for label, metrics := range metricsMap {
		metricsSummaries = append(metricsSummaries, MetricsSummary{
			Label:   label,
			Metrics: metrics,
		})
	}

	if err != nil {
		return nil, err
	}

	return metricsSummaries, nil
}

func (l *LoadService) GetLastLoadTestResult(param GetLastLoadTestResultParam) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stateQueryParam := GetLoadTestExecutionStateParam{
		NsId:  param.NsId,
		MciId: param.MciId,
		VmId:  param.VmId,
	}
	state, err := l.loadRepo.GetLoadTestExecutionStateTx(ctx, stateQueryParam)

	if err != nil {
		log.Error().Msgf("Error fetching load test execution state infos; %v", err)
		return nil, err
	}

	loadTestKey := state.LoadTestKey
	fileName := fmt.Sprintf("%s_result.csv", loadTestKey)
	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)
	toFilePath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)
	resultMap, err := appendResultRawData(toFilePath)
	if err != nil {
		return nil, err
	}

	var resultSummaries []ResultSummary

	for label, results := range resultMap {
		resultSummaries = append(resultSummaries, ResultSummary{
			Label:   label,
			Results: results,
		})
	}

	formattedDate, err := resultFormat(param.Format, resultSummaries)

	if err != nil {
		return nil, err
	}

	return formattedDate, nil
}

func (l *LoadService) GetLastLoadTestMetrics(param GetLastLoadTestResultParam) ([]MetricsSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stateQueryParam := GetLoadTestExecutionStateParam{
		NsId:  param.NsId,
		MciId: param.MciId,
		VmId:  param.VmId,
	}

	state, err := l.loadRepo.GetLoadTestExecutionStateTx(ctx, stateQueryParam)

	if err != nil {
		log.Error().Msgf("Error fetching load test execution state infos; %v", err)
		return nil, err
	}

	loadTestKey := state.LoadTestKey

	if !state.WithMetrics {
		log.Error().Msgf("%s does not contain metrics collection", loadTestKey)
		return nil, errors.New("metrics does not collected while performance evaluation")
	}

	metrics := []string{"cpu", "disk", "memory", "network"}
	resultFolderPath := utils.JoinRootPathWith("/result/" + loadTestKey)

	metricsMap := make(map[string][]*MetricsRawData)

	for _, v := range metrics {

		fileName := fmt.Sprintf("%s_%s_result.csv", loadTestKey, v)
		toPath := fmt.Sprintf("%s/%s", resultFolderPath, fileName)

		metricsMap, err = appendMetricsRawData(metricsMap, toPath)
		if err != nil {
			return nil, err
		}

	}

	var metricsSummaries []MetricsSummary

	for label, metrics := range metricsMap {
		metricsSummaries = append(metricsSummaries, MetricsSummary{
			Label:   label,
			Metrics: metrics,
		})
	}

	if err != nil {
		return nil, err
	}

	return metricsSummaries, nil
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

func appendResultRawData(filePath string) (map[string][]*ResultRawData, error) {
	var resultMap = make(map[string][]*ResultRawData)

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

		tr := &ResultRawData{
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
		if _, ok := resultMap[label]; !ok {
			resultMap[label] = []*ResultRawData{tr}
		} else {
			resultMap[label] = append(resultMap[label], tr)
		}
	}

	return resultMap, nil
}

func appendMetricsRawData(mrds map[string][]*MetricsRawData, filePath string) (map[string][]*MetricsRawData, error) {
	csvRows, err := utils.ReadCSV(filePath)
	if err != nil || csvRows == nil {
		return nil, err
	}

	if len(*csvRows) <= 1 {
		return nil, errors.New("metrics data file is empty")
	}

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

		rd := &MetricsRawData{
			Value:     value,
			Unit:      u,
			IsError:   isError,
			Timestamp: t,
		}

		if _, ok := mrds[label]; !ok {
			mrds[label] = []*MetricsRawData{rd}
		} else {
			mrds[label] = append(mrds[label], rd)
		}
	}

	return mrds, nil
}

func aggregate(resultRawDatas []ResultSummary) []*LoadTestStatistics {
	var statistics []*LoadTestStatistics

	for i := range resultRawDatas {

		record := resultRawDatas[i]
		var requestCount, totalElapsed, totalBytes, totalSentBytes, errorCount int
		var elapsedList []int
		if len(record.Results) < 1 {
			continue
		}

		startTime := record.Results[0].Timestamp
		endTime := record.Results[0].Timestamp
		for _, record := range record.Results {
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

		labelStat := LoadTestStatistics{
			Label:         record.Label,
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

		statistics = append(statistics, &labelStat)
	}

	return statistics
}

func resultFormat(format constant.ResultFormat, resultSummaries []ResultSummary) (any, error) {
	if resultSummaries == nil {
		return nil, nil
	}

	switch format {
	case constant.Aggregate:
		return aggregate(resultSummaries), nil
	}

	return resultSummaries, nil
}
