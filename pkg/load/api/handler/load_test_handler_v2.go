package handler

import (
	"context"
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
		loadTesterReq := api.LoadTesterReq{}

		if err := c.Bind(&loadTesterReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "request body is incorrect",
			})

		}

		if loadTesterReq.NsId == "" || loadTesterReq.McisId == "" || loadTesterReq.VmId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "ns, mcis and vm id is essential",
			})
		}

		_, err := services.InstallLoadTesterV2(&loadTesterReq)

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
