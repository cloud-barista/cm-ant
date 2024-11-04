package load

import (
	"context"
	"time"

	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/internal/utils"
)

// LoadService represents a service for managing load operations.
type LoadService struct {
	loadRepo        *LoadRepository
	tumblebugClient *tumblebug.TumblebugClient
}

// NewLoadService creates a new instance of LoadService.
func NewLoadService(loadRepo *LoadRepository, client *tumblebug.TumblebugClient) *LoadService {
	return &LoadService{
		loadRepo:        loadRepo,
		tumblebugClient: client,
	}
}

func (l *LoadService) Readyz() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := l.loadRepo.db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	err = l.tumblebugClient.ReadyzWithContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (l *LoadService) GetAllLoadTestExecutionState(param GetAllLoadTestExecutionStateParam) (GetAllLoadTestExecutionStateResult, error) {
	var res GetAllLoadTestExecutionStateResult
	var states []LoadTestExecutionStateResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetAllLoadExecutionStates called with param: %+v", param)
	result, totalRows, err := l.loadRepo.GetPagingLoadTestExecutionStateTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution state infos: %v", err)
		return res, err
	}

	utils.LogInfof("Fetched %d monitoring agent infos", len(result))

	for _, loadTestExecutionState := range result {
		state := mapLoadTestExecutionStateResult(loadTestExecutionState)
		state.LoadGeneratorInstallInfo = mapLoadGeneratorInstallInfoResult(loadTestExecutionState.LoadGeneratorInstallInfo)
		states = append(states, state)
	}

	res.LoadTestExecutionStates = states
	res.TotalRow = totalRows

	return res, nil
}

func (l *LoadService) GetLoadTestExecutionState(param GetLoadTestExecutionStateParam) (LoadTestExecutionStateResult, error) {
	var res LoadTestExecutionStateResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetLoadTestExecutionState called with param: %+v", param)
	state, err := l.loadRepo.GetLoadTestExecutionStateTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution state infos: %v", err)
		return res, err
	}

	res = mapLoadTestExecutionStateResult(state)
	res.LoadGeneratorInstallInfo = mapLoadGeneratorInstallInfoResult(state.LoadGeneratorInstallInfo)
	return res, nil
}

func (l *LoadService) GetAllLoadTestExecutionInfos(param GetAllLoadTestExecutionInfosParam) (GetAllLoadTestExecutionInfosResult, error) {
	var res GetAllLoadTestExecutionInfosResult
	var rs []LoadTestExecutionInfoResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetAllLoadTestExecutionInfos called with param: %+v", param)
	result, totalRows, err := l.loadRepo.GetPagingLoadTestExecutionHistoryTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution infos: %v", err)
		return res, err
	}

	utils.LogInfof("Fetched %d load test execution infos:", len(result))

	for _, r := range result {
		rs = append(rs, mapLoadTestExecutionInfoResult(r))
	}

	res.TotalRow = totalRows
	res.LoadTestExecutionInfos = rs

	return res, nil
}

func (l *LoadService) GetLoadTestExecutionInfo(param GetLoadTestExecutionInfoParam) (LoadTestExecutionInfoResult, error) {
	var res LoadTestExecutionInfoResult
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.LogInfof("GetLoadTestExecutionInfo called with param: %+v", param)
	executionInfo, err := l.loadRepo.GetLoadTestExecutionInfoTx(ctx, param)

	if err != nil {
		utils.LogErrorf("Error fetching load test execution state infos: %v", err)
		return res, err
	}

	return mapLoadTestExecutionInfoResult(executionInfo), nil
}

func mapLoadTestExecutionHttpInfoResult(h LoadTestExecutionHttpInfo) LoadTestExecutionHttpInfoResult {
	return LoadTestExecutionHttpInfoResult{
		ID:       h.ID,
		Method:   h.Method,
		Protocol: h.Protocol,
		Hostname: h.Hostname,
		Port:     h.Port,
		Path:     h.Path,
		BodyData: h.BodyData,
	}
}

