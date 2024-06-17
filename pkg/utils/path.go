package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	rootPath = getRootPath()
)

func getRootPath() string {
	envRootPath := os.Getenv("ANT_ROOT_PATH")
	if envRootPath != "" {
		return envRootPath
	}

	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	return basePath[0 : len(basePath)-len("/pkg/utils")]
}

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
