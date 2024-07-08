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
	OnPreparing ExecutionStatus = "on_preparing"
	OnRunning   ExecutionStatus = "on_running"
	OnFetching  ExecutionStatus = "on_fetching"
	Successed   ExecutionStatus = "successed"
	Failed      ExecutionStatus = "failed"

	Processing ExecutionStatus = "processing"
	Fetching   ExecutionStatus = "fetching"
	Success    ExecutionStatus = "success"
)
