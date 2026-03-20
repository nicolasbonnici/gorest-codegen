package codegen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicolasbonnici/gorest/config"
)

func generateRouteWithAuth(method, path, handler, resource, httpMethod string, authCfg *AuthConfig) string {
	requiresAuth := authCfg != nil && authCfg.RequiresAuth(resource, httpMethod)

	if requiresAuth {
		return fmt.Sprintf(`
	if authMiddleware != nil {
		router.%s("/%s", authMiddleware, %s)
	} else {
		router.%s("/%s", %s)
	}
`, method, path, handler, method, path, handler)
	}

	return fmt.Sprintf(`	router.%s("/%s", %s)
`, method, path, handler)
}

func generateResourceForStruct(apiDir string, structName string, authCfg *AuthConfig) {
	resourceFile := filepath.Join(apiDir, strings.ToLower(structName)+".go")

	cfg, _ := LoadConfig()
	projectRoot, _ := findProjectRoot()
	modelsDir := cfg.Codegen.Output.Models
	if !filepath.IsAbs(modelsDir) {
		modelsDir = filepath.Join(projectRoot, modelsDir)
	}
	modelPath := filepath.Join(modelsDir, strings.ToLower(structName)+".go")
	fields := extractStructFields(modelPath, structName)

	code := generateProcessorResourceFromModel(structName, fields, authCfg)
	if err := os.WriteFile(resourceFile, []byte(code), 0644); err != nil {
		log.Fatalf("failed to write resource for %s: %v", structName, err)
	}
	log.Printf("🧩 Generated processor-based API resource for model: %s → %s", structName, resourceFile)
}

func needsAuthForAnyEndpoint(authCfg *AuthConfig, resourceName string) bool {
	if authCfg == nil {
		return false
	}
	return authCfg.RequiresAuth(resourceName, "GET") ||
		authCfg.RequiresAuth(resourceName, "POST") ||
		authCfg.RequiresAuth(resourceName, "PUT") ||
		authCfg.RequiresAuth(resourceName, "DELETE")
}

func hasFieldNamed(fields []StructField, fieldName string) bool {
	for _, field := range fields {
		if field.Name == fieldName {
			return true
		}
	}
	return false
}

func generateFieldMapping(fields []StructField) string {
	var fieldMappingPairs []string
	for _, field := range fields {
		if field.DBTag != "" && field.JSONTag != "" {
			if field.DBTag == "password" || field.DTOTag == "write" {
				continue
			}
			fieldMappingPairs = append(fieldMappingPairs, fmt.Sprintf(`"%s": "%s"`, field.JSONTag, field.DBTag))
		}
	}
	return strings.Join(fieldMappingPairs, ", ")
}

func checkHooksExist(lowerStructName string) bool {
	projectRoot, _ := findProjectRoot()
	hookFilePath := filepath.Join(projectRoot, "hooks", lowerStructName+".go")
	if _, err := os.Stat(hookFilePath); err == nil {
		return true
	}
	return false
}

func buildImportPaths(moduleName string, cfg *config.Config) (modelsImport, dtosImport string) {
	modelsImport = moduleName
	dtosImport = moduleName

	if cfg.Codegen.Output.Models != "models" && cfg.Codegen.Output.Models != "" {
		modelsPath := strings.TrimPrefix(cfg.Codegen.Output.Models, "./")
		modelsImport = moduleName + "/" + strings.ReplaceAll(modelsPath, string(filepath.Separator), "/")
	}
	if cfg.Codegen.Output.DTOs != "dtos" && cfg.Codegen.Output.DTOs != "" {
		dtosPath := strings.TrimPrefix(cfg.Codegen.Output.DTOs, "./")
		dtosImport = moduleName + "/" + strings.ReplaceAll(dtosPath, string(filepath.Separator), "/")
	}

	modelsImport = strings.TrimSuffix(modelsImport, "/models") + "/models"
	dtosImport = strings.TrimSuffix(dtosImport, "/dtos") + "/dtos"
	return
}

func shouldIncludeInDTO(field StructField) bool {
	return field.DTOTag != "-" && field.DTOTag != "write"
}

