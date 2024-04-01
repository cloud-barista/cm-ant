package handler

import (
	"github.com/cloud-barista/cm-ant/pkg/load/api"
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func GetAllRemoteConnection() echo.HandlerFunc {
	return func(c echo.Context) error {

		result, err := services.GetAllRemoteConnection()

		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
				"message": "something went wrong.try again.",
			})

		}

		var responseEnv []api.LoadEnvRes

		for _, loadEnv := range result {
			var load api.LoadEnvRes
			load.EnvId = loadEnv.ID
			load.InstallLocation = loadEnv.InstallLocation
			load.RemoteConnectionType = loadEnv.RemoteConnectionType
			load.Username = loadEnv.Username
			load.PublicIp = loadEnv.PublicIp
			load.Cert = loadEnv.Cert
			load.NsId = loadEnv.NsId
			load.McisId = loadEnv.McisId

			responseEnv = append(responseEnv, load)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message": "success",
			"result":  responseEnv,
		})
	}
}

func DeleteRemoteConnection() echo.HandlerFunc {
	return func(c echo.Context) error {
		envId := c.Param("envId")

		if envId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "remote connection id is empty",
			})
		}

		err := services.DeleteRemoteConnection(envId)

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
