package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/antcontainer"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type LoadTestProperties struct {
	Threads   string `form:"threads" binding:"required,gte=1"`
	RampTime  string `form:"rampTime"`
	LoopCount string `form:"loopCount" binding:"required,gte=1"`
	Protocol  string `form:"protocol" bind:"required"`
	Hostname  string `form:"hostname" bind:"required"`
	Port      string `form:"port"`
	Path      string `form:"path"`
}

func NewLoadTestProperties() LoadTestProperties {
	return LoadTestProperties{
		Threads:   "1",
		RampTime:  "1",
		LoopCount: "1",
		Protocol:  "http",
	}
}

func LoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestProperties := NewLoadTestProperties()
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
		formattedTimestamp := fmt.Sprintf("%d", currentTime.UnixMilli())

		folderPath := fmt.Sprintf("temp/%s", formattedTimestamp)
		err := utils.CreateFolder(folderPath)

		if err != nil {
			log.Printf("Error while creating folder: %s; %v\n", folderPath, err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while creating folder: %s", folderPath),
			})
			return
		}

		reportFolderPath := fmt.Sprintf("%s/test_%s_report", folderPath, formattedTimestamp)
		err = utils.CreateFolder(reportFolderPath)
		if err != nil {
			log.Printf("Error while creating folder: %s; %v\n", reportFolderPath, err)
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
			log.Printf("Error while writing properties filePath: %s; %v\n", propertiesFilePath, err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while writing properties filePath: %s", propertiesFilePath),
			})
			return
		}

		cmdStr := jmeterExecutionCmdGenerator(propertiesFilePath, "sample_test_plan.jmx", folderPath, formattedTimestamp, reportFolderPath)

		result, err := utils.SysCall(cmdStr)
		if err != nil {
			log.Printf("Error while executing jmeter cmd: %s; %v\n", result, err)
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

func SlaveCreateHandler() gin.HandlerFunc {
	
	return func(c *gin.Context) {
		containerManager := antcontainer.CManager
		err := containerManager.BuildJMeterDockerImage()
		if err != nil {
			log.Printf("Error while creating container; %v\n", err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while creating container; %v", err),
			})

			return
		}
	}
}

func jmeterExecutionCmdGenerator(propertiesPath, testFilePath, folderPath, timeStamp, reportFolderPath string) string {
	return fmt.Sprintf("third_party/jmeter/apache-jmeter-5.6.3/bin/jmeter.sh -p %s -n -t %s -l %s/test_%s_result.csv -e -o %s", propertiesPath, testFilePath, folderPath, timeStamp, reportFolderPath)
}
