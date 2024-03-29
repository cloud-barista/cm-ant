package handler

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func RegisterRemoteConnection() echo.HandlerFunc {
	return func(c echo.Context) error {
		remoteConnectionReq := api.RemoteConnectionReq{}

		if err := c.Bind(&remoteConnectionReq); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("pass me correct body; %v", remoteConnectionReq),
			})
		}

		if err := remoteConnectionReq.Validate(); err != nil {
			log.Printf("error while execute [RegisterRemoteConnection()]; %s\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": fmt.Sprintf("%s", err),
			})
		}

		err := services.RegisterRemoteConnection(remoteConnectionReq)

		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}

func GetAllRemoteConnection() echo.HandlerFunc {
	return func(c echo.Context) error {

		result, err := services.GetAllRemoteConnection()

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

func DeleteRemoteConnection() echo.HandlerFunc {
	return func(c echo.Context) error {
		remoteConnectionId := c.Param("remoteConnectionId")

		if remoteConnectionId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "remote connection id is empty",
			})
		}

		err := services.DeleteRemoteConnection(remoteConnectionId)

		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
		})
	}
}
