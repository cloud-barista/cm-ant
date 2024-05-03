package handler

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"net/http"

	"github.com/labstack/echo/v4"
)

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
		return nil
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
