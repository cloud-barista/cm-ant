package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/constant"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cm-ant/pkg/load/services"
)

// GetLoadTestMetricsHandler
// @Id				LoadTestMetrics
// @Summary			Get the result of single load test metrics
// @Description		Get the result of metrics for target server.
// @Tags			[Load Test Result]
// @Accept			json
// @Produce			json
// @Param			loadTestKey query 		string true 	"load test key"
// @Success			200	{object}		interface{}
// @Failure			400	{object}		string			"loadTestKey must be passed"
// @Failure			500	{object}		string			"sorry, internal server error while getting load test result;"
// @Router			/ant/api/v1/load/result/metrics 	[get]
func GetLoadTestMetricsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKey := c.QueryParam("loadTestKey")
		format := c.QueryParam("format")

		if strings.TrimSpace(loadTestKey) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"status":  "bad request",
				"message": "",
			})
		}
		result, err := services.GetLoadTestMetrics(loadTestKey, format)

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

// GetAllLoadConfigHandler
// @Id				LoadExecutionConfigs
// @Summary			Get all load execution config
// @Description		Get all the load test execution configurations.
// @Tags			[Load Test Configuration]
// @Accept			json
// @Produce			json
// @Success			200	{object}			[]api.LoadExecutionRes
// @Failure			500	{object}			string								"something went wrong.try again."
// @Router			/ant/api/v1/load/config 		[get]
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

// GetLoadConfigHandler
// @Id				LoadExecutionConfig
// @Summary			Get load execution config
// @Description		Get a load test execution config by load test key.
// @Tags			[Load Test Configuration]
// @Accept			json
// @Produce			json
// @Param			loadTestKey 			path 					string 			true	"load test eky"
// @Success			200	{object}			api.LoadExecutionRes
// @Failure			500	{object}			string									"something went wrong. try again."
// @Router			/ant/api/v1/load/config/{loadTestKey}		[get]
func GetLoadConfigHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKey := c.Param("loadTestKey")

		if loadTestKey == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "load test key is not set",
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

// GetAllLoadExecutionStateHandler
// @Id				LoadExecutionStates
// @Summary			Get all load execution state
// @Description		Get all the load test execution state.
// @Tags			[Load Test State]
// @Accept			json
// @Produce			json
// @Success			200	{object}			[]api.LoadExecutionStateRes
// @Failure			500	{object}			string								"something went wrong.try again."
// @Router			/ant/api/v1/load/state 		[get]
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

// GetLoadExecutionStateHandler
// @Id				LoadExecutionState
// @Summary			Get load execution state
// @Description		Get a load test execution state by load test key.
// @Tags			[Load Test State]
// @Accept			json
// @Produce			json
// @Param			loadTestKey 			path 					string 			true	"load test key"
// @Success			200	{object}			api.LoadExecutionStateRes
// @Failure			500	{object}			string								"something went wrong. try again."
// @Router			/ant/api/v1/load/state/{loadTestKey} 		[get]
func GetLoadExecutionStateHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKey := c.Param("loadTestKey")
		if loadTestKey == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "load test key is not set",
			})

		}

		result, err := services.GetLoadExecutionState(loadTestKey)

		if err != nil {
			log.Printf("error while get load test execution state; %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong. try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
			"result":  result,
		})
	}
}

