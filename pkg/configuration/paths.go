package configuration

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _  = runtime.Caller(0)
	basePath    = filepath.Dir(b)
	rootPath    = basePath[0 : len(basePath)-len("/pkg/configuration")]
	homePath, _ = homedir.Dir()
)

func RootPath() string {
	return rootPath
}

func JoinRootPathWith(path string) string {
	if path == "" {
		return rootPath
	}

	if path[:1] != "/" {
		path = "/" + path
	}

	return fmt.Sprintf("%s%s", rootPath, path)
}
