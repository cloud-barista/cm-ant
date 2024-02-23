package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/domain"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

const (
	jmeterPath    = "third_party/jmeter/apache-jmeter-5.3/bin/jmeter"
	testPlanFile  = "sample_test_plan.jmx"
	resultCpu     = "cpu"
	resultDisk    = "disk"
	resultGeneral = "general"
	resultMemory  = "memory"
	resultNetwork = "network"
	resultSwap    = "swap"
	resultTcp     = "tcp"
)

func GetLoadTestResultHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		testId := c.Query("testId")

		if len(strings.TrimSpace(testId)) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "testId must be passed",
			})
			return
		}

		file, err := os.Open(fmt.Sprintf("temp/%s/result_%s_%s.csv", testId, resultGeneral, testId))
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)

		// Read all the records
		records, err := reader.ReadAll()
		if err != nil {
			log.Println("Error:", err)
			return
		}

		var buf bytes.Buffer

		for _, record := range records {
			for _, value := range record {
				buf.WriteString(fmt.Sprintf("%s\t", value))
			}
			buf.WriteString("\n")
		}

		c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"result": buf.String(),
		})

	}
}

func ExecuteLoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestProperties := domain.NewLoadTestProperties()
		if err := c.ShouldBindBodyWith(&loadTestProperties, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestProperties),
			})
			return
		}

		log.Printf("request body: %+v\n", loadTestProperties)

		currentTime := time.Now()
		testId := fmt.Sprintf("%d-%s", currentTime.UnixMilli(), uuid.New().String())
		loadTestProperties.TestId = testId

		testFolderPath := fmt.Sprintf("temp/%s", testId)
		err := utils.CreateFolder(testFolderPath)

		if err != nil {
			log.Printf("Error while creating folder: %s; %v\n", testFolderPath, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while creating folder: %s", testFolderPath),
			})
			return
		}

		propertiesFilePath := fmt.Sprintf("%s/config_%s.properties", testFolderPath, testId)
		propertiesData := utils.StructToMap(loadTestProperties)
		err = utils.WritePropertiesFile(propertiesFilePath, propertiesData, true)
		if err != nil {
			log.Printf("Error while writing properties filePath: %s; %v\n", propertiesFilePath, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while writing properties filePath: %s", propertiesFilePath),
			})
			return
		}

		cmdStr := jmeterCliExecutionCmdGenerator(propertiesFilePath, testPlanFile, testFolderPath, testId)
		log.Printf("this is cmd : %s\n", cmdStr)
		err = utils.SyncSysCall(cmdStr)
		if err != nil {
			log.Printf("Error while executing jmeter cmd; %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": "Error while executing jmeter cmd;",
			})

			return
		}

		resultFilePath := fmt.Sprintf("temp/%s/result_%s_%s.csv", testId, resultGeneral, testId)
		reportFolderPath := fmt.Sprintf("temp/%s/report_%s", testId, testId)
		cmdGenerateReport := jmeterReportGenerateCmdGenerator(resultFilePath, reportFolderPath)

		log.Printf("this is cmdGenerateReport : %s\n", cmdGenerateReport)
		err = utils.SyncSysCall(cmdGenerateReport)

		if err != nil {
			log.Printf("Error while executing jmeter cmd; %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": "Error while executing jmeter cmd;",
			})

			return
		}

		c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"testId": testId,
		})
	}
}

/*
-n : non gui mode
-p : use properties file
-t : location of jmeter test script
-l : location of result file
-e : html 리포트 생성
-o : html 리포트 저장 디렉토리 설정
-R : remore ips for distribute test
*/

// func jmeterExecutionCmdGenerator(propertiesPath, testScriptPath, testFolderPath, testId, reportFolderPath string) string {
// 	return fmt.Sprintf("%s -n -f -p \"%s\" -t \"%s\" -l \"%s/result_general_%s.csv\" -e -o \"%s\"", jmeterPath, propertiesPath, testScriptPath, testFolderPath, testId, reportFolderPath)
// }

func jmeterCliExecutionCmdGenerator(propertiesPath, testScriptPath, testFolderPath, testId string) string {
	return fmt.Sprintf("%s -n -f -p \"%s\" -t \"%s\" -l \"%s/result_general_%s.csv\"", jmeterPath, propertiesPath, testScriptPath, testFolderPath, testId)
}

func jmeterReportGenerateCmdGenerator(resultFilePath, reportFolderPath string) string {
	return fmt.Sprintf("%s -g %s -o %s", jmeterPath, resultFilePath, reportFolderPath)
}
