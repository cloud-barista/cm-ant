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

type ResourceType string

const (
	VM       ResourceType = "VM"
	VNet     ResourceType = "VNet"
	DataDisk ResourceType = "DataDisk"
	Etc      ResourceType = "Etc"

	NLB           ResourceType = "NLB"
	Namespace     ResourceType = "Namespace"
	VMImage       ResourceType = "VMImage"
	SecurityGroup ResourceType = "SecurityGroup"
	KeyPair       ResourceType = "KeyPair"
	Database      ResourceType = "Database"
	Application   ResourceType = "Application"
	WebServe      ResourceType = "Web Serve"
)

type AwsService string

const (
	// arn:aws:ec2:ap-northeast-1:050864702683:network-interface/eni-04c34238a6db01e9b
	AwsVpc          AwsService = "Amazon Virtual Private Cloud"           // arn:aws:ec2:<regionId>:<ownerId>:network-interface/eni-07dd0cea3916a5c99, arn:aws:ec2:<regionId>:<ownerId>:elastic-ip/eipalloc-0429227700117b866
	AwsEC2          AwsService = "Amazon Elastic Compute Cloud - Compute" // i-005d85400a201b7f3
	AwsEC2Other     AwsService = "EC2 - Other"                            // i-005d85400a201b7f3, vol-0033c8dcc5aeb70c9
	AwsCostExplorer AwsService = "AWS Cost Explorer"                      // NoResourceId
	AwsTax          AwsService = "Tax"                                    // NoResourceId

	AwsGlue            AwsService = "AWS Glue"
	AwsKMS             AwsService = "AWS Key Management Service"
	AwsLambda          AwsService = "AWS Lambda"
	AwsMigrationHub    AwsService = "AWS Migration Hub Refactor Spaces"
	AwsSecretManager   AwsService = "AWS Secrets Manager"
	AwsStepFunction    AwsService = "AWS Step Functions"
	AwsSystemManager   AwsService = "AWS Systems Manager"
	AwsApiGateway      AwsService = "Amazon API Gateway"
	AwsECR             AwsService = "Amazon EC2 Container Registry (ECR)"
	AwsEKS             AwsService = "Amazon Elastic Container Service for Kubernetes"
	AwsELB             AwsService = "Amazon Elastic Load Balancing"
	AwsLightsail       AwsService = "Amazon Lightsail"
	AwsLocationService AwsService = "Amazon Location Service"
	AwsRegistrar       AwsService = "Amazon Registrar"
	AwsRDS             AwsService = "Amazon Relational Database Service"
	AwsRoute53         AwsService = "Amazon Route 53"
	AwsSNS             AwsService = "Amazon Simple Notification Service"
	AwsSQS             AwsService = "Amazon Simple Queue Service"
	AwsS3              AwsService = "Amazon Simple Storage Service"
	AwsCloudWatch      AwsService = "AmazonCloudWatch"
)
