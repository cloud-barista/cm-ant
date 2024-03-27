package handler

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"log"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GetLoadTestResultHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		testId := c.Param("testId")

		if len(strings.TrimSpace(testId)) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "testId must be passed",
			})
			return
		}
		result, err := services.GetLoadTestResult(testId)

		if err != nil {
			log.Printf("sorry, internal server error while getting load test result; %s\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "sorry, internal server error while getting load test result;",
			})
			return
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"result": result,
		})

	}
}

func StopLoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestPropertyReq := api.LoadTestPropertyReq{}

		if err := c.ShouldBindBodyWith(&loadTestPropertyReq, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestPropertyReq),
			})
			return
		}

		if loadTestPropertyReq.PropertiesId == "" {
			log.Println("error while execute [StopLoadTestHandler()]; no passing propertiesId")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("pass propertiesId if you want to stop test"),
			})
			return
		}

		err := services.StopLoadTest(loadTestPropertyReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("sorry, internal server error while executing load test; %+v", loadTestPropertyReq),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	}
}

func RunLoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestPropertyReq := api.LoadTestPropertyReq{}

		if err := c.ShouldBindBodyWith(&loadTestPropertyReq, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestPropertyReq),
			})
			return
		}

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
		loadInstallReq := api.LoadEnvReq{}

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
