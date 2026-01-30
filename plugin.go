package codegen

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/config"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/plugin"
)

// CodegenPlugin implements the plugin.Plugin interface
type CodegenPlugin struct {
	db        database.Database
	appConfig *config.Config
}

// NewPlugin creates a new instance of the codegen plugin
func NewPlugin() plugin.Plugin {
	return &CodegenPlugin{}
}

// Name returns the plugin name
func (p *CodegenPlugin) Name() string {
	return "codegen"
}

// Initialize initializes the plugin with configuration
func (p *CodegenPlugin) Initialize(cfg map[string]any) error {
	// Extract injected dependencies
	if db, ok := cfg["database"].(database.Database); ok {
		p.db = db
	}
	if appConfig, ok := cfg["config"].(*config.Config); ok {
		p.appConfig = appConfig
	}
	return nil
}

// Handler returns a no-op middleware handler (codegen is CLI-only)
func (p *CodegenPlugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// Commands returns the list of available commands
func (p *CodegenPlugin) Commands() []plugin.Command {
	return []plugin.Command{
		&ModelsCommand{plugin: p},
		&ResourcesCommand{plugin: p},
		&OpenAPICommand{plugin: p},
		&AllCommand{plugin: p},
	}
}
