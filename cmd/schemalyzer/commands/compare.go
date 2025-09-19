package commands

import (
	"context"
	"fmt"
	"os"
	
	"github.com/nechja/schemalyzer/internal/compare"
	"github.com/nechja/schemalyzer/internal/output"
	"github.com/nechja/schemalyzer/pkg/models"
	"github.com/spf13/cobra"
)

var (
	sourceType   string
	sourceConn   string
	sourceSchema string
	targetType   string
	targetConn   string
	targetSchema string
	outputFormat string
	outputFile   string
	ignorePatterns []string
	tablesOnly   bool
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two database schemas",
	Long:  `Compare schemas between two databases and output the differences`,
	RunE:  runCompare,
}

func init() {
	compareCmd.Flags().StringVar(&sourceType, "source-type", "", "Source database type (postgresql, mysql, oracle)")
	compareCmd.Flags().StringVar(&sourceConn, "source-conn", "", "Source database connection string")
	compareCmd.Flags().StringVar(&sourceSchema, "source-schema", "", "Source schema name")
	compareCmd.Flags().StringVar(&targetType, "target-type", "", "Target database type (postgresql, mysql, oracle)")
	compareCmd.Flags().StringVar(&targetConn, "target-conn", "", "Target database connection string")
	compareCmd.Flags().StringVar(&targetSchema, "target-schema", "", "Target schema name")
	compareCmd.Flags().StringVar(&outputFormat, "format", "text", "Output format (json, yaml, text, summary)")
	compareCmd.Flags().StringVar(&outputFile, "output", "", "Output file path (default: stdout)")
	compareCmd.Flags().StringSliceVar(&ignorePatterns, "ignore", []string{}, "Ignore patterns (e.g., 'table:temp_*', 'constraint:SYS_*', '*_audit')")
	compareCmd.Flags().BoolVar(&tablesOnly, "tables-only", false, "Compare only tables and their structure (no procedures, functions, triggers)")
	
	_ = compareCmd.MarkFlagRequired("source-type")
	_ = compareCmd.MarkFlagRequired("source-conn")
	_ = compareCmd.MarkFlagRequired("source-schema")
	_ = compareCmd.MarkFlagRequired("target-type")
	_ = compareCmd.MarkFlagRequired("target-conn")
	_ = compareCmd.MarkFlagRequired("target-schema")
}

func runCompare(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Create source reader
	sourceReader, err := createReader(sourceType)
	if err != nil {
		return fmt.Errorf("failed to create source reader: %w", err)
	}
	defer sourceReader.Close()
	
	// Connect to source
	if err := sourceReader.Connect(ctx, sourceConn); err != nil {
		return fmt.Errorf("failed to connect to source database: %w", err)
	}
	
	// Get source schema
	fmt.Fprintf(os.Stderr, "Reading source schema: %s\n", sourceSchema)
	sourceSchemaData, err := sourceReader.GetSchema(ctx, sourceSchema)
	if err != nil {
		return fmt.Errorf("failed to read source schema: %w", err)
	}
	
	// Create target reader
	targetReader, err := createReader(targetType)
	if err != nil {
		return fmt.Errorf("failed to create target reader: %w", err)
	}
	defer targetReader.Close()
	
	// Connect to target
	if err := targetReader.Connect(ctx, targetConn); err != nil {
		return fmt.Errorf("failed to connect to target database: %w", err)
	}
	
	// Get target schema
	fmt.Fprintf(os.Stderr, "Reading target schema: %s\n", targetSchema)
	targetSchemaData, err := targetReader.GetSchema(ctx, targetSchema)
	if err != nil {
		return fmt.Errorf("failed to read target schema: %w", err)
	}
	
	// Filter schemas if --tables-only is set
	if tablesOnly {
		sourceSchemaData = filterTablesOnly(sourceSchemaData)
		targetSchemaData = filterTablesOnly(targetSchemaData)
	}
	
	// Compare schemas
	fmt.Fprintf(os.Stderr, "Comparing schemas...\n")
	
	var comparer *compare.Comparer
	if len(ignorePatterns) > 0 {
		ignoreConfig, err := models.NewIgnoreConfig(ignorePatterns)
		if err != nil {
			return fmt.Errorf("failed to parse ignore patterns: %w", err)
		}
		comparer = compare.NewComparerWithIgnore(ignoreConfig)
	} else {
		comparer = compare.NewComparer()
	}
	
	result := comparer.Compare(sourceSchemaData, targetSchemaData)
	result.SourceDatabase = fmt.Sprintf("%s://%s", sourceType, sourceSchema)
	result.TargetDatabase = fmt.Sprintf("%s://%s", targetType, targetSchema)
	
	// Format output
	formatter := output.NewFormatter(output.OutputFormat(outputFormat))
	outputData, err := formatter.Format(result)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	
	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Output written to: %s\n", outputFile)
	} else {
		fmt.Print(string(outputData))
	}
	
	// Return error with exit code if differences found
	if len(result.Differences) > 0 {
		// Return silent error - output has already been displayed
		return NewExitError(ExitCodeMismatch, "")
	}
	
	return nil
}