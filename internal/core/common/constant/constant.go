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
	OnProcessing ExecutionStatus = "on_processing"
	OnFetching   ExecutionStatus = "on_fetching"
	Successed    ExecutionStatus = "successed"
	TestFailed   ExecutionStatus = "test_failed"
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
	Ok IconCode = "IC0001"
	Fail IconCode = "IC0002"
	Pending IconCode = "IC0003"
)
