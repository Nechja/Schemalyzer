package commands

import (
	"context"
	"fmt"
	"os"
	
	"github.com/nechja/schemalyzer/internal/schema"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export database schema to file",
	Long:  `Export database schema to JSON or YAML file`,
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVar(&sourceType, "type", "", "Database type (postgresql, mysql, oracle)")
	exportCmd.Flags().StringVar(&sourceConn, "conn", "", "Database connection string")
	exportCmd.Flags().StringVar(&sourceSchema, "schema", "", "Schema name to export")
	exportCmd.Flags().StringVar(&outputFile, "output", "", "Output file path (required)")
	exportCmd.Flags().BoolVar(&tablesOnly, "tables-only", false, "Export only tables and their structure (no procedures, functions, triggers)")
	_ = exportCmd.MarkFlagRequired("type")
	_ = exportCmd.MarkFlagRequired("conn")
	_ = exportCmd.MarkFlagRequired("schema")
	_ = exportCmd.MarkFlagRequired("output")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Create reader
	reader, err := createReader(sourceType)
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	// Connect
	if err := reader.Connect(ctx, sourceConn); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Get schema
	fmt.Fprintf(os.Stderr, "Reading schema: %s\n", sourceSchema)
	schemaData, err := reader.GetSchema(ctx, sourceSchema)
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}
	
	// Filter schema if --tables-only is set
	if tablesOnly {
		schemaData = filterTablesOnly(schemaData)
	}
	
	// Save to file
	loader := schema.NewLoader()
	if err := loader.SaveToFile(schemaData, outputFile); err != nil {
		return fmt.Errorf("failed to save schema: %w", err)
	}
	
	fmt.Fprintf(os.Stderr, "Schema exported to: %s\n", outputFile)
	return nil
}