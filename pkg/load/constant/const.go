package constant

type OsType string

const (
	UBUNTU OsType = "ubuntu"
	CENTOS OsType = "centos"
	DEBIAN OsType = "debian"
	ARCH   OsType = "arch"
)

type AccessType string

const (
	Local  AccessType = "local"
	Remote AccessType = "remote"
)

type RemoteConnectionType string

const (
	Password   RemoteConnectionType = "password"
	PrivateKey RemoteConnectionType = "privateKey"
	BuiltIn    RemoteConnectionType = "builtIn"
)
