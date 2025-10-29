package tracer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var goModRootPath string = func() string {
	dir, _ := os.Getwd()
	for dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}()

func callbackInfo() (string, string) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	funcName := "unknown"
	if f := runtime.FuncForPC(pc); f != nil {
		funcName = f.Name()
		funcName = shortenFuncName(funcName)
	}

	shortPath := shortenModulePath(file)
	modulePath := fmt.Sprintf("%s:%d", shortPath, line)
	actionName := funcName

	return modulePath, actionName
}

func shortenModulePath(fullPath string) string {
	if goModRootPath == "" {
		return fullPath
	}

	rel, err := filepath.Rel(goModRootPath, fullPath)
	if err != nil {
		return fullPath
	}
	return rel
}

func shortenFuncName(fullFuncName string) string {
	if idx := strings.LastIndex(fullFuncName, "/"); idx != -1 {
		fullFuncName = fullFuncName[idx+1:]
	}
	fullFuncName = strings.TrimPrefix(fullFuncName, "(*")
	fullFuncName = strings.TrimSuffix(fullFuncName, ")")
	return fullFuncName
}
