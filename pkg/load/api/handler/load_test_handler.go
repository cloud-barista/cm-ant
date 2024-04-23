package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cm-ant/pkg/load/services"
)

// GetLoadTestResult by format
//
// @Summary			Get the the result of single load test result
// @Description		After start load test, get the result of load test.
// @Tags			[Load Test]
// @Accept			json
// @Produce			json
// @Param			loadTestKey query 		string true 	"load test key"
// @Param			format 	query 		string false 	"format of load test result aggregate"
// @Success			200	{object}		interface{}
// @Failure			400	{object}		string			"loadTestKey must be passed"
// @Failure			500	{object}		string			"sorry, internal server error while getting load test result;"
// @Router			/ant/load/result 	[get]
func GetLoadTestResultHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKey := c.QueryParam("loadTestKey")
		format := c.QueryParam("format")

		if strings.TrimSpace(loadTestKey) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"status":  "bad request",
				"message": "",
			})
		}
		result, err := services.GetLoadTestResult(loadTestKey, format)

		if err != nil {
			log.Printf("sorry, internal server error while getting load test result; %s\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while getting load test result;",
			})
		}
		var marBuf bytes.Buffer

		enc := json.NewEncoder(&marBuf)

		if err := enc.Encode(result); err != nil {
			return err
		}

		resultBytes := marBuf.Bytes()

		header := c.Response().Header()

		header.Set("Content-Type", "application/json")
		header.Set("Content-Encoding", "gzip")

		var gzBuf bytes.Buffer

		gz := gzip.NewWriter(&gzBuf)

		if _, err := gz.Write(resultBytes); err != nil {
			log.Printf("sorry, internal server error while getting load test result; %s\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while getting load test result;",
			})
		}
		if err := gz.Close(); err != nil {
			log.Printf("sorry, internal server error while getting load test result; %s\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while getting load test result;",
			})
		}

		c.Response().WriteHeader(http.StatusOK)
		c.Response().Write(gzBuf.Bytes())

		return nil
	}
}

// StopLoadTest
//
// @Summary			Stop load test
// @Description		After start load test, stop the load test by passing the load test key.
// @Tags			[Load Test]
// @Accept			json
// @Produce			json
// @Param			loadTestKeyReq	body 	api.LoadTestKeyReq	true 	"load test key"
// @Success			200	{object}			string					"success"
// @Failure			400	{object}			string					"pass propertiesId if you want to stop test"
// @Failure			500	{object}			string					"sorry, internal server error while executing load test;"
// @Router			/ant/load/stop 			[post]
func StopLoadTestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKeyReq := api.LoadTestKeyReq{}

		if err := c.Bind(&loadTestKeyReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestKeyReq),
			})
		}

		if loadTestKeyReq.LoadTestKey == "" {
			log.Println("error while execute [StopLoadTestHandler()]; no passing propertiesId")
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "pass propertiesId if you want to stop test",
			})
		}

		err := services.StopLoadTest(loadTestKeyReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while executing load test;",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

// StartLoadTest
//
// @Summary			Start load test
// @Description		Start load test. Load Environment Id must be passed or Load Environment must be defined.
// @Tags			[Load Test]
// @Accept			json
// @Produce			json
// @Param			loadTestReq 	body 	api.LoadExecutionConfigReq 			true 	"load test execution configuration request"
// @Success			200	{object}			map[string]string					`{ "testKey": testKey, "envId": envId, "loadExecutionConfigId": loadExecutionConfigId, "message": "success" }`
// @Failure			400	{object}			string								"load test environment is not correct"
// @Failure			500	{object}			string								"sorry, internal server error while executing load test;"
// @Router			/ant/load/start 		[post]
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
				"message": "load test environment is not correct",
			})
		}

		envId, testKey, loadExecutionConfigId, err := services.ExecuteLoadTest(&loadTestReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while executing load test;",
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

// InstallLoadGenerator
//
// @Summary			Install load test tool
// @Description		Install load generation tools in the delivered load test environment
// @Tags			[Load Test]
// @Accept			json
// @Produce			json
// @Param			loadEnvReq 		body 	api.LoadEnvReq 			true 		"load test environment request"
// @Success			200	{object}			map[string]string					`{ "message": "success", "result":  createdEnvId }`
// @Failure			400	{object}			string								"load test environment is not correct"
// @Failure			500	{object}			string								"sorry, internal server error while executing load test;"
// @Router			/ant/load/install 		[post]
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
				"message": "if you install on remote, pass nsId, mcisId and username",
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

func GetAllLoadConfigHandler() echo.HandlerFunc {
	return func(c echo.Context) error {

		result, err := services.GetAllLoadExecutionConfig()

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

func GetLoadConfigHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKey := c.Param("loadTestKey")

		if loadTestKey == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "execution config id is not set",
			})

		}

		result, err := services.GetLoadExecutionConfig(loadTestKey)

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

// GetAllLoadExecutionState
//
// @Summary			Get all load execution state
// @Description		Get all the load test execution state.
// @Tags			[Load Test]
// @Accept			json
// @Produce			json
// @Success			200	{object}			api.LoadExecutionStateRes
// @Failure			500	{object}			string								"something went wrong.try again."
// @Router			/ant/load/state 		[get]
func GetAllLoadExecutionStateHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		result, err := services.GetAllLoadExecutionState()

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
