package codegen

import (
	"fmt"

	"github.com/nicolasbonnici/gorest-codegen/codegen"
	"github.com/nicolasbonnici/gorest/config"
	"github.com/nicolasbonnici/gorest/plugin"
)

// ModelsCommand generates model structs from database schema
type ModelsCommand struct {
	plugin *CodegenPlugin
}

func (c *ModelsCommand) Name() string {
	return "models"
}

func (c *ModelsCommand) Description() string {
	return "Generate model structs from database schema"
}

func (c *ModelsCommand) Run(ctx *plugin.CommandContext) *plugin.CommandResult {
	ctx.ProgressCallback("Loading database schema...")
	tables := codegen.LoadSchema(c.plugin.db)

	ctx.ProgressCallback("Generating model structs...")
	codegen.GenerateStructs(tables)

	ctx.ProgressCallback("Models generated successfully")

	return &plugin.CommandResult{
		Success:      true,
		FilesCreated: []string{"generated/models/*.go"},
		Message:      "Model generation completed successfully",
	}
}

// ResourcesCommand generates REST API resources and DTOs from models
type ResourcesCommand struct {
	plugin *CodegenPlugin
}

func (c *ResourcesCommand) Name() string {
	return "resources"
}

func (c *ResourcesCommand) Description() string {
	return "Generate REST API resources and DTOs from models"
}

func (c *ResourcesCommand) Run(ctx *plugin.CommandContext) *plugin.CommandResult {
	ctx.ProgressCallback("Loading configuration...")

	// Try to get config from plugin first, fallback to context
	var cfg *config.Config
	if c.plugin.appConfig != nil {
		cfg = c.plugin.appConfig
	} else if ctx.Config != nil {
		if contextCfg, ok := ctx.Config.(*config.Config); ok {
			cfg = contextCfg
		}
	}

	if cfg == nil {
		// Load from current directory as fallback
		loadedCfg, err := config.Load(".")
		if err != nil {
			return &plugin.CommandResult{
				Success: false,
				Error:   fmt.Errorf("failed to load config: %w", err),
			}
		}
		cfg = loadedCfg
	}

	ctx.ProgressCallback("Building authentication configuration...")
	authCfg := codegen.GetAuthConfigFromConfig(cfg)

	ctx.ProgressCallback("Generating API resources...")
	codegen.GenerateAPI(authCfg)

	ctx.ProgressCallback("Resources generated successfully")

	return &plugin.CommandResult{
		Success: true,
		FilesCreated: []string{
			"generated/resources/*.go",
			"generated/dtos/*.go",
		},
		Message: "Resource generation completed successfully",
	}
}

// OpenAPICommand generates OpenAPI schema file
type OpenAPICommand struct {
	plugin *CodegenPlugin
}

func (c *OpenAPICommand) Name() string {
	return "openapi"
}

func (c *OpenAPICommand) Description() string {
	return "Generate OpenAPI schema file"
}

func (c *OpenAPICommand) Run(ctx *plugin.CommandContext) *plugin.CommandResult {
	ctx.ProgressCallback("Generating OpenAPI schema...")
	tables := codegen.LoadSchema(c.plugin.db)
	codegen.GenerateOpenAPI(tables)

	ctx.ProgressCallback("OpenAPI schema generated successfully")

	return &plugin.CommandResult{
		Success:      true,
		FilesCreated: []string{"generated/openapi/schema.yaml"},
		Message:      "OpenAPI generation completed successfully",
	}
}

// AllCommand runs all code generation steps
type AllCommand struct {
	plugin *CodegenPlugin
}

func (c *AllCommand) Name() string {
	return "all"
}

func (c *AllCommand) Description() string {
	return "Run all code generation steps (models, resources, openapi)"
}

func (c *AllCommand) Run(ctx *plugin.CommandContext) *plugin.CommandResult {
	ctx.ProgressCallback("Running: models")
	modelsCmd := &ModelsCommand{plugin: c.plugin}
	modelsResult := modelsCmd.Run(ctx)
	if !modelsResult.Success {
		return modelsResult
	}

	ctx.ProgressCallback("Running: resources")
	resourcesCmd := &ResourcesCommand{plugin: c.plugin}
	resourcesResult := resourcesCmd.Run(ctx)
	if !resourcesResult.Success {
		return resourcesResult
	}

	ctx.ProgressCallback("Running: openapi")
	openapiCmd := &OpenAPICommand{plugin: c.plugin}
	openapiResult := openapiCmd.Run(ctx)
	if !openapiResult.Success {
		return openapiResult
	}

	// Combine all created files
	allFiles := make([]string, 0)
	allFiles = append(allFiles, modelsResult.FilesCreated...)
	allFiles = append(allFiles, resourcesResult.FilesCreated...)
	allFiles = append(allFiles, openapiResult.FilesCreated...)

	return &plugin.CommandResult{
		Success:      true,
		FilesCreated: allFiles,
		Message:      "All code generation completed successfully",
	}
}
