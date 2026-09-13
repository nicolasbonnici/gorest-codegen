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

// generatedFilePerm is the mode for every emitted file. Generated Go sources
// carry no secrets and have to stay readable by whatever user later builds the
// tree, so the group/other read bits are deliberate rather than an oversight.
const generatedFilePerm os.FileMode = 0644

// writeGeneratedFile centralises the emit so the permission choice above is
// stated once instead of at each call site.
func writeGeneratedFile(path, code string) error {
	return os.WriteFile(path, []byte(code), generatedFilePerm) // #nosec G306
}
