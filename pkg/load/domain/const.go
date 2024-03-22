package domain

type OsType string

const (
	UBUNTU OsType = "ubuntu"
	CENTOS OsType = "centos"
	DEBIAN OsType = "debian"
	ARCH   OsType = "arch"
)

type AccessType string

const (
	LOCAL  AccessType = "local"
	REMOTE AccessType = "remote"
)
