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
		testId := c.Param("testId")

		if len(strings.TrimSpace(testId)) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"status":  "bad request",
				"message": "testId must be passed",
			})
		}
		result, err := services.GetLoadTestResult(testId)

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
		loadTestPropertyReq := api.LoadTestPropertyReq{}

		if err := c.Bind(&loadTestPropertyReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestPropertyReq),
			})
		}

		if loadTestPropertyReq.PropertiesId == "" {
			log.Println("error while execute [StopLoadTestHandler()]; no passing propertiesId")
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass propertiesId if you want to stop test"),
			})
		}

		err := services.StopLoadTest(loadTestPropertyReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": fmt.Sprintf("sorry, internal server error while executing load test; %+v", loadTestPropertyReq),
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

func RunLoadTestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadTestPropertyReq := api.LoadTestPropertyReq{}

		if err := c.Bind(&loadTestPropertyReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("request param is incorrect; %+v", loadTestPropertyReq),
			})
		}

		loadTestId, err := services.ExecuteLoadTest(loadTestPropertyReq)

		if err != nil {
			log.Printf("error while executing load test; %+v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": fmt.Sprintf("sorry, internal server error while executing load test; %+v", loadTestPropertyReq),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"testId":  loadTestId,
			"message": "success",
		})
	}
}

func InstallLoadGeneratorHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		loadInstallReq := api.LoadEnvReq{}

		if err := c.Bind(&loadInstallReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass me correct body; %v", loadInstallReq),
			})

		}

		if err := loadInstallReq.Validate(); err != nil {
			log.Printf("error while execute [InstallLoadGeneratorHandler()]; %s\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("if you install on remote, pass nsId, mcisId and username"),
			})
		}

		err := services.InstallLoadGenerator(loadInstallReq)

		if err != nil {
			log.Printf("error while executing load test; %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}
