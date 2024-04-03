package constant

type OsType string

const (
	UBUNTU OsType = "ubuntu"
	CENTOS OsType = "centos"
	DEBIAN OsType = "debian"
	ARCH   OsType = "arch"
)

type InstallLocation string

const (
	Local  InstallLocation = "local"
	Remote InstallLocation = "remote"
)

type RemoteConnectionType string

const (
	Password   RemoteConnectionType = "password"
	PrivateKey RemoteConnectionType = "privateKey"
	BuiltIn    RemoteConnectionType = "builtIn"
)

type ExecutionStatus string

const (
	Progress ExecutionStatus = "progress"
	Success  ExecutionStatus = "success"
	Failed   ExecutionStatus = "failed"
)
