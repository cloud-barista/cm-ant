package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
)

type LoadTestProperties struct {
	Threads  int64  `form:"threads" binding:"omitempty"`
	RampTime int64  `form:"rampTime" binding:"omitempty"`
	Loop     int64  `form:"loop" binding:"omitempty"`
	Hostname string `form:"hostname" bind:"omitempty"`
	Port     int64  `form:"port" bind:"omitempty"`
	Path     string `form:"path" bind:"omitempty"`
}

func NewLoadTestProperties() LoadTestProperties {
	return LoadTestProperties{
		Threads:  1,
		RampTime: 1,
		Loop:     1,
		Hostname: "localhost",
		Port:     1324,
		Path:     "/milkyway/cpus",
	}
}

func LoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestProperties := NewLoadTestProperties()

		if err := c.ShouldBindQuery(&loadTestProperties); err != nil {
			// handle query bind error
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "request param is incorrect",
			})
			return
		}

		currentTime := time.Now()
		formattedTimestamp := currentTime.Format("20060102150405")

		folderPath := fmt.Sprintf("temp/%s", formattedTimestamp)
		err := utils.CreateFolder(folderPath)

		if err != nil {
			fmt.Printf("Error while creating folder: %s; %v\n", folderPath, err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while creating folder: %s", folderPath),
			})
			return
		}

		reportFolderPath := fmt.Sprintf("%s/test_%s_report", folderPath, formattedTimestamp)
		err = utils.CreateFolder(reportFolderPath)
		if err != nil {
			fmt.Printf("Error while creating folder: %s; %v\n", reportFolderPath, err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while creating folder: %s", reportFolderPath),
			})
			return
		}

		propertiesFilePath := fmt.Sprintf("temp/%s/%s_config.properties", formattedTimestamp, formattedTimestamp)
		propertiesData := utils.StructToMap(loadTestProperties)
		err = utils.WritePropertiesFile(propertiesFilePath, propertiesData)
		if err != nil {
			fmt.Printf("Error while writing properties filePath: %s; %v\n", propertiesFilePath, err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while writing properties filePath: %s", propertiesFilePath),
			})
			return
		}

		cmdStr := jmeterExecutionCmdGenerator(propertiesFilePath, "sample_test_plan.jmx", folderPath, formattedTimestamp, reportFolderPath)

		result, err := utils.SysCall(cmdStr)
		if err != nil {
			fmt.Printf("Error while executing jmeter cmd: %s; %v\n", result, err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while executing jmeter cmd: %s", result),
			})

			return
		}

		c.JSON(http.StatusOK, map[string]string{
			"status":    "ok",
			"resultKey": formattedTimestamp,
		})
	}
}

func jmeterExecutionCmdGenerator(propertiesPath, testFilePath, folderPath, timeStamp, reportFolderPath string) string {
	return fmt.Sprintf("third_party/jmeter/apache-jmeter-5.6.3/bin/jmeter.sh -p %s -n -t %s -l %s/test_%s_result.csv -e -o %s", propertiesPath, testFilePath, folderPath, timeStamp, reportFolderPath)
}
