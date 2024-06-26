package app

import (
	"net/http"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/labstack/echo/v4"
)

func (s *AntServer) getAllLoadEnvironments(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) installLoadGenerator(c echo.Context) error {
	var req InstallLoadGeneratorReq

	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "load generator installation info is not correct.")
	}

	if req.InstallLocation == "" ||
		(req.InstallLocation != constant.Remote && req.InstallLocation != constant.Local) {
		return errorResponse(http.StatusBadRequest, "available install locations are remote or local.")
	}

	if req.LoadGeneratorType != constant.Jmeter {
		req.LoadGeneratorType = constant.Jmeter
	}

	// call service layer install load generator

	return successResponse(
		c,
		"load generator is successfully installed",
		"done",
	)
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
// @Router /api/v1/load/monitoring/agent/install [post]
func (s *AntServer) installMonitoringAgent(c echo.Context) error {
	var req MonitoringAgentInstallationReq

	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "monitoring agent installation info is not correct.")
	}

	arg := load.MonitoringAgentInstallationParams{
		NsId:   req.NsId,
		McisId: req.McisId,
		VmIds:  req.VmIds,
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

// getAllMonitoringAgentInfos handler function that retrieves monitoring agent information.
// @Id				GetAllMonitoringAgentInfos
// @Summary Retrieve Monitoring Agent Information
// @Description Retrieve monitoring agent information based on specified criteria.
// @Tags MonitoringAgentManagement
// @Accept json
// @Produce json
// @Param nsId query string false "Namespace ID" default:""
// @Param mcisId query string false "MCIS ID" default:""
// @Param vmId query string false "VM ID" default:""
// @Param size query integer false "Number of results per page" default:"10"
// @Param page query integer false "Page number for pagination" default:"1"
// @Success 200 {object} app.AntResponse[load.GetAllMonitoringAgentInfoResult] "Successfully retrieved monitoring agent information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/monitoring/agents [get]
func (s *AntServer) getAllMonitoringAgentInfos(c echo.Context) error {
	var req GetAllMonitoringAgentInfosReq
	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request parameters")
	}
	if req.Size < 1 || req.Size > 10 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}

	arg := load.GetAllMonitoringAgentInfosParam{
		Page:   req.Page,
		Size:   req.Size,
		NsId:   req.NsId,
		McisId: req.McisId,
		VmId:   req.VmId,
	}

	result, err := s.services.loadService.GetAllMonitoringAgentInfos(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to retrieve monitoring agent information")
	}

	return successResponse(c, "Successfully retrieved monitoring agent information", result)
}

// uninstallMonitoringAgent handler function that initiates the uninstallation of monitoring agents.
// @Id             UninstallMonitoringAgent
// @Summary        Uninstall Monitoring Agents
// @Description    Uninstall monitoring agents from specified VMs or all VMs in an MCIS.
// @Tags           MonitoringAgentManagement
// @Accept         json
// @Produce        json
// @Param body body app.MonitoringAgentInstallationReq true "Monitoring Agent Uninstallation Request"
// @Success 200 {object} app.AntResponse[int64] "Number of affected results"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/monitoring/agents/uninstall [post]
func (s *AntServer) uninstallMonitoringAgent(c echo.Context) error {
	var req MonitoringAgentInstallationReq

	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "monitoring agent uninstallation info is not correct.")
	}

	arg := load.MonitoringAgentInstallationParams{
		NsId:   req.NsId,
		McisId: req.McisId,
		VmIds:  req.VmIds,
	}

	affectedResults, err := s.services.loadService.UninstallMonitoringAgent(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, err.Error())
	}

	return successResponse(
		c,
		"monitoring agent is successfully uninstalled",
		affectedResults,
	)
}
