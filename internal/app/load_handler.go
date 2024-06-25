package app

import (
	"net/http"

	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/labstack/echo/v4"
)

func (s *AntServer) getAllLoadEnvironments(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) installLoadTester(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) uninstallLoadTester(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) runLoadTest(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) stopLoadTest(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getLoadTestResult(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getLoadTestMetrics(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getAllLoadConfig(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getLoadConfig(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getAllLoadExecutionState(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getLoadExecutionState(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

// installMonitoringAgent handler function that handles a monitoring request request to collect metric.
// @Id				InstallMonitoringAgent
// @Summary Install Metrics Monitoring Agent
// @Description Install a monitoring agent on specific MCIS.
// @Tags MonitoringAgentManagement
// @Accept json
// @Produce json
// @Param body body app.MonitoringAgentInstallationReq true "Monitoring Agent Installation Request"
// @Success 200 {object} app.AntResponse[load.MonitoringAgentInstallationResult] "Successfully installed monitoring agent"
// @Failure 400 {object} app.AntResponse[string] "Monitoring agent installation info is not correct."
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/monitoring/agent [post]
func (s *AntServer) installMonitoringAgent(c echo.Context) error {
	var req MonitoringAgentInstallationReq

	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "monitoring agent installation info is not correct.")
	}

	arg := load.MonitoringAgentInstallationParams{
		NsId:   req.NsId,
		McisId: req.McisId,
	}

	result, err := s.services.loadService.InstallMonitoringAgent(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, err.Error())
	}

	return successResponse(
		c,
		"monitoring agent is successfully installed",
		result,
	)
}

func (s *AntServer) getAllAgentInstallInfo(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) uninstallAgent(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}