func mapLoadTestExecutionStateResult(state LoadTestExecutionState) LoadTestExecutionStateResult {
	return LoadTestExecutionStateResult{
		ID:                          state.ID,
		LoadTestKey:                 state.LoadTestKey,
		ExecutionStatus:             state.ExecutionStatus,
		StartAt:                     state.StartAt,
		FinishAt:                    state.FinishAt,
		TotalExpectedExcutionSecond: state.TotalExpectedExcutionSecond,
		FailureMessage:              state.FailureMessage,
		CompileDuration:             state.CompileDuration,
		ExecutionDuration:           state.ExecutionDuration,
		CreatedAt:                   state.CreatedAt,
		UpdatedAt:                   state.UpdatedAt,
	}
}

func mapLoadGeneratorServerResult(s LoadGeneratorServer) LoadGeneratorServerResult {
	return LoadGeneratorServerResult{
		ID:              s.ID,
		Csp:             s.Csp,
		Region:          s.Region,
		Zone:            s.Zone,
		PublicIp:        s.PublicIp,
		PrivateIp:       s.PrivateIp,
		PublicDns:       s.PublicDns,
		MachineType:     s.MachineType,
		Status:          s.Status,
		SshPort:         s.SshPort,
		Lat:             s.Lat,
		Lon:             s.Lon,
		Username:        s.Username,
		VmId:            s.VmId,
		StartTime:       s.StartTime,
		AdditionalVmKey: s.AdditionalVmKey,
		Label:           s.Label,
		CreatedAt:       s.CreatedAt,
	}
}

func mapLoadGeneratorInstallInfoResult(install LoadGeneratorInstallInfo) LoadGeneratorInstallInfoResult {
	var servers []LoadGeneratorServerResult
	for _, s := range install.LoadGeneratorServers {
		servers = append(servers, mapLoadGeneratorServerResult(s))
	}

	return LoadGeneratorInstallInfoResult{
		ID:                   install.ID,
		InstallLocation:      install.InstallLocation,
		InstallType:          install.InstallType,
		InstallPath:          install.InstallPath,
		InstallVersion:       install.InstallVersion,
		Status:               install.Status,
		CreatedAt:            install.CreatedAt,
		UpdatedAt:            install.UpdatedAt,
		PublicKeyName:        install.PublicKeyName,
		PrivateKeyName:       install.PrivateKeyName,
		LoadGeneratorServers: servers,
	}
}

func mapLoadTestExecutionInfoResult(executionInfo LoadTestExecutionInfo) LoadTestExecutionInfoResult {
	var httpResults []LoadTestExecutionHttpInfoResult
	for _, h := range executionInfo.LoadTestExecutionHttpInfos {
		httpResults = append(httpResults, mapLoadTestExecutionHttpInfoResult(h))
	}

	executionState := mapLoadTestExecutionStateResult(executionInfo.LoadTestExecutionState)
	installInfo := mapLoadGeneratorInstallInfoResult(executionInfo.LoadGeneratorInstallInfo)

	return LoadTestExecutionInfoResult{
		ID:                         executionInfo.ID,
		LoadTestKey:                executionInfo.LoadTestKey,
		TestName:                   executionInfo.TestName,
		VirtualUsers:               executionInfo.VirtualUsers,
		Duration:                   executionInfo.Duration,
		RampUpTime:                 executionInfo.RampUpTime,
		RampUpSteps:                executionInfo.RampUpSteps,
		AgentHostname:              executionInfo.AgentHostname,
		AgentInstalled:             executionInfo.AgentInstalled,
		CompileDuration:            executionInfo.CompileDuration,
		ExecutionDuration:          executionInfo.ExecutionDuration,
		LoadTestExecutionHttpInfos: httpResults,
		LoadTestExecutionState:     executionState,
		LoadGeneratorInstallInfo:   installInfo,
	}
}
