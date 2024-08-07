package constant

type InstallLocation string

const (
	Local  InstallLocation = "local"
	Remote InstallLocation = "remote"
)

type LoadGeneratorType string

const (
	Jmeter LoadGeneratorType = "jmeter"
)

type ExecutionStatus string

const (
	OnPreparing  ExecutionStatus = "on_preparing"
	OnRunning    ExecutionStatus = "on_running"
	OnFetching   ExecutionStatus = "on_fetching"
	Successed    ExecutionStatus = "successed"
	TestFailed   ExecutionStatus = "test_failed"
	UpdateFailed ExecutionStatus = "update_failed"
	ResultFailed ExecutionStatus = "result_failed"

	Failed ExecutionStatus = "failed"

	Processing ExecutionStatus = "processing"
	Fetching   ExecutionStatus = "fetching"
	Success    ExecutionStatus = "success"
)

type ResultFormat string

const (
	Normal    ResultFormat = "normal"
	Aggregate ResultFormat = "aggregate"
)

type PricePolicy string

const (
	OnDemand PricePolicy = "OnDemand"
)

type PriceUnit string

const (
	PerHour PriceUnit = "PerHour"
	PerYear PriceUnit = "PerYear"
)

type PriceCurrency string

const (
	USD PriceCurrency = "USD"
	KRW PriceCurrency = "KRW"
)

type MemoryUnit string

const (
	GIB MemoryUnit = "GiB"
)
