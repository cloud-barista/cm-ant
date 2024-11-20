package load

import (
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func generateCSVData(numRecords int) (string, error) {
	file, err := os.CreateTemp("/tmp", "large_data_*.csv")
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = writer.WriteString("timeStamp,elapsed,label,responseCode,responseMessage,threadName,dataType,success,failureMessage,bytes,sentBytes,grpThreads,allThreads,URL,Latency,IdleTime,Connect\n")
	if err != nil {
		return "", err
	}

	// Generate random data and write to CSV
	for i := 0; i < numRecords; i++ {
		timestamp := time.Now().Add(time.Duration(i) * time.Millisecond).UnixMilli()
		elapsed := rand.Intn(1000)                                 // Random elapsed time
		label := fmt.Sprintf("Label%d", rand.Intn(10))             // Random label
		responseCode := rand.Intn(500)                             // Random response code
		responseMessage := fmt.Sprintf("Message%d", rand.Intn(10)) // Random response message
		threadName := fmt.Sprintf("Thread%d", rand.Intn(10))       // Random thread name
		dataType := rand.Intn(10)                                  // Random data type
		success := rand.Intn(2) == 1                               // Random success (true or false)
		failureMessage := ""                                       // Can be filled with some failure message if needed
		bytes := rand.Intn(1000)                                   // Random bytes
		sentBytes := rand.Intn(1000)                               // Random sent bytes
		grpThreads := rand.Intn(100)                               // Random group threads
		allThreads := rand.Intn(200)                               // Random all threads
		url := "https://example.com"                               // Example URL
		latency := rand.Intn(500)                                  // Random latency
		idleTime := rand.Intn(200)                                 // Random idle time
		connection := rand.Intn(100)                               // Random connection

		_, err := writer.WriteString(fmt.Sprintf("%d,%d,%s,%d,%s,%s,%d,%t,%s,%d,%d,%d,%d,%s,%d,%d,%d\n",
			timestamp, elapsed, label, responseCode, responseMessage, threadName, dataType, success, failureMessage,
			bytes, sentBytes, grpThreads, allThreads, url, latency, idleTime, connection))
		if err != nil {
			return "", err
		}
	}

	err = writer.Flush()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

// {"level":"info","time":"2024-11-20T14:46:21+09:00","message":"Time taken to process CSV: 10.501558875s"}
func TestAppendResultRawData(t *testing.T) {
	numRecords := 100_000_00 // 10 million

	filePath, err := generateCSVData(numRecords)
	assert.NoError(t, err)
	defer os.Remove(filePath)
	start := time.Now()

	_, err = appendResultRawData(filePath)
	assert.NoError(t, err)

	duration := time.Since(start)
	log.Info().Msgf("Time taken to process CSV: %v", duration)
}