func shouldIncludeInCreateUpdate(field StructField) bool {
	dbTag := strings.ToLower(field.DBTag)
	if dbTag == FieldID || dbTag == FieldCreatedAt || dbTag == FieldUpdatedAt {
		return false
	}
	return field.DTOTag != "-" && field.DTOTag != "read"
}

func generateModelToDTOFields(fields []StructField) string {
	var builder strings.Builder
	for _, field := range fields {
		if shouldIncludeInDTO(field) {
			builder.WriteString(fmt.Sprintf("\t\t%s: m.%s,\n", field.Name, field.Name))
		}
	}
	return builder.String()
}

func generateCreateUpdateFields(fields []StructField) string {
	var builder strings.Builder
	for _, field := range fields {
		if shouldIncludeInCreateUpdate(field) {
			builder.WriteString(fmt.Sprintf("\t\t%s: dto.%s,\n", field.Name, field.Name))
		}
	}
	return builder.String()
}

func generateConversionFunctions(structName string, fields []StructField) string {
	modelToDTOFields := generateModelToDTOFields(fields)
	createDTOToModelFields := generateCreateUpdateFields(fields)
	updateDTOToModelFields := generateCreateUpdateFields(fields)

	lowerStructName := strings.ToLower(structName)
	return fmt.Sprintf(`func modelTo%sDTO(m models.%s) dtos.%sDTO {
	return dtos.%sDTO{
%s	}
}

func %sCreateDTOToModel(dto dtos.%sCreateDTO) models.%s {
	return models.%s{
%s	}
}

func %sUpdateDTOToModel(dto dtos.%sUpdateDTO) models.%s {
	return models.%s{
%s	}
}
`, structName, structName, structName, structName, modelToDTOFields,
		lowerStructName, structName, structName, structName, createDTOToModelFields,
		lowerStructName, structName, structName, structName, updateDTOToModelFields)
}

