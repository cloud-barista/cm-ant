package handler

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/load/services"
)

func GetLoadTestResultHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		envId := c.QueryParam("envId")
		testKey := c.QueryParam("testKey")
		format := c.QueryParam("format")

		if len(strings.TrimSpace(testKey)) == 0 || len(strings.TrimSpace(envId)) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"status":  "bad request",
				"message": "testKey and envId must be passed",
			})
		}
		result, err := services.GetLoadTestResult(envId, testKey, format)

		if err != nil {
			log.Printf("sorry, internal server error while getting load test result; %s\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while getting load test result;",
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"status": "ok",
			"result": result,
		})
	}
}

func StopLoadTestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestReq := api.LoadExecutionConfigReq{}

		if err := c.Bind(&loadTestReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestReq),
			})
		}

		if loadTestReq.LoadTestKey == "" || loadTestReq.EnvId == "" {
			log.Println("error while execute [StopLoadTestHandler()]; no passing propertiesId")
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass propertiesId if you want to stop test"),
			})
		}

		err := services.StopLoadTest(loadTestReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": fmt.Sprintf("sorry, internal server error while executing load test;"),
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

func RunLoadTestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestReq := api.LoadExecutionConfigReq{}

		if err := c.Bind(&loadTestReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestReq),
			})
		}

		if loadTestReq.LoadEnvReq.Validate() != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("load test environment is not correct"),
			})
		}

		envId, testKey, loadExecutionConfigId, err := services.ExecuteLoadTest(&loadTestReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": fmt.Sprintf("sorry, internal server error while executing load test;"),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"testKey":               testKey,
			"envId":                 envId,
			"loadExecutionConfigId": loadExecutionConfigId,
			"message":               "success",
		})
	}
}

func InstallLoadGeneratorHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadEnvReq := api.LoadEnvReq{}

		if err := c.Bind(&loadEnvReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass me correct body; %v", loadEnvReq),
			})

		}

		if err := loadEnvReq.Validate(); err != nil {
			log.Printf("error while execute [InstallLoadGeneratorHandler()]; %s\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("if you install on remote, pass nsId, mcisId and username"),
			})
		}

		createdEnvId, err := services.InstallLoadGenerator(&loadEnvReq)

		if err != nil {
			log.Printf("error while executing load test; %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
			"result":  createdEnvId,
		})
	}
}

func GetLoadConfig() echo.HandlerFunc {
	return func(c echo.Context) error {
		configId := c.Param("configId")

		if configId == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "execution config id is not set",
			})

		}

		result, err := services.GetLoadExecutionConfigById(configId)

		if err != nil {
			log.Printf("error while get load test execution config; %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
			"result":  result,
		})
	}
}
