package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/labstack/echo/v4"
)

// getAllLoadGeneratorInstallInfo handler function that retrieves all load generator installation information.
// @Id GetAllLoadGeneratorInstallInfo
// @Summary Get All Load Generator Install Info
// @Description Retrieve a list of all installed load generators with pagination support.
// @Tags LoadGeneratorManagement
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination (default 1)"
// @Param size query int false "Number of items per page (default 10, max 10)"
// @Param status query string false "Filter by status"
// @Success 200 {object} app.AntResponse[load.GetAllLoadGeneratorInstallInfoResult] "Successfully retrieved monitoring agent information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve monitoring agent information"
// @Router /api/v1/load/generators [get]
func (s *AntServer) getAllLoadGeneratorInstallInfo(c echo.Context) error {
	var req GetAllLoadGeneratorInstallInfoReq
	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request parameters")
	}
	if req.Size < 1 || req.Size > 10 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}

	arg := load.GetAllLoadGeneratorInstallInfoParam{
		Page:   req.Page,
		Size:   req.Size,
		Status: req.Status,
	}

	result, err := s.services.loadService.GetAllLoadGeneratorInstallInfo(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to retrieve monitoring agent information")
	}

	return successResponse(c, "Successfully retrieved monitoring agent information", result)
}

// installLoadGenerator handler function that handles a load generator installation request.
// @Id InstallLoadGenerator
// @Summary Install Load Generator
// @Description Install a load generator either locally or remotely.
// @Tags LoadGeneratorManagement
// @Accept json
// @Produce json
// @Param body body app.InstallLoadGeneratorReq true "Load Generator Installation Request"
// @Success 200 {object} app.AntResponse[load.LoadGeneratorInstallInfoResult] "Successfully installed load generator"
// @Failure 400 {object} app.AntResponse[string] "Load generator installation info is not correct.| available install locations are remote or local."
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/generators [post]
func (s *AntServer) installLoadGenerator(c echo.Context) error {
	var req InstallLoadGeneratorReq

	utils.LogInfo("Received request to install load generator")

	if err := c.Bind(&req); err != nil {
		utils.LogError("Failed to bind request:", err)
		return errorResponse(http.StatusBadRequest, "load generator installation info is not correct.")
	}

	if req.InstallLocation == "" ||
		(req.InstallLocation != constant.Remote && req.InstallLocation != constant.Local) {
		utils.LogError("Invalid install location:", req.InstallLocation)
		return errorResponse(http.StatusBadRequest, "available install locations are remote or local.")
	}

	utils.LogInfo("Calling service layer to install load generator")

	// call service layer install load generator
	param := load.InstallLoadGeneratorParam{
		InstallLocation: req.InstallLocation,
		Coordinates:     []string{"37.53/127.02"},
	}
	result, err := s.services.loadService.InstallLoadGenerator(param)

	if err != nil {
		utils.LogError("Error installing load generator:", err)
		return errorResponse(http.StatusBadRequest, err.Error())
	}

	utils.LogInfo("Load generator installed successfully")

	return successResponse(
		c,
		"load generator is successfully installed",
		result,
	)
}

// uninstallLoadGenerator handler function that handles a load generator uninstallation request.
// @Id UninstallLoadGenerator
// @Summary Uninstall Load Generator
// @Description Uninstall a previously installed load generator.
// @Tags LoadGeneratorManagement
// @Accept json
// @Produce json
// @Param loadGeneratorInstallInfoId path string true "load generator install info id"
// @Success 200 {object} app.AntResponse[string] "Successfully uninstall load generator"
// @Failure 400 {object} app.AntResponse[string] "Load generator installation info id must be set."
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/generators/{loadGeneratorInstallInfoId} [delete]
func (s *AntServer) uninstallLoadGenerator(c echo.Context) error {
	loadGeneratorInstallInfoId := c.Param("loadGeneratorInstallInfoId")

	if strings.TrimSpace(loadGeneratorInstallInfoId) == "" {
		return errorResponse(http.StatusBadRequest, "load generator installation info id must be set.")
	}

	arg := load.UninstallLoadGeneratorParam{
		LoadGeneratorInstallInfoId: loadGeneratorInstallInfoId,
	}
	err := s.services.loadService.UninstallLoadGenerator(arg)

	if err != nil {
		return errorResponse(http.StatusBadRequest, err.Error())
	}

	return successResponse(
		c,
		fmt.Sprintf("Successfully uninstall load generator: %s", loadGeneratorInstallInfoId),
		"done",
	)
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
