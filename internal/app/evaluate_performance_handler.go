package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

const (
	seoul = "37.53/127.02"
)

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
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
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
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve monitoring agent information")
	}

	return successResponseJson(c, "Successfully retrieved monitoring agent information", result)
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

	log.Info().Msg("Received request to install load generator")

	if err := c.Bind(&req); err != nil {
		log.Error().Msgf("Failed to bind request:", err)
		return errorResponseJson(http.StatusBadRequest, "load generator installation info is not correct.")
	}

	if req.InstallLocation == "" ||
		(req.InstallLocation != constant.Remote && req.InstallLocation != constant.Local) {
		log.Error().Msgf("Invalid install location:", req.InstallLocation)
		return errorResponseJson(http.StatusBadRequest, "available install locations are remote or local.")
	}

	log.Info().Msgf("Calling service layer to install load generator")

	// call service layer install load generator
	param := load.InstallLoadGeneratorParam{
		InstallLocation: req.InstallLocation,
		Coordinates:     []string{seoul},
	}
	result, err := s.services.loadService.InstallLoadGenerator(param)

	if err != nil {
		log.Error().Msgf("Error installing load generator:", err)
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	log.Info().Msg("Load generator installed successfully")

	return successResponseJson(
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
		return errorResponseJson(http.StatusBadRequest, "Load generator installation info id must be set.")
	}

	cvt, err := strconv.Atoi(loadGeneratorInstallInfoId)
	if err != nil {
		errorResponseJson(http.StatusBadRequest, "Load generator install info id must be number.")
	}

	arg := load.UninstallLoadGeneratorParam{
		LoadGeneratorInstallInfoId: uint(cvt),
	}

	err = s.services.loadService.UninstallLoadGenerator(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Ant server has got error. please try again.")
	}

	return successResponseJson(
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
		return errorResponseJson(http.StatusBadRequest, "load test running info is not correct.")
	}

	if req.LoadGeneratorInstallInfoId != uint(0) {
		req.InstallLoadGenerator = InstallLoadGeneratorReq{}
	} else if req.InstallLoadGenerator.InstallLocation != constant.Local &&
		req.InstallLoadGenerator.InstallLocation != constant.Remote {
		return errorResponseJson(http.StatusBadRequest, "invalid load test install location")
	}

	if strings.TrimSpace(req.TestName) == "" {
		req.TestName = uuid.New().String()
	}

	if v, err := strconv.Atoi(strings.TrimSpace(req.VirtualUsers)); err != nil && (v < 1 || v > 100) {
		log.Error().Msgf("virtual user count is invalid")
		return errorResponseJson(http.StatusBadRequest, "virtual user  is not correct. the range must be in 1 to 100")
	}

	if v, err := strconv.Atoi(strings.TrimSpace(req.Duration)); err != nil && (v < 1 || v > 300) {
		log.Error().Msgf("duration is invalid")
		return errorResponseJson(http.StatusBadRequest, "duration is not correct. the range must be in 1 to 300")
	}

	if v, err := strconv.Atoi(strings.TrimSpace(req.RampUpTime)); err != nil && (v < 1 || v > 60) {
		log.Error().Msgf("ramp up time is invalid")
		return errorResponseJson(http.StatusBadRequest, "ramp up time is not correct. the range must be in 1 to 60")
	}

	if v, err := strconv.Atoi(strings.TrimSpace(req.RampUpSteps)); err != nil && (v < 1 || v > 20) {
		log.Error().Msgf("ramp up steps is invalid")
		return errorResponseJson(http.StatusBadRequest, "ramp up steps is not correct. the range must be in 1 to 20")
	}

	if len(req.HttpReqs) == 0 {
		log.Error().Msgf("http request have to have at least one; %s", req.TestName)
		return errorResponseJson(http.StatusBadRequest, "http request have to have at least one or more")
	}

	var https []load.RunLoadTestHttpParam
	for _, h := range req.HttpReqs {
		if h.Method == "" {
			log.Error().Msgf("method for load test is empty. cannot start load test; %s", req.TestName)
			return errorResponseJson(http.StatusBadRequest, "load test running info is not correct.; Method")
		}

		if h.Protocol == "" {
			h.Protocol = "http"
		}

		if h.Hostname == "" {
			log.Error().Msgf("hostname for load test is empty. cannot start load test; %s", req.TestName)
			return errorResponseJson(http.StatusBadRequest, "load test running info is not correct.; Hostname")
		}

		if t, err := strconv.Atoi(h.Port); err != nil && (t < 1 || t > 65353) {
			log.Error().Msgf("port range is not valid. check the range of ports")
			return errorResponseJson(http.StatusBadRequest, "load test running info is not correct.; Port")
		}

		hh := load.RunLoadTestHttpParam{
			Method:   strings.TrimSpace(h.Method),
			Protocol: strings.TrimSpace(h.Protocol),
			Hostname: strings.TrimSpace(h.Hostname),
			Port:     strings.TrimSpace(h.Port),
			Path:     strings.TrimSpace(h.Path),
			BodyData: strings.TrimSpace(h.BodyData),
		}

		https = append(https, hh)
	}

	arg := load.RunLoadTestParam{
		InstallLoadGenerator: load.InstallLoadGeneratorParam{
			InstallLocation: req.InstallLoadGenerator.InstallLocation,
			Coordinates:     []string{seoul},
		},
		LoadGeneratorInstallInfoId: req.LoadGeneratorInstallInfoId,
		TestName:                   strings.TrimSpace(req.TestName),
		VirtualUsers:               strings.TrimSpace(req.VirtualUsers),
		Duration:                   strings.TrimSpace(req.Duration),
		RampUpTime:                 strings.TrimSpace(req.RampUpTime),
		RampUpSteps:                strings.TrimSpace(req.RampUpSteps),

		CollectAdditionalSystemMetrics: req.CollectAdditionalSystemMetrics,
		AgentHostname:                  strings.TrimSpace(req.AgentHostname),

		NsId:  strings.TrimSpace(req.NsId),
		MciId: strings.TrimSpace(req.MciId),
		VmId:  strings.TrimSpace(req.VmId),

		HttpReqs: https,
	}

	loadTestKey, err := s.services.loadService.RunLoadTest(arg)

	if err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	return successResponseJson(
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
		return errorResponseJson(http.StatusBadRequest, "load test stop info is not correct.")
	}

	if strings.TrimSpace(req.LoadTestKey) == "" ||
		(strings.TrimSpace(req.NsId) == "" || strings.TrimSpace(req.MciId) == "" || strings.TrimSpace(req.VmId) == "") {
		return errorResponseJson(http.StatusBadRequest, "load test stop info is not correct.")
	}

	arg := load.StopLoadTestParam{
		LoadTestKey: req.LoadTestKey,
	}

	err := s.services.loadService.StopLoadTest(arg)

	if err != nil {
		return errorResponseJson(http.StatusBadRequest, err.Error())
	}

	return successResponseJson(
		c,
		fmt.Sprintf("Successfully stop load generator. Load test key: %s", req.LoadTestKey),
		"done",
	)
}

