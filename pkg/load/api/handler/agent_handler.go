package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/services"

	"github.com/labstack/echo/v4"
)

// InstallAgent
// @Id				InstallAgent
// @Summary			Install jmeter perfmon agent for metrics collection
// @Description		Install an agent to collect server metrics during load testing such as CPU and memory.
// @Tags			[Agent - for Development]
// @Accept			json
// @Produce			json
// @Param			loadEnvReq 		body 	api.AntTargetServerReq 			true 		"agent target server req"
// @Success			200	{object}			map[string]string								`{ "message": "success", "agentId":  agentId }`
// @Failure			400	{object}			map[string]string								`{ "message": "nsId and mcisId must set", }`
// @Failure			400	{object}			map[string]string								`{ "message": "pass me correct body;", }`
// @Failure			500	{object}			map[string]string								"internal server error"
// @Router			/ant/api/v1/load/agent 		[post]
func InstallAgent() echo.HandlerFunc {
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

		err := services.InstallAgent(agentReq)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": err,
			})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

// GetAllAgentInstallInfo
// @Id				GetAllAgentInstallInfo
// @Summary			Get all agent installation information
// @Description		Get all agent installation nsId, mcisId, vmId, status.
// @Tags			[Agent - for Development]
// @Accept			json
// @Produce			json
// @Success			200	{object}		map[string]string					`{ "message": "success", "result":  result, }`
// @Failure			500	{object}		map[string]string					`{ "message": "something went wrong.try again.", }`
// @Router			/ant/api/v1/load/agent 		[get]
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

// UninstallAgent
// @Id				UninstallAgent
// @Summary			Uninstall jmeter perfmon agent for metrics collection
// @Description		Uninstall an agent to collect server metrics during load testing such as CPU and memory.
// @Tags			[Agent - for Development]
// @Accept			json
// @Produce			json
// @Param			agentInstallInfoId 			path 	string 			true 		"agent installation info id"
// @Success			200	{object}		map[string]string					`{ "message": "success", }`
// @Failure		400	{object}			map[string]string					`{ "message": "pass me correct agentId", }`
// @Failure			500	{object}		map[string]string					`{ "message": "error message", }`
// @Router			/ant/api/v1/load/agent/{agentInstallInfoId} 		[delete]
func UninstallAgent() echo.HandlerFunc {
	return func(c echo.Context) error {
		agentId := c.Param("agentInstallInfoId")

		if agentId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "pass me correct agentId",
			})
		}

		err := services.UninstallAgent(agentId)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": fmt.Sprintln("something went wrong", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}
