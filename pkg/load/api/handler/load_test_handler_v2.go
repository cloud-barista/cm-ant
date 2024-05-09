package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

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

		err := services.InstallLoadTesterV2(&loadTesterReq)

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