// getLoadTestResult handler function that retrieves a specific load test result.
// @Id GetLoadTestResult
// @Summary Get load test result
// @Description Retrieve load test result based on provided parameters.
// @Tags [Load Test Result]
// @Accept json
// @Produce json
// @Param loadTestKey query string true "Load test key"
// @Param format query string false "Result format (normal or aggregate)"
// @Success 200 {object} app.JsonResult{[normal]=app.AntResponse[[]load.ResultSummary],[aggregate]=app.AntResponse[[]load.LoadTestStatistics]} "Successfully retrieved load test metrics"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test result"
// @Router /api/v1/load/tests/result [get]
func (s *AntServer) getLoadTestResult(c echo.Context) error {
	var req GetLoadTestResultReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	if strings.TrimSpace(req.LoadTestKey) == "" {
		return errorResponseJson(http.StatusBadRequest, "pass correct load test key")
	}

	if req.Format == "" {
		req.Format = constant.Normal
	} else if req.Format != constant.Normal && req.Format != constant.Aggregate {
		req.Format = constant.Normal
	}

	arg := load.GetLoadTestResultParam{
		LoadTestKey: req.LoadTestKey,
		Format:      req.Format,
	}

	result, err := s.services.loadService.GetLoadTestResult(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test result")
	}

	return successResponseJson(c, "Successfully retrieved load test result", result)
}

