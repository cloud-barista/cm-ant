package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/load/domain"
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	resultGeneral = "general"
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

func StopLoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestPropertyReq := domain.LoadTestPropertyReq{}

		if err := c.ShouldBindBodyWith(&loadTestPropertyReq, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestPropertyReq),
			})
			return
		}

		if err := loadTestPropertyReq.LoadEnvReq.Validate(); err != nil {
			log.Printf("error while execute [RunLoadTestHandler()]; %s\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("if you run on remote, pass nsId, mcisId and username"),
			})
			return
		}

		// TODO add goroutine, sse to get result asynchronously
		loadTestId, err := services.ExecuteLoadTest(loadTestPropertyReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("sorry, internal server error while executing load test; %+v", loadTestPropertyReq),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"testId":  loadTestId,
			"message": "success",
		})
	}
}

func RunLoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestPropertyReq := domain.LoadTestPropertyReq{}

		if err := c.ShouldBindBodyWith(&loadTestPropertyReq, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestPropertyReq),
			})
			return
		}

		if err := loadTestPropertyReq.LoadEnvReq.Validate(); err != nil {
			log.Printf("error while execute [RunLoadTestHandler()]; %s\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("if you run on remote, pass nsId, mcisId and username"),
			})
			return
		}

		// TODO add goroutine, sse to get result asynchronously
		loadTestId, err := services.ExecuteLoadTest(loadTestPropertyReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("sorry, internal server error while executing load test; %+v", loadTestPropertyReq),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"testId":  loadTestId,
			"message": "success",
		})
	}
}

func InstallLoadGeneratorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadInstallReq := domain.LoadEnvReq{}

		if err := c.ShouldBindBodyWith(&loadInstallReq, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("pass me correct body; %v", loadInstallReq),
			})
			return
		}

		if err := loadInstallReq.Validate(); err != nil {
			log.Printf("error while execute [InstallLoadGeneratorHandler()]; %s\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("if you install on remote, pass nsId, mcisId and username"),
			})
			return
		}

		err := services.InstallLoadGenerator(loadInstallReq)

		if err != nil {
			log.Printf("error while executing load test; %+v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "something went wrong.try again.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	}
}
