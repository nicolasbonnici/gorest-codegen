package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func FindProjectRoot() (string, error) {
	return findProjectRoot()
}

var (
	projectRootMu    sync.RWMutex
	projectRootCache = map[string]string{}
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projectRootMu.RLock()
	cached, ok := projectRootCache[dir]
	projectRootMu.RUnlock()
	if ok {
		return cached, nil
	}

	start := dir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRootMu.Lock()
			projectRootCache[start] = dir
			projectRootMu.Unlock()
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