// getLoadTestMetrics handler function that retrieves metrics for a specific load test.
// @Id GetLoadTestMetrics
// @Summary Get load test metrics
// @Description Retrieve load test metrics based on provided parameters.
// @Tags [Load Test Result]
// @Accept json
// @Produce json
// @Param loadTestKey query string true "Load test key"
// @success 200 {object} app.AntResponse[[]load.MetricsSummary] "Successfully retrieved load test metrics"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test metrics"
// @Router /api/v1/load/tests/result/metrics [get]
func (s *AntServer) getLoadTestMetrics(c echo.Context) error {
	var req GetLoadTestResultReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	if strings.TrimSpace(req.LoadTestKey) == "" {
		return errorResponseJson(http.StatusBadRequest, "pass correct load test key")
	}

	arg := load.GetLoadTestResultParam{
		LoadTestKey: req.LoadTestKey,
		Format:      constant.Normal,
	}

	result, err := s.services.loadService.GetLoadTestMetrics(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test metrics")
	}

	return successResponseJson(c, "Successfully retrieved load test metrics", result)
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
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
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
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve all load test execution information")
	}

	return successResponseJson(c, "Successfully retrieved load test execution information", result)
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
		return errorResponseJson(http.StatusBadRequest, "Load test key must be set.")
	}

	arg := load.GetLoadTestExecutionInfoParam{
		LoadTestKey: loadTestKey,
	}

	result, err := s.services.loadService.GetLoadTestExecutionInfo(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test execution state information")
	}

	return successResponseJson(c, "Successfully retrieved load test execution state information", result)
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
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
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
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve all load test execution state information")
	}

	return successResponseJson(c, "Successfully retrieved load test execution state information", result)
}

// getLoadTestExecutionState handler function that retrieves a load test execution state last executed with given ids.
// @Id GetLastLoadTestExecutionState
// @Summary Get Last Load Test Execution State
// @Description Retrieve a last load test execution state by given ids.
// @Tags [Load Test State Management]
// @Accept json
// @Produce json
// @Param nsId query string true "nsId"
// @Param mciId query string true "mciId"
// @Param vmId query string true "Load test key"
// @Success 200 {object} app.AntResponse[load.LoadTestExecutionStateResult] "Successfully retrieved load test execution state information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test execution state information"
// @Router /api/v1/load/tests/state/last [get]
func (s *AntServer) getLastLoadTestExecutionState(c echo.Context) error {
	var req GetLastLoadTestExecutionStateReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	if strings.TrimSpace(req.NsId) == "" || strings.TrimSpace(req.MciId) == "" || strings.TrimSpace(req.VmId) == "" {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	arg := load.GetLoadTestExecutionStateParam{
		NsId:  req.NsId,
		MciId: req.MciId,
		VmId:  req.VmId,
	}

	result, err := s.services.loadService.GetLoadTestExecutionState(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test execution state information")
	}

	return successResponseJson(c, "Successfully retrieved load test execution state information", result)

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
		return errorResponseJson(http.StatusBadRequest, "Load test key must be set.")
	}

	arg := load.GetLoadTestExecutionStateParam{
		LoadTestKey: loadTestKey,
	}

	result, err := s.services.loadService.GetLoadTestExecutionState(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test execution state information")
	}

	return successResponseJson(c, "Successfully retrieved load test execution state information", result)

}

// installMonitoringAgent handler function that handles a monitoring request request to collect metric.
// @Id				InstallMonitoringAgent
// @Summary Install Metrics Monitoring Agent
// @Description Install a monitoring agent on specific mci.
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
		return errorResponseJson(http.StatusBadRequest, "monitoring agent installation info is not correct.")
	}

	arg := load.MonitoringAgentInstallationParams{
		NsId:  req.NsId,
		MciId: req.MciId,
		VmIds: req.VmIds,
	}

	result, err := s.services.loadService.InstallMonitoringAgent(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
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
// @Param mciId query string false "MCI ID" default:""
// @Param vmId query string false "VM ID" default:""
// @Param size query integer false "Number of results per page" default:"10"
// @Param page query integer false "Page number for pagination" default:"1"
// @Success 200 {object} app.AntResponse[load.GetAllMonitoringAgentInfoResult] "Successfully retrieved monitoring agent information"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/monitoring/agent [get]
func (s *AntServer) getAllMonitoringAgentInfos(c echo.Context) error {
	var req GetAllMonitoringAgentInfosReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}
	if req.Size < 1 || req.Size > 10 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}

	arg := load.GetAllMonitoringAgentInfosParam{
		Page:  req.Page,
		Size:  req.Size,
		NsId:  req.NsId,
		MciId: req.MciId,
		VmId:  req.VmId,
	}

	result, err := s.services.loadService.GetAllMonitoringAgentInfos(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve monitoring agent information")
	}

	return successResponseJson(c, "Successfully retrieved monitoring agent information", result)
}

