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
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/labstack/echo/v4"
)

func InstallLoadTesterHandlerV2() echo.HandlerFunc {
	return func(c echo.Context) error {
		antTargetServerReq := api.AntTargetServerReq{}

		if err := c.Bind(&antTargetServerReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "request body is incorrect",
			})

		}

		if antTargetServerReq.NsId == "" || antTargetServerReq.McisId == "" || antTargetServerReq.VmId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "ns, mcis and vm id is essential",
			})
		}

		_, err := services.InstallLoadTesterV2(&antTargetServerReq)

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
func UninstallLoadTesterHandlerV2() echo.HandlerFunc {
	return func(c echo.Context) error {

		envId := c.Param("envId")

		if strings.TrimSpace(envId) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "load tester install environment id is essential",
			})
		}

		err := services.UninstallLoadTesterV2(envId)

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
func RunLoadTestHandlerV2() echo.HandlerFunc {
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

		loadTestKey, err := services.ExecuteLoadTestV2(&loadTestReq)

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

func StopLoadTestHandlerV2() echo.HandlerFunc {
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

func GetLoadTestResultHandlerV2() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestKey := c.QueryParam("loadTestKey")
		format := c.QueryParam("format")

		if strings.TrimSpace(loadTestKey) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"status":  "bad request",
				"message": "",
			})
		}
		result, err := services.GetLoadTestResultV2(loadTestKey, format)

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
