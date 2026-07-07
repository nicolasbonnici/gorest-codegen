package codegen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// fileFingerprint identifies a source file by size and modification time so
// cached parse results are reused only while the file on disk is unchanged.
type fileFingerprint struct {
	size    int64
	modUnix int64
}

// astCacheEntry holds a parsed source file together with lazily-derived struct
// field metadata. The generators parse each model file once and then read its
// fields multiple times (for the DTO and the resource), so both the AST and the
// per-struct field lists are worth caching.
type astCacheEntry struct {
	fp     fileFingerprint
	file   *ast.File
	mu     sync.Mutex
	fields map[string][]StructField
}

var (
	astCacheMu sync.RWMutex
	astCache   = map[string]*astCacheEntry{}
)

// loadASTEntry parses path and caches the result, re-parsing only when the
// file's fingerprint changes. On stat failure it falls back to an uncached
// parse so the caller still observes the underlying parse error.
func loadASTEntry(path string) (*astCacheEntry, error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors)
		if err != nil {
			return nil, err
		}
		return &astCacheEntry{file: file}, nil
	}

	fp := fileFingerprint{size: info.Size(), modUnix: info.ModTime().UnixNano()}

	astCacheMu.RLock()
	entry, ok := astCache[path]
	astCacheMu.RUnlock()
	if ok && entry.fp == fp {
		return entry, nil
	}

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	entry = &astCacheEntry{fp: fp, file: file}

	astCacheMu.Lock()
	astCache[path] = entry
	astCacheMu.Unlock()
	return entry, nil
}

// structFields returns the cached field metadata for structName, computing it
// from the parsed file on first request. The returned slice is shared and must
// be treated as read-only by callers.
func (e *astCacheEntry) structFields(structName string) []StructField {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.fields != nil {
		if f, ok := e.fields[structName]; ok {
			return f
		}
	}

	f := fieldsForStruct(e.file, structName)
	if e.fields == nil {
		e.fields = make(map[string][]StructField)
	}
	e.fields[structName] = f
	return f
}

// titleCaserPool reuses cases.Caser instances, which are relatively expensive to
// construct and not safe for concurrent use; the pool hands each caller its own.
var titleCaserPool = sync.Pool{New: func() any {
	c := cases.Title(language.English)
	return &c
}}

func borrowTitleCaser() (*cases.Caser, func()) {
	c := titleCaserPool.Get().(*cases.Caser)
	return c, func() { titleCaserPool.Put(c) }
}

// builderPool reuses byte buffers for the small per-field emit helpers. The
// helpers return bytes.Buffer.String(), which copies, so a buffer is safe to
// reset and return to the pool immediately afterwards.
var builderPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func borrowBuilder() (*bytes.Buffer, func()) {
	b := builderPool.Get().(*bytes.Buffer)
	b.Reset()
	return b, func() { builderPool.Put(b) }
}
