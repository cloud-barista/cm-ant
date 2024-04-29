package handler

import (
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/labstack/echo/v4"
)

// GetAllLoadEnvironments
// @Id 				LoadEnvironments
// @Summary			Get the list of load test environments
// @Description		Get all of the load test environments
// @Tags			[Load Test Environment]
// @Accept			json
// @Produce			json
// @Success			200	{object}	[]api.LoadEnvRes
// @Failure			500	{object}	string
// @Router			/ant/api/v1/env [get]
func GetAllLoadEnvironments() echo.HandlerFunc {
	return func(c echo.Context) error {

		result, err := services.GetAllLoadEnvironments()

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

func DeleteLoadEnvironments() echo.HandlerFunc {
	return func(c echo.Context) error {
		envId := c.Param("envId")

		if envId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"message": "remote connection id is empty",
			})
		}

		err := services.DeleteLoadEnvironment(envId)

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
