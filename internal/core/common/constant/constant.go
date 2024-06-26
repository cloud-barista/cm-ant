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
	Processing ExecutionStatus = "processing"
	Fetching   ExecutionStatus = "fetching"
	Success    ExecutionStatus = "success"
	Failed     ExecutionStatus = "failed"
)
