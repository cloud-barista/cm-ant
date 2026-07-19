package constant

import "strings"

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
	OnProcessing ExecutionStatus = "on_processing"
	OnFetching   ExecutionStatus = "on_fetching"
	Successed    ExecutionStatus = "successed"
	TestFailed   ExecutionStatus = "test_failed"
)

// ExecutionStep identifies a stage of a load test run (FR-MA2-PERF-007-08).
type ExecutionStep string

// The phases a run goes through. These are what the console shows as the top row.
const (
	StepPrecheck         ExecutionStep = "precheck"
	StepGeneratorInstall ExecutionStep = "generator_install"
	StepAgentInstall     ExecutionStep = "agent_install"
	StepJmxPrepare       ExecutionStep = "jmx_prepare"
	StepJmeterRun        ExecutionStep = "jmeter_run"
	StepResultFetch      ExecutionStep = "result_fetch"
)

// Sub-steps. A phase on its own cannot say where the time went: a run that spent 27 minutes
// unable to reach the metric agent looked exactly like one that was busy generating load,
// because both are "jmeter_run". Each sub-step below is something that can be waited on, so
// it carries its own start and finish time (BAR-1553).
const (
	// precheck — answered in seconds, before anything is provisioned
	SubTargetExists     ExecutionStep = "precheck.target_exists"
	SubTargetRunning    ExecutionStep = "precheck.target_running"
	SubTargetReachable  ExecutionStep = "precheck.target_reachable"
	SubMetricPortOpen   ExecutionStep = "precheck.metric_port_open"
	SubRemoteCommand    ExecutionStep = "precheck.remote_command"
	SubGeneratorPrecond ExecutionStep = "precheck.generator_precond"

	// generator
	SubGeneratorLookup    ExecutionStep = "generator_install.lookup"
	SubGeneratorAlive     ExecutionStep = "generator_install.verify_alive"
	SubGeneratorProvision ExecutionStep = "generator_install.provision"
	SubGeneratorInstall   ExecutionStep = "generator_install.install"
	SubGeneratorVerify    ExecutionStep = "generator_install.verify_install"

	// target metric agent — installed, started and answering are three different things
	SubAgentInstall ExecutionStep = "agent_install.install"
	SubAgentProcess ExecutionStep = "agent_install.process_up"
	SubAgentPort    ExecutionStep = "agent_install.port_reachable"

	// plan
	SubPlanGenerate ExecutionStep = "jmx_prepare.generate"
	SubPlanTransfer ExecutionStep = "jmx_prepare.transfer"

	// load
	SubLoadStart  ExecutionStep = "jmeter_run.start"
	SubLoadRampUp ExecutionStep = "jmeter_run.ramp_up"
	SubLoadHold   ExecutionStep = "jmeter_run.hold"
	SubLoadExit   ExecutionStep = "jmeter_run.exit"

	// collect — per file, because one missing metric file used to fail the whole collection
	SubFileResult  ExecutionStep = "result_fetch.file_result"
	SubFileCpu     ExecutionStep = "result_fetch.file_cpu"
	SubFileMemory  ExecutionStep = "result_fetch.file_memory"
	SubFileDisk    ExecutionStep = "result_fetch.file_disk"
	SubFileNetwork ExecutionStep = "result_fetch.file_network"
	SubPersist     ExecutionStep = "result_fetch.persist"
)

// Parent returns the phase a sub-step belongs to, or "" for a phase itself.
func (s ExecutionStep) Parent() ExecutionStep {
	if i := strings.IndexByte(string(s), '.'); i > 0 {
		return ExecutionStep(s[:i])
	}
	return ""
}

// StepStatus is the per-step lifecycle status (FR-MA2-PERF-007-08).
type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepRunning StepStatus = "running"
	StepOk      StepStatus = "ok"
	StepFailed  StepStatus = "failed"
	StepSkipped StepStatus = "skipped"
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

type ResourceType string

const (
	VM       ResourceType = "VM"
	VNet     ResourceType = "VNet"
	DataDisk ResourceType = "DataDisk"
	Etc      ResourceType = "Etc"

	// NLB           ResourceType = "NLB"
	// Namespace     ResourceType = "Namespace"
	// VMImage       ResourceType = "VMImage"
	// SecurityGroup ResourceType = "SecurityGroup"
	// KeyPair       ResourceType = "KeyPair"
	// Database      ResourceType = "Database"
	// Application   ResourceType = "Application"
	// WebServe      ResourceType = "Web Serve"
)

type AwsService string

const (
	// arn:aws:ec2:<regionId>:<ownerId>:network-interface/eni-xxxxxx, arn:aws:ec2:<regionId>:<ownerId>:elastic-ip/eipalloc-xxxxx
	AwsVpc          AwsService = "Amazon Virtual Private Cloud"
	AwsEC2          AwsService = "Amazon Elastic Compute Cloud - Compute" // i-xxxxx
	AwsEC2Other     AwsService = "EC2 - Other"                            // i-xxxxxx, vol-xxxxxx
	AwsCostExplorer AwsService = "AWS Cost Explorer"                      // NoResourceId
	AwsTax          AwsService = "Tax"                                    // NoResourceId

	// AwsGlue            AwsService = "AWS Glue"
	// AwsKMS             AwsService = "AWS Key Management Service"
	// AwsLambda          AwsService = "AWS Lambda"
	// AwsMigrationHub    AwsService = "AWS Migration Hub Refactor Spaces"
	// AwsSecretManager   AwsService = "AWS Secrets Manager"
	// AwsStepFunction    AwsService = "AWS Step Functions"
	// AwsSystemManager   AwsService = "AWS Systems Manager"
	// AwsApiGateway      AwsService = "Amazon API Gateway"
	// AwsECR             AwsService = "Amazon EC2 Container Registry (ECR)"
	// AwsEKS             AwsService = "Amazon Elastic Container Service for Kubernetes"
	// AwsELB             AwsService = "Amazon Elastic Load Balancing"
	// AwsLightsail       AwsService = "Amazon Lightsail"
	// AwsLocationService AwsService = "Amazon Location Service"
	// AwsRegistrar       AwsService = "Amazon Registrar"
	// AwsRDS             AwsService = "Amazon Relational Database Service"
	// AwsRoute53         AwsService = "Amazon Route 53"
	// AwsSNS             AwsService = "Amazon Simple Notification Service"
	// AwsSQS             AwsService = "Amazon Simple Queue Service"
	// AwsS3              AwsService = "Amazon Simple Storage Service"
	// AwsCloudWatch      AwsService = "AmazonCloudWatch"
)

type CostAggregationType string

const (
	Daily   CostAggregationType = "daily"
	Weekly  CostAggregationType = "weekly"
	Monthly CostAggregationType = "monthly"
)

type OrderType string

const (
	Asc  OrderType = "asc"
	Desc OrderType = "desc"
)

type IconCode string

const (
	Ok      IconCode = "IC0001"
	Fail    IconCode = "IC0002"
	Pending IconCode = "IC0003"
)
