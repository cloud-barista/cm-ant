package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"strings"

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
