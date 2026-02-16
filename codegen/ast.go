package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

type StructField struct {
	Name      string
	Type      string
	JSONTag   string
	DBTag     string
	DTOTag    string
	IsPointer bool
}

func parseStructs(path string) []string {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
	if err != nil {
		log.Printf("parse error in %s: %v", path, err)
		return nil
	}

	structs := []string{}
	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, ok := ts.Type.(*ast.StructType); ok {
				structs = append(structs, ts.Name.Name)
			}
		}
	}
	return structs
}

func extractFieldType(fieldType ast.Expr) (typeName string, isPointer bool) {
	switch t := fieldType.(type) {
	case *ast.Ident:
		return t.Name, false
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name, true
		}
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok {
				return pkg.Name + "." + sel.Sel.Name, true
			}
		}
		return "", true
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name + "." + t.Sel.Name, false
		}
	case *ast.InterfaceType:
		return "interface{}", false
	case *ast.MapType:
		return "map[string]interface{}", false
	}
	return "", false
}

func extractFieldTags(field *ast.Field) (jsonTag, dbTag, dtoTag string) {
	if field.Tag == nil {
		return "", "", ""
	}
	tag := field.Tag.Value
	jsonTag = extractTag(tag, "json")
	jsonTag = strings.Split(jsonTag, ",")[0]
	dbTag = extractTag(tag, "db")
	dtoTag = extractTag(tag, "dto")
	return
}

func processStructField(field *ast.Field) *StructField {
	if len(field.Names) == 0 {
		return nil
	}

	fieldName := field.Names[0].Name
	fieldType, isPointer := extractFieldType(field.Type)
	jsonTag, dbTag, dtoTag := extractFieldTags(field)

	return &StructField{
		Name:      fieldName,
		Type:      fieldType,
		JSONTag:   jsonTag,
		DBTag:     dbTag,
		DTOTag:    dtoTag,
		IsPointer: isPointer,
	}
}

func extractStructFields(path string, structName string) []StructField {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
	if err != nil {
		log.Printf("parse error in %s: %v", path, err)
		return nil
	}

	var fields []StructField
	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != structName {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range st.Fields.List {
				if processedField := processStructField(field); processedField != nil {
					fields = append(fields, *processedField)
				}
			}
		}
	}
	return fields
}

func extractTag(tagString, key string) string {
	tagString = strings.Trim(tagString, "`")
	for _, tag := range strings.Fields(tagString) {
		if strings.HasPrefix(tag, key+":") {
			value := strings.TrimPrefix(tag, key+":")
			value = strings.Trim(value, `"`)
			return value
		}
	}
	return ""
}

func extractStructFieldsFromAST(st *ast.StructType) []StructField {
	var fields []StructField

	for _, field := range st.Fields.List {
		if processedField := processStructField(field); processedField != nil {
			fields = append(fields, *processedField)
		}
	}

	return fields
}

func extractDTOsFromResourceFile(path string) map[string]DTOSchema {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
	if err != nil {
		log.Printf("parse error in %s: %v", path, err)
		return nil
	}

	dtos := make(map[string]DTOSchema)

	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if !strings.HasSuffix(ts.Name.Name, "DTO") {
				continue
			}

			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			fields := extractStructFieldsFromAST(st)
			dtos[ts.Name.Name] = DTOSchema{
				Name:   ts.Name.Name,
				Fields: fields,
			}
		}
	}

	return dtos
}