// InstallLoadTesterHandler
// @Id				InstallLoadTester
// @Summary			Install load test tester
// @Description		Install load test tester in the delivered load test environment
// @Tags			[Load Test Tester]
// @Accept			json
// @Produce			json
// @Param			loadEnvReq 		body 	api.LoadEnvReq 			true 		"load test environment request"
// @Success			200	{object}			map[string]string					`{ "message": "success", "result":  createdEnvId }`
// @Failure			400	{object}			string								"load test environment is not correct"
// @Failure			500	{object}			string								"sorry, internal server error while executing load test;"
// @Router			/ant/api/v1/load/tester 		[post]
func InstallLoadTesterHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		antLoadEnvReq := api.LoadEnvReq{}

		if err := c.Bind(&antLoadEnvReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "request body is incorrect",
			})

		}

		if antLoadEnvReq.InstallLocation == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "install location must set local/remote",
			})
		}

		if antLoadEnvReq.InstallLocation == constant.Remote &&
			(antLoadEnvReq.NsId == "" || antLoadEnvReq.McisId == "" || antLoadEnvReq.VmId == "") {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "ns, mcis and vm id is essential if you want to install on remote",
			})
		}

		_, err := services.InstallLoadTester(&antLoadEnvReq)

		if err != nil {
			log.Printf("error while executing load test; %+v", err)
			if errors.Is(err, context.DeadlineExceeded) {
				return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
					"message": "execution time is too long",
				})
			}

			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong. try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

// UninstallLoadTesterHandler
// @Id				UninstallLoadTester
// @Summary			Uninstall load test tester
// @Description		Uninstall load test tester in the delivered load test environment
// @Tags			[Load Test Tester]
// @Accept			json
// @Produce			json
// @Param			envId 			path 	string 			true 		"load test environment id"
// @Success			200	{object}			map[string]string					`{ "message": "success" }`
// @Failure			400	{object}			string								"pass me correct body;"
// @Failure			500	{object}			string								"something went wrong.try again."
// @Router			/ant/api/v1/load/tester/{envId} 		[delete]
func UninstallLoadTesterHandler() echo.HandlerFunc {
	return func(c echo.Context) error {

		envId := c.Param("envId")

		if strings.TrimSpace(envId) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "load tester install environment id is essential",
			})
		}

		err := services.UninstallLoadTester(envId)

		if err != nil {
			log.Printf("error while uninstall load test tool; %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

// RunLoadTestHandler
// @Id				StartLoadTest
// @Summary			Start load test
// @Description		Start load test. Load Environment Id must be passed or Load Environment must be defined.
// @Tags			[Load Test Execution]
// @Accept			json
// @Produce			json
// @Param			loadTestReq 	body 	api.LoadExecutionConfigReq 			true 	"load test execution configuration request"
// @Success			200	{object}			map[string]string					`{ "testKey": testKey, "envId": envId, "loadExecutionConfigId": loadExecutionConfigId, "message": "success" }`
// @Failure			400	{object}			string								"load test environment is not correct"
// @Failure			500	{object}			string								"sorry, internal server error while executing load test;"
// @Router			/ant/api/v1/load/start 		[post]
func RunLoadTestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestReq := api.LoadExecutionConfigReq{}

		if err := c.Bind(&loadTestReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestReq),
			})
		}

		loadTestKey, err := services.ExecuteLoadTest(&loadTestReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "sorry, internal server error while executing load test;",
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"loadTestKey": loadTestKey,
			"message":     "success",
		})

	}
}

// StopLoadTestHandler
// @Id				StopLoadTest
// @Summary			Stop load test
// @Description		After start load test, stop the load test by passing the load test key.
// @Tags			[Load Test Execution]
// @Accept			json
// @Produce			json
// @Param			loadTestKeyReq	body 	api.LoadTestKeyReq	true 	"load test key"
// @Success			200	{object}			string					"success"
// @Failure			400	{object}			string					"pass propertiesId if you want to stop test"
// @Failure			500	{object}			string					"sorry, internal server error while executing load test;"
// @Router			/ant/api/v1/load/stop 			[post]
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

// GetLoadTestResultHandler
// @Id				LoadTestResult
// @Summary			Get the result of single load test result
// @Description		After start load test, get the result of load test.
// @Tags			[Load Test Result]
// @Accept			json
// @Produce			json
// @Param			loadTestKey query 		string true 	"load test key"
// @Param			format 	query 		string false 	"format of load test result aggregate"
// @Success			200	{object}		interface{}
// @Failure			400	{object}		string			"loadTestKey must be passed"
// @Failure			500	{object}		string			"sorry, internal server error while getting load test result;"
// @Router			/ant/api/v1/load/result 	[get]
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
