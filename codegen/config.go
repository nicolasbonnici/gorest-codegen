package codegen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nicolasbonnici/gorest/config"
)

var (
	configMu    sync.Mutex
	configCache = map[string]*config.Config{}
)

func LoadConfig() (*config.Config, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	configMu.Lock()
	defer configMu.Unlock()
	if cfg, ok := configCache[projectRoot]; ok {
		return cfg, nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}
	configCache[projectRoot] = cfg
	return cfg, nil
}

func GetModelsPath(cfg *config.Config) (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, cfg.Codegen.Output.Models), nil
}

func GetResourcesPath(cfg *config.Config) (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, cfg.Codegen.Output.Resources), nil
}

func GetDTOsPath(cfg *config.Config) (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, cfg.Codegen.Output.DTOs), nil
}

func GetOpenAPIPath(cfg *config.Config) (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, cfg.Codegen.Output.OpenAPI), nil
}

func GetConfigPath(cfg *config.Config) (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, cfg.Codegen.Output.Config), nil
}

func GetRoutesPath(cfg *config.Config) (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, cfg.Codegen.Output.Resources, "routes.go"), nil
}

var (
	moduleNameMu    sync.Mutex
	moduleNameCache = map[string]string{}
)

func getModuleName() string {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "github.com/nicolasbonnici/gorest"
	}

	moduleNameMu.Lock()
	defer moduleNameMu.Unlock()
	if name, ok := moduleNameCache[projectRoot]; ok {
		return name
	}

	name := readModuleName(projectRoot)
	moduleNameCache[projectRoot] = name
	return name
}

func readModuleName(projectRoot string) string {
	goModPath := filepath.Join(projectRoot, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "github.com/nicolasbonnici/gorest"
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	return "github.com/nicolasbonnici/gorest"
}
