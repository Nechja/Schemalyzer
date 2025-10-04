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
	exportCmd.Flags().BoolVar(&withStats, "with-stats", false, "Include schema statistics (table count, column count, etc.)")
	exportCmd.Flags().BoolVar(&withRowCount, "with-row-count", false, "Include row counts for each table")
	exportCmd.Flags().BoolVar(&withSamples, "with-samples", false, "Include sample values for each column")
	exportCmd.Flags().IntVar(&sampleSize, "sample-size", 3, "Number of sample values to collect per column (default: 3)")
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

	// Collect statistics if requested
	if withStats || withRowCount || withSamples {
		if err := collectStatistics(ctx, reader, schemaData); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to collect some statistics: %v\n", err)
			// Continue even if statistics collection fails
		}
	}

	// Save to file
	loader := schema.NewLoader()
	if err := loader.SaveToFile(schemaData, outputFile); err != nil {
		return fmt.Errorf("failed to save schema: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Schema exported to: %s\n", outputFile)
	return nil
}