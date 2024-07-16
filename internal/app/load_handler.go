package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/labstack/echo/v4"
)

const (
	seoul = "37.53/127.02"
)

func (s *AntServer) readyz(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "CM-Ant API server is running",
	})
}

// getAllLoadGeneratorInstallInfo handler function that retrieves all load generator installation information.
// @Id GetAllLoadGeneratorInstallInfo
// @Summary Get All Load Generator Install Info
// @Description Retrieve a list of all installed load generators with pagination support.
// @Tags [Load Generator Management]
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
// @Tags [Load Generator Management]
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
		Coordinates:     []string{seoul},
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
// @Tags [Load Generator Management]
// @Accept json
// @Produce json
// @Param loadGeneratorInstallInfoId path string true "load generator install info id"
// @Success 200 {object} app.AntResponse[string] "Successfully uninstall load generator"
// @Failure 400 {object} app.AntResponse[string] "Load generator installation info id must be set."
// @Failure 400 {object} app.AntResponse[string] "Load generator install info id must be number."
// @Failure 500 {object} app.AntResponse[string] "ant server has got error. please try again."
// @Router /api/v1/load/generators/{loadGeneratorInstallInfoId} [delete]
func (s *AntServer) uninstallLoadGenerator(c echo.Context) error {
	loadGeneratorInstallInfoId := c.Param("loadGeneratorInstallInfoId")

	if strings.TrimSpace(loadGeneratorInstallInfoId) == "" {
		return errorResponse(http.StatusBadRequest, "Load generator installation info id must be set.")
	}

	cvt, err := strconv.Atoi(loadGeneratorInstallInfoId)
	if err != nil {
		errorResponse(http.StatusBadRequest, "Load generator install info id must be number.")
	}

	arg := load.UninstallLoadGeneratorParam{
		LoadGeneratorInstallInfoId: uint(cvt),
	}

	err = s.services.loadService.UninstallLoadGenerator(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Ant server has got error. please try again.")
	}

	return successResponse(
		c,
		fmt.Sprintf("Successfully uninstall load generator: %s", loadGeneratorInstallInfoId),
		"done",
	)
}

// runLoadTest handler function that initiates a load test.
// @Id RunLoadTest
// @Summary Run Load Test
// @Description Start a load test using the provided load test configuration.
// @Tags [Load Test Execution Management]
// @Accept json
// @Produce json
// @Param body body app.RunLoadTestReq true "Run Load Test Request"
// @Success 200 {object} app.AntResponse[string] "{loadTestKey}"
// @Failure 400 {object} app.AntResponse[string] "load test running info is not correct."
// @Failure 400 {object} app.AntResponse[string] "load test install location is invalid."
// @Failure 500 {object} app.AntResponse[string] "ant server has got error. please try again."
// @Router /api/v1/load/tests/run [post]
func (s *AntServer) runLoadTest(c echo.Context) error {
	var req RunLoadTestReq

	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "load test running info is not correct.")
	}

	if req.LoadGeneratorInstallInfoId != uint(0) {
		req.InstallLoadGenerator = InstallLoadGeneratorReq{}
	} else if req.InstallLoadGenerator.InstallLocation != constant.Local &&
		req.InstallLoadGenerator.InstallLocation != constant.Remote {
		return errorResponse(http.StatusBadRequest, "load test install location is invalid.")
	}

	var https []load.RunLoadTestHttpParam
	for _, h := range req.HttpReqs {
		hh := load.RunLoadTestHttpParam{
			Method:   h.Method,
			Protocol: h.Protocol,
			Hostname: h.Hostname,
			Port:     h.Port,
			Path:     h.Path,
			BodyData: h.BodyData,
		}

		https = append(https, hh)
	}

	arg := load.RunLoadTestParam{

		InstallLoadGenerator: load.InstallLoadGeneratorParam{
			InstallLocation: req.InstallLoadGenerator.InstallLocation,
			Coordinates:     []string{seoul},
		},
		LoadGeneratorInstallInfoId: req.LoadGeneratorInstallInfoId,
		TestName:                   req.TestName,
		VirtualUsers:               req.VirtualUsers,
		Duration:                   req.Duration,
		RampUpTime:                 req.RampUpTime,
		RampUpSteps:                req.RampUpSteps,
		Hostname:                   req.Hostname,
		Port:                       req.Port,
		AgentInstalled:             req.AgentInstalled,
		AgentHostname:              req.AgentHostname,
		HttpReqs:                   https,
	}

	loadTestKey, err := s.services.loadService.RunLoadTest(arg)

	if err != nil {
		return errorResponse(http.StatusBadRequest, err.Error())
	}

	return successResponse(
		c,
		fmt.Sprintf("Successfully run load test. Load test key: %s", loadTestKey),
		loadTestKey,
	)
}

// stopLoadTest handler function that stops a running load test.
// @Id StopLoadTest
// @Summary Stop Load Test
// @Description Stop a running load test using the provided load test key.
// @Tags [Load Test Execution Management]
// @Accept json
// @Produce json
// @Param body body app.StopLoadTestReq true "Stop Load Test Request"
// @Success 200 {object} app.AntResponse[string] "done"
// @Failure 400 {object} app.AntResponse[string] "load test running info is not correct."
// @Failure 500 {object} app.AntResponse[string] "ant server has got error. please try again."
// @Router /api/v1/load/tests/stop [post]
func (s *AntServer) stopLoadTest(c echo.Context) error {
	var req StopLoadTestReq

	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "load test stop info is not correct.")
	}

	if strings.TrimSpace(req.LoadTestKey) == "" {
		return errorResponse(http.StatusBadRequest, "load test stop info is not correct.")
	}

	arg := load.StopLoadTestParam{
		LoadTestKey: req.LoadTestKey,
	}

	err := s.services.loadService.StopLoadTest(arg)

	if err != nil {
		return errorResponse(http.StatusBadRequest, err.Error())
	}

	return successResponse(
		c,
		fmt.Sprintf("Successfully stop load generator. Load test key: %s", req.LoadTestKey),
		"done",
	)
}

func (s *AntServer) getLoadTestResult(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

func (s *AntServer) getLoadTestMetrics(c echo.Context) error {

	return c.JSON(http.StatusOK, c.Request().RequestURI)
}

// getAllLoadTestExecutionInfos handler function that retrieves all load test execution information.
// @Id GetAllLoadTestExecutionInfos
// @Summary Get All Load Test Execution Information
// @Description Retrieve a list of all load test execution information with pagination support.
// @Tags [Load Test Execution Management]
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination (default 1)"
// @Param size query int false "Number of items per page (default 10, max 10)"
// @Success 200 {object} app.AntResponse[load.GetAllLoadTestExecutionInfosResult] "Successfully retrieved load test execution information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve all load test execution information"
// @Router /api/v1/load/tests/infos [get]
func (s *AntServer) getAllLoadTestExecutionInfos(c echo.Context) error {
	var req GetAllLoadTestExecutionHistoryReq
	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request parameters")
	}
	if req.Size < 1 || req.Size > 10 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}

	arg := load.GetAllLoadTestExecutionInfosParam{
		Page: req.Page,
		Size: req.Size,
	}

	result, err := s.services.loadService.GetAllLoadTestExecutionInfos(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to retrieve all load test execution information")
	}

	return successResponse(c, "Successfully retrieved load test execution information", result)
}

// getLoadTestExecutionInfo handler function that retrieves a specific load test execution state by key.
// @Id GetLoadTestExecutionInfo
// @Summary Get Load Test Execution State
// @Description Retrieve the load test execution state information for a specific load test key.
// @Tags [Load Test Execution Management]
// @Accept json
// @Produce json
// @Param loadTestKey path string true "Load test key"
// @Success 200 {object} app.AntResponse[load.LoadTestExecutionInfoResult] "Successfully retrieved load test execution state information"
// @Failure 400 {object} app.AntResponse[string] "Load test key must be set."
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test execution state information"
// @Router /api/v1/load/tests/infos/{loadTestKey} [get]
func (s *AntServer) getLoadTestExecutionInfo(c echo.Context) error {
	loadTestKey := c.Param("loadTestKey")

	if strings.TrimSpace(loadTestKey) == "" {
		return errorResponse(http.StatusBadRequest, "Load test key must be set.")
	}

	arg := load.GetLoadTestExecutionInfoParam{
		LoadTestKey: loadTestKey,
	}

	result, err := s.services.loadService.GetLoadTestExecutionInfo(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to retrieve load test execution state information")
	}

	return successResponse(c, "Successfully retrieved load test execution state information", result)
}

// getAllLoadTestExecutionState handler function that retrieves all load test execution states.
// @Id GetAllLoadTestExecutionState
// @Summary Get All Load Test Execution State
// @Description Retrieve a list of all load test execution states with pagination support.
// @Tags [Load Test State Management]
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination (default 1)"
// @Param size query int false "Number of items per page (default 10, max 10)"
// @Param loadTestKey query string false "Filter by load test key"
// @Param executionStatus query string false "Filter by execution status"
// @Success 200 {object} app.AntResponse[load.GetAllLoadTestExecutionStateResult] "Successfully retrieved load test execution state information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test execution state information"
// @Router /api/v1/load/tests/state [get]
func (s *AntServer) getAllLoadTestExecutionState(c echo.Context) error {
	var req GetAllLoadTestExecutionStateReq
	if err := c.Bind(&req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request parameters")
	}
	if req.Size < 1 || req.Size > 10 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}

	arg := load.GetAllLoadTestExecutionStateParam{
		Page:            req.Page,
		Size:            req.Size,
		LoadTestKey:     req.LoadTestKey,
		ExecutionStatus: req.ExecutionStatus,
	}

	result, err := s.services.loadService.GetAllLoadTestExecutionState(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to retrieve all load test execution state information")
	}

	return successResponse(c, "Successfully retrieved load test execution state information", result)
}

// getLoadTestExecutionState handler function that retrieves a load test execution state by key.
// @Id GetLoadTestExecutionState
// @Summary Get Load Test Execution State
// @Description Retrieve a load test execution state by load test key.
// @Tags [Load Test State Management]
// @Accept json
// @Produce json
// @Param loadTestKey path string true "Load test key"
// @Success 200 {object} app.AntResponse[load.LoadTestExecutionStateResult] "Successfully retrieved load test execution state information"
// @Failure 400 {object} app.AntResponse[string] "Load test key must be set."
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test execution state information"
// @Router /api/v1/load/tests/state/{loadTestKey} [get]
func (s *AntServer) getLoadTestExecutionState(c echo.Context) error {
	loadTestKey := c.Param("loadTestKey")

	if strings.TrimSpace(loadTestKey) == "" {
		return errorResponse(http.StatusBadRequest, "Load test key must be set.")
	}

	arg := load.GetLoadTestExecutionStateParam{
		LoadTestKey: loadTestKey,
	}

	result, err := s.services.loadService.GetLoadTestExecutionState(arg)

	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to retrieve load test execution state information")
	}

	return successResponse(c, "Successfully retrieved load test execution state information", result)

}

// installMonitoringAgent handler function that handles a monitoring request request to collect metric.
// @Id				InstallMonitoringAgent
// @Summary Install Metrics Monitoring Agent
// @Description Install a monitoring agent on specific MCIS.
// @Tags [Monitoring Agent Management]
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
// @Tags [Monitoring Agent Management]
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
// @Tags           [Monitoring Agent Management]
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
