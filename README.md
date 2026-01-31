# GoREST Codegen Plugin

[![CI](https://github.com/nicolasbonnici/gorest-codegen/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-codegen/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-codegen)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-codegen)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Code generation plugin for [GoREST](https://github.com/nicolasbonnici/gorest) that automatically generates models, REST resources, DTOs, and OpenAPI schemas from your database schema.

## Features

- **Model Generation**: Generate Go structs from database tables with proper field types and JSON tags
- **Resource Generation**: Create complete REST API handlers with CRUD operations
- **DTO Generation**: Automatically generate Data Transfer Objects for API requests/responses
- **OpenAPI Schema**: Generate OpenAPI 3.0 specification from your database schema
- **Multi-Database Support**: Works with PostgreSQL, MySQL, and SQLite
- **Plugin Architecture**: Integrates seamlessly with GoREST's plugin system

## Installation

```bash
go get github.com/nicolasbonnici/gorest-codegen@latest
```

## Usage

### As a Plugin

Register the codegen plugin in your GoREST application:

```go
import (
    "github.com/nicolasbonnici/gorest"
    codegenPlugin "github.com/nicolasbonnici/gorest-codegen"
    "github.com/nicolasbonnici/gorest/pluginloader"
)

func init() {
    pluginloader.RegisterPluginFactory("codegen", codegenPlugin.NewPlugin)
}

func main() {
    app := gorest.New()
    // ... configure your app ...
    app.Start()
}
```

Then enable it in your `gorest.yaml`:

```yaml
plugins:
  - name: codegen
    enabled: true
```

The plugin will be initialized with access to your database and configuration, allowing you to invoke commands programmatically.

### As a Standalone CLI

Build the standalone binary:

```bash
cd cmd/codegen
go build -o codegen
```

Run code generation commands:

```bash
# Generate models from database schema
./codegen models

# Generate REST resources and DTOs
./codegen resources

# Generate OpenAPI schema
./codegen openapi

# Run all generation steps
./codegen all
```

### Using go run

You can also run it directly without building:

```bash
go run github.com/nicolasbonnici/gorest-codegen/cmd/codegen@latest all
```

## Commands

### models

Generates Go struct models from your database schema.

```bash
codegen models
```

Output location: `generated/models/` (configurable in `gorest.yaml`)

### resources

Generates REST API resource handlers and DTOs from your models.

```bash
codegen resources
```

Output location:
- Resources: `generated/resources/`
- DTOs: `generated/dtos/`

### openapi

Generates OpenAPI 3.0 schema from your database schema.

```bash
codegen openapi
```

Output location: `generated/openapi/schema.yaml`

### all

Runs all code generation steps in sequence.

```bash
codegen all
```

## Configuration

Configure code generation in your `gorest.yaml`:

```yaml
database:
  url: "postgres://user:pass@localhost:5432/mydb"

codegen:
  output:
    models: "generated/models"
    resources: "generated/resources"
    dtos: "generated/dtos"
    openapi: "generated/openapi"

  auth:
    enabled: true
    user_model: "User"
    identifier_field: "Email"
    password_field: "Password"
```

## Example Workflow

1. **Design your database schema**

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    firstname VARCHAR(100),
    lastname VARCHAR(100),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

2. **Run code generation**

```bash
codegen all
```

3. **Generated files**

```
generated/
├── models/
│   ├── user.go       # User model struct
│   └── post.go       # Post model struct
├── resources/
│   ├── user.go       # User REST handlers (List, Get, Create, Update, Delete)
│   ├── post.go       # Post REST handlers
│   └── routes.go     # Route registration
├── dtos/
│   ├── user.go       # User DTOs (CreateUserDTO, UpdateUserDTO)
│   └── post.go       # Post DTOs
└── openapi/
    └── schema.yaml   # OpenAPI 3.0 specification
```

4. **Use generated code in your app**

```go
import (
    "github.com/nicolasbonnici/gorest"
    "yourproject/generated/routes"
)

func main() {
    app := gorest.New()

    // Register generated routes
    routes.RegisterRoutes(app)

    app.Start()
}
```

## Integration with Other Plugins

The codegen plugin can be used with other GoREST plugins like `gorest-benchmark`:

```go
// In gorest-benchmark
import codegenPlugin "github.com/nicolasbonnici/gorest-codegen"

func (p *BenchmarkPlugin) Dependencies() []string {
    return []string{"codegen"}
}

// Access codegen via dependency injection
func (p *BenchmarkPlugin) Initialize(cfg map[string]any) error {
    if deps, ok := cfg[plugin.ConfigKeyDependencies]; ok {
        if depsMap, ok := deps.(map[string]plugin.Plugin); ok {
            p.codegenPlugin = depsMap["codegen"]
        }
    }
    return nil
}
```

## Testing

Run the integration tests:

```bash
# Set up test database
export DATABASE_URL_TEST="postgres://postgres:postgres@localhost:5433/mydb_test?sslmode=disable"

# Run tests
go test -tags=integration ./codegen
```

Tests require a running database instance. The test suite supports PostgreSQL, MySQL, and SQLite.

## Development

### Building from Source

```bash
git clone https://github.com/nicolasbonnici/gorest-codegen.git
cd gorest-codegen
go mod download
go build ./...
```

### Running Tests

```bash
# Unit tests (if any)
go test ./...

# Integration tests with PostgreSQL
DATABASE_URL_TEST="postgres://localhost/test" go test -tags=integration ./codegen

# Integration tests with SQLite
DATABASE_URL_TEST="sqlite::memory:" go test -tags=integration ./codegen
```

## Requirements

- Go 1.25.1 or later
- GoREST v0.4.8 or later
- Database (PostgreSQL, MySQL, or SQLite)

---

## Git Hooks

This directory contains git hooks for the GoREST plugin to maintain code quality.

### Available Hooks

#### pre-commit

Runs before each commit to ensure code quality:
- **Linting**: Runs `make lint` to check code style and potential issues
- **Tests**: Runs `make test` to verify all tests pass

### Installation

#### Automatic Installation

Run the install script from the project root:

```bash
./.githooks/install.sh
```

#### Manual Installation

Copy the hooks to your `.git/hooks` directory:

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---


## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

- [GoREST Documentation](https://github.com/nicolasbonnici/gorest)
- [Report Issues](https://github.com/nicolasbonnici/gorest-codegen/issues)

## Related Projects

- [GoREST](https://github.com/nicolasbonnici/gorest) - The main framework
- [gorest-auth](https://github.com/nicolasbonnici/gorest-auth) - Authentication plugin
- [gorest-benchmark](https://github.com/nicolasbonnici/gorest-benchmark) - Benchmarking plugin
