package codegen

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// extractStructFieldsUncached reproduces the pre-optimization per-call cost:
// a fresh FileSet + full parse on every invocation.
func extractStructFieldsUncached(path, structName string) []StructField {
	node, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors)
	if err != nil {
		return nil
	}
	return fieldsForStruct(node, structName)
}

func TestCachedFieldsMatchUncached(t *testing.T) {
	path := writeBenchModel(t)
	want := extractStructFieldsUncached(path, "User")
	for i := 0; i < 3; i++ {
		if got := extractStructFields(path, "User"); !reflect.DeepEqual(got, want) {
			t.Fatalf("cached fields differ from uncached on call %d:\n got %#v\nwant %#v", i, got, want)
		}
	}
}

func BenchmarkExtractStructFieldsUncached(b *testing.B) {
	path := writeBenchModel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := extractStructFieldsUncached(path, "User"); len(got) != 8 {
			b.Fatalf("expected 8 fields, got %d", len(got))
		}
	}
}

const benchModel = `package models

import "time"

type User struct {
	Id        string     ` + "`json:\"id,omitempty\" db:\"id\"`" + `
	Email     string     ` + "`json:\"email\" db:\"email\"`" + `
	Firstname string     ` + "`json:\"firstname\" db:\"firstname\"`" + `
	Lastname  string     ` + "`json:\"lastname\" db:\"lastname\"`" + `
	Password  string     ` + "`json:\"password,omitempty\" db:\"password\" dto:\"write\"`" + `
	UserId    int64      ` + "`json:\"userId\" db:\"user_id\"`" + `
	CreatedAt *time.Time ` + "`json:\"created_at,omitempty\" db:\"created_at\"`" + `
	UpdatedAt *time.Time ` + "`json:\"updated_at,omitempty\" db:\"updated_at\"`" + `
}

func (User) TableName() string { return "users" }
`

func writeBenchModel(tb testing.TB) string {
	tb.Helper()
	dir := tb.TempDir()
	path := filepath.Join(dir, "user.go")
	if err := os.WriteFile(path, []byte(benchModel), 0644); err != nil {
		tb.Fatalf("write model: %v", err)
	}
	return path
}

func BenchmarkExtractStructFields(b *testing.B) {
	path := writeBenchModel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := extractStructFields(path, "User"); len(got) != 8 {
			b.Fatalf("expected 8 fields, got %d", len(got))
		}
	}
}

func BenchmarkParseStructs(b *testing.B) {
	path := writeBenchModel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := parseStructs(path); len(got) != 1 {
			b.Fatalf("expected 1 struct, got %d", len(got))
		}
	}
}