func generateProcessorResourceFromModel(structName string, fields []StructField, authCfg *AuthConfig) string {
	resourceName := strings.ToLower(structName)
	lowerStructName := strings.ToLower(structName)
	pluralResourceName := Pluralize(resourceName)

	listRoute := generateRouteWithAuth("Get", pluralResourceName, "res.List", pluralResourceName, "GET", authCfg)
	getRoute := generateRouteWithAuth("Get", pluralResourceName+"/:id", "res.Get", pluralResourceName, "GET", authCfg)
	postRoute := generateRouteWithAuth("Post", pluralResourceName, "res.Create", pluralResourceName, "POST", authCfg)
	putRoute := generateRouteWithAuth("Put", pluralResourceName+"/:id", "res.Update", pluralResourceName, "PUT", authCfg)
	deleteRoute := generateRouteWithAuth("Delete", pluralResourceName+"/:id", "res.Delete", pluralResourceName, "DELETE", authCfg)

	needsAuthContext := needsAuthForAnyEndpoint(authCfg, pluralResourceName)

	routesSignature := "router fiber.Router, db database.Database, paginationLimit, paginationMaxLimit int, pluginRegistry *plugin.PluginRegistry"

	authMiddlewareSetup := ""
	if needsAuthContext {
		authMiddlewareSetup = `
	var authMiddleware fiber.Handler
	if authPlugin, ok := pluginRegistry.Get("auth"); ok {
		authMiddleware = authPlugin.Handler()
	}
`
	}

	hasUserIdField := hasFieldNamed(fields, "UserId")

	userIdEnricher := ""
	if hasUserIdField {
		userIdEnricher = `
		ContextEnrichers: []processor.ContextEnricher{
			processor.UserIDEnricher("UserId"),
		},`
	}

	conversionFuncs := generateConversionFunctions(structName, fields)
	fieldMappingStr := generateFieldMapping(fields)
	hasHooks := checkHooksExist(lowerStructName)

	moduleName := getModuleName()
	cfg, _ := LoadConfig()

	modelsImport, dtosImport := buildImportPaths(moduleName, cfg)

	importsSection := fmt.Sprintf(`import (
	"%s"
	"%s"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/plugin"
	"github.com/nicolasbonnici/gorest/processor"`, dtosImport, modelsImport)

	if hasHooks {
		hooksImport := moduleName + "/hooks"
		importsSection += fmt.Sprintf(`
	"%s"`, hooksImport)
	}
	importsSection += `
)`

	crudInit := fmt.Sprintf("crud.New[models.%s](db)", structName)
	if hasHooks {
		crudInit = fmt.Sprintf("crud.NewWithHooks[models.%s](db, &hooks.%sHooks{})", structName, structName)
	}

	fieldMappingPairs := ""
	if fieldMappingStr != "" {
		fieldMappingPairs = fmt.Sprintf(`
		FieldMap: map[string]string{%s},`, fieldMappingStr)
	}

	return fmt.Sprintf(`// Code generated by GoREST. DO NOT EDIT.

package resources

%s

type %sResource struct {
	processor processor.Processor[models.%s, dtos.%sCreateDTO, dtos.%sUpdateDTO, dtos.%sDTO]
}

func New%sResource(db database.Database, paginationLimit, paginationMaxLimit int) *%sResource {
	%sCRUD := %s

	%sProcessor := processor.New(processor.ProcessorConfig[
		models.%s,
		dtos.%sCreateDTO,
		dtos.%sUpdateDTO,
		dtos.%sDTO,
	]{
		DB:                 db,
		CRUD:               %sCRUD,
		PaginationLimit:    paginationLimit,
		PaginationMaxLimit: paginationMaxLimit,%s
		Converter: &processor.FuncConverter[models.%s, dtos.%sCreateDTO, dtos.%sUpdateDTO, dtos.%sDTO]{
			CreateToModel: %sCreateDTOToModel,
			UpdateToModel: %sUpdateDTOToModel,
			ModelToDTO:    modelTo%sDTO,
		},%s
	})

	return &%sResource{
		processor: %sProcessor,
	}
}

func Register%sRoutes(%s) {
	res := New%sResource(db, paginationLimit, paginationMaxLimit)
%s%s%s%s%s%s
}

%s

// List %s
// @Summary List %s
// @Tags %s
// @Produce json,application/ld+json
// @Success 200 {object} pagination.HydraCollection
// @Router /%s [get]
func (r *%sResource) List(c *fiber.Ctx) error {
	return r.processor.GetAll(c)
}

// Get %s by ID
// @Summary Get %s
// @Tags %s
// @Produce json,application/ld+json
// @Param id path string true "ID"
// @Success 200 {object} dtos.%sDTO
// @Router /%s/{id} [get]
func (r *%sResource) Get(c *fiber.Ctx) error {
	return r.processor.GetByID(c)
}

// Create %s
// @Summary Create %s
// @Tags %s
// @Accept json
// @Produce json,application/ld+json
// @Param input body dtos.%sCreateDTO true "New %s"
// @Success 201 {object} dtos.%sDTO
// @Router /%s [post]
func (r *%sResource) Create(c *fiber.Ctx) error {
	return r.processor.Create(c)
}

// Update %s
// @Summary Update %s
// @Tags %s
// @Accept json
// @Produce json,application/ld+json
// @Param id path string true "ID"
// @Param input body dtos.%sUpdateDTO true "Updated %s"
// @Success 200 {object} dtos.%sDTO
// @Router /%s/{id} [put]
func (r *%sResource) Update(c *fiber.Ctx) error {
	return r.processor.Update(c)
}

// Delete %s
// @Summary Delete %s
// @Tags %s
// @Param id path string true "ID"
// @Success 204
// @Router /%s/{id} [delete]
func (r *%sResource) Delete(c *fiber.Ctx) error {
	return r.processor.Delete(c)
}
`,
		importsSection,
		structName, structName, structName, structName, structName,
		structName, structName,
		lowerStructName, crudInit,
		lowerStructName,
		structName, structName, structName, structName,
		lowerStructName,
		fieldMappingPairs,
		structName, structName, structName, structName,
		lowerStructName, lowerStructName, structName,
		userIdEnricher,
		structName, lowerStructName,
		structName, routesSignature, structName,
		authMiddlewareSetup,
		listRoute, getRoute, postRoute, putRoute, deleteRoute,
		conversionFuncs,
		structName, structName, structName, pluralResourceName, structName,
		structName, structName, structName, structName, pluralResourceName, structName,
		structName, structName, structName, structName, structName, structName, pluralResourceName, structName,
		structName, structName, structName, structName, structName, structName, pluralResourceName, structName,
		structName, structName, structName, pluralResourceName, structName)
}