// uninstallMonitoringAgent handler function that initiates the uninstallation of monitoring agents.
// @Id             UninstallMonitoringAgent
// @Summary        Uninstall Monitoring Agents
// @Description    Uninstall monitoring agents from specified VMs or all VMs in an mci.
// @Tags           [Monitoring Agent Management]
// @Accept         json
// @Produce        json
// @Param body body app.MonitoringAgentInstallationReq true "Monitoring Agent Uninstallation Request"
// @Success 200 {object} app.AntResponse[int64] "Number of affected results"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Internal Server Error"
// @Router /api/v1/load/monitoring/agent/uninstall [post]
func (s *AntServer) uninstallMonitoringAgent(c echo.Context) error {
	var req MonitoringAgentInstallationReq

	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "monitoring agent uninstallation info is not correct.")
	}

	arg := load.MonitoringAgentInstallationParams{
		NsId:  req.NsId,
		MciId: req.MciId,
		VmIds: req.VmIds,
	}

	affectedResults, err := s.services.loadService.UninstallMonitoringAgent(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, err.Error())
	}

	return successResponseJson(
		c,
		"monitoring agent is successfully uninstalled",
		affectedResults,
	)
}

// getLastLoadTestResult handler function that retrieves a specific load test result.
// @Id GetLastLoadTestResult
// @Summary Get last load test result by ns, mci, vm
// @Description Retrieve last load test result based on provided parameters.
// @Tags [Load Test Result]
// @Accept json
// @Produce json
// @Param nsId query string true "ns id"
// @Param mciId query string true "mci id"
// @Param vmId query string true "vm id"
// @Param format query string false "Result format (normal or aggregate)"
// @Success 200 {object} app.JsonResult{[normal]=app.AntResponse[[]load.ResultSummary],[aggregate]=app.AntResponse[[]load.LoadTestStatistics]} "Successfully retrieved load test metrics"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test result"
// @Router /api/v1/load/tests/result/last [get]
func (s *AntServer) getLastLoadTestResult(c echo.Context) error {
	var req GetLastLoadTestResultReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	if strings.TrimSpace(req.NsId) == "" || strings.TrimSpace(req.MciId) == "" || strings.TrimSpace(req.VmId) == "" {
		return errorResponseJson(http.StatusBadRequest, "pass correct nsId / mciId / vmId")
	}

	if req.Format == "" {
		req.Format = constant.Normal
	} else if req.Format != constant.Normal && req.Format != constant.Aggregate {
		req.Format = constant.Normal
	}

	arg := load.GetLastLoadTestResultParam{
		NsId:   req.NsId,
		MciId:  req.MciId,
		VmId:   req.VmId,
		Format: req.Format,
	}

	result, err := s.services.loadService.GetLastLoadTestResult(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test result")
	}

	return successResponseJson(c, "Successfully retrieved load test result", result)
}

// getLastLoadTestMetrics handler function that retrieves metrics for a specific load test.
// @Id GetLastLoadTestMetrics
// @Summary Get last load test metrics by ns, mci, vm
// @Description Retrieve last load test metrics based on provided parameters.
// @Tags [Load Test Result]
// @Accept json
// @Produce json
// @Param nsId query string true "ns id"
// @Param mciId query string true "mci id"
// @Param vmId query string true "vm id"
// @Param format query string false "Result format (normal for the moment)"
// @success 200 {object} app.AntResponse[[]load.MetricsSummary] "Successfully retrieved load test metrics"
// @Failure 400 {object} app.AntResponse[string] "Invalid request parameters"
// @Failure 500 {object} app.AntResponse[string] "Failed to retrieve load test metrics"
// @Router /api/v1/load/tests/result/metrics/last [get]
func (s *AntServer) getLastLoadTestMetrics(c echo.Context) error {
	var req GetLastLoadTestResultReq
	if err := c.Bind(&req); err != nil {
		return errorResponseJson(http.StatusBadRequest, "Invalid request parameters")
	}

	if strings.TrimSpace(req.NsId) == "" || strings.TrimSpace(req.MciId) == "" || strings.TrimSpace(req.VmId) == "" {
		return errorResponseJson(http.StatusBadRequest, "pass correct nsId / mciId / vmId")
	}

	arg := load.GetLastLoadTestResultParam{
		NsId:   req.NsId,
		MciId:  req.MciId,
		VmId:   req.VmId,
		Format: constant.Normal,
	}

	result, err := s.services.loadService.GetLastLoadTestMetrics(arg)

	if err != nil {
		return errorResponseJson(http.StatusInternalServerError, "Failed to retrieve load test metrics")
	}

	return successResponseJson(c, "Successfully retrieved load test metrics", result)
}
