package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/services"

	"github.com/labstack/echo/v4"
)

func InstallAgentV2() echo.HandlerFunc {
	return func(c echo.Context) error {
		agentReq := api.AntTargetServerReq{}

		if err := c.Bind(&agentReq); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass me correct body; %v\n", agentReq),
			})

		}

		if agentReq.NsId == "" || agentReq.McisId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "nsId and mcisId must set",
			})
		}

		err := services.InstallAgentV2(agentReq)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": err,
			})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

func GetAllAgentInstallInfo() echo.HandlerFunc {
	return func(c echo.Context) error {

		result, err := services.GetAllAgentInstallInfo()

		if err != nil {
			log.Println(err)
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

func UninstallAgentV2() echo.HandlerFunc {
	return func(c echo.Context) error {
		agentId := c.Param("agentInstallInfoId")

		if agentId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "pass me correct agentId",
			})
		}

		err := services.UninstallAgentV2(agentId)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintln("something went wrong", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}
