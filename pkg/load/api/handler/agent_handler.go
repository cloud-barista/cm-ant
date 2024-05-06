package handler

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"net/http"

	"github.com/labstack/echo/v4"
)

// InstallAgent
// @Id				InstallAgent
// @Summary			Install jmeter perfmon agent for metrics collection
// @Description		Install an agent to collect server metrics during load testing such as CPU and memory.
// @Tags			[Agent - for Development]
// @Accept			json
// @Produce			json
// @Param			loadEnvReq 		body 	api.AgentReq 			true 		"agent install request"
// @Success			200	{object}			map[string]string					`{ "message": "success", "agentId":  agentId }`
// @Failure			400	{object}			string								"request body is not correct"
// @Failure			500	{object}			string								"internal server error"
// @Router			/ant/api/v1/load/agent 		[post]
func InstallAgent() echo.HandlerFunc {
	return func(c echo.Context) error {
		agentReq := api.AgentReq{}

		if err := c.Bind(&agentReq); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass me correct body; %v\n", agentReq),
			})

		}

		agentId, err := services.InstallAgent(agentReq)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintln("something went wrong", err),
			})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"agentId": agentId,
			"message": "success",
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
// @Param			agentId 			path 	string 			true 		"agentId"
// @Success			200	{object}		map[string]string					`{ "message": "success" }`
// @Failure			400	{object}		string								"request body is not correct"
// @Failure			500	{object}		string								"internal server error"
// @Router			/ant/api/v1/load/agent/{agentId} 		[delete]
func UninstallAgent() echo.HandlerFunc {
	return func(c echo.Context) error {
		agentId := c.Param("agentId")

		if agentId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass me correct agentId\n"),
			})
		}

		err := services.UninstallAgent(agentId)

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

func MockMigration() echo.HandlerFunc {
	return func(c echo.Context) error {
		err := services.MockMigration("")
		if err != nil {
			return err
		}
		return nil
	}
}
