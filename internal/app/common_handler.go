package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Id AntServerReadiness
// @Summary Check CB-Ant API server readiness
// @Description This endpoint checks if the CB-Ant API server is ready by verifying the status of both the load service and the cost service. If either service is unavailable, it returns a 503 status indicating the server is not ready.
// @Tags [Server Health]
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "CM-Ant API server is ready"
// @Failure 503 {object} map[string]string "CB-Ant API server is not ready"
// @Router /readyz [get]
func (s *AntServer) readyz(c echo.Context) error {
	err := s.services.loadService.Readyz()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, map[string]string{
			"message": "CM-Ant API server is not ready",
		})
	}

	err = s.services.costService.Readyz()

	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, map[string]string{
			"message": "CM-Ant API server is not ready",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "CM-Ant API server is ready",
	})
}
