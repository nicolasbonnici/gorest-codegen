package main

import (
	"fmt"
	"os"

	"github.com/nicolasbonnici/gorest/config"
	"github.com/nicolasbonnici/gorest/database"
	_ "github.com/nicolasbonnici/gorest/database/mysql"
	_ "github.com/nicolasbonnici/gorest/database/postgres"
	_ "github.com/nicolasbonnici/gorest/database/sqlite"
	"github.com/nicolasbonnici/gorest/plugin"
	codegen "github.com/nicolasbonnici/gorest-codegen"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := database.Open("", cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close database: %v\n", err)
		}
	}()

	// Initialize plugin
	p := codegen.NewPlugin()
	err = p.Initialize(map[string]any{
		"database": db,
		"config":   cfg,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}

	// Get command provider
	cmdProvider, ok := p.(plugin.CommandProvider)
	if !ok {
		fmt.Fprintf(os.Stderr, "Plugin does not support commands\n")
		os.Exit(1)
	}

	commandName := os.Args[1]

	// Find and execute command
	for _, cmd := range cmdProvider.Commands() {
		if cmd.Name() == commandName {
			result := cmd.Run(&plugin.CommandContext{
				Config: cfg,
				Args:   os.Args[2:],
				ProgressCallback: func(msg string) {
					fmt.Printf("  → %s\n", msg)
				},
			})

			if !result.Success {
				fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
				os.Exit(1)
			}
			fmt.Println(result.Message)
			os.Exit(0)
		}
	}

	// Command not found
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", commandName)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println("GoREST Code Generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  codegen <command>")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  models      Generate model structs from database schema")
	fmt.Println("  resources   Generate REST API resources and DTOs from models")
	fmt.Println("  openapi     Generate OpenAPI schema file")
	fmt.Println("  all         Run all code generation steps")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  codegen models")
	fmt.Println("  codegen resources")
	fmt.Println("  codegen all")
}
