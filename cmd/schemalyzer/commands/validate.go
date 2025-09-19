package commands

import (
	"context"
	"fmt"
	"os"
	
	"github.com/nechja/schemalyzer/internal/compare"
	"github.com/nechja/schemalyzer/internal/schema"
	"github.com/nechja/schemalyzer/pkg/models"
	"github.com/spf13/cobra"
)

var (
	goldenFile   string
	pipelineMode bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate database schema against a golden file",
	Long:  `Validate database schema against a golden JSON/YAML file.
Perfect for CI/CD pipelines - returns exit code 0 if schemas match, 2 if they differ.`,
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().StringVar(&sourceType, "type", "", "Database type (postgresql, mysql, oracle)")
	validateCmd.Flags().StringVar(&sourceConn, "conn", "", "Database connection string")
	validateCmd.Flags().StringVar(&sourceSchema, "schema", "", "Schema name to validate")
	validateCmd.Flags().StringVar(&goldenFile, "golden", "", "Golden schema file (JSON or YAML)")
	validateCmd.Flags().BoolVar(&pipelineMode, "pipeline", false, "Pipeline mode: minimal output, only exit codes")
	validateCmd.Flags().StringSliceVar(&ignorePatterns, "ignore", []string{}, "Ignore patterns (e.g., 'table:temp_*', 'constraint:SYS_*', '*_audit')")
	_ = validateCmd.MarkFlagRequired("type")
	_ = validateCmd.MarkFlagRequired("conn")
	_ = validateCmd.MarkFlagRequired("schema")
	_ = validateCmd.MarkFlagRequired("golden")
}

func runValidate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Load golden schema from file
	loader := schema.NewLoader()
	goldenSchema, err := loader.LoadFromFile(goldenFile)
	if err != nil {
		return fmt.Errorf("failed to load golden schema: %w", err)
	}
	
	// Create reader for current database
	reader, err := createReader(sourceType)
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	// Connect to database
	if err := reader.Connect(ctx, sourceConn); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Get current schema
	currentSchema, err := reader.GetSchema(ctx, sourceSchema)
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}
	
	// Compare schemas
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
	
	result := comparer.Compare(goldenSchema, currentSchema)
	
	// In pipeline mode, only output if there are differences
	if pipelineMode {
		if len(result.Differences) > 0 {
			fmt.Fprintf(os.Stderr, "FAIL: Schema validation failed - %d differences found\n", len(result.Differences))
			return NewExitErrorf(ExitCodeMismatch, "validation failed with %d differences", len(result.Differences))
		}
		// Success - no output
		return nil
	}
	
	// Normal mode - output results
	if len(result.Differences) == 0 {
		fmt.Println("✓ Schema matches golden file")
		return nil
	}
	
	// Output differences
	fmt.Printf("✗ Schema validation failed - %d differences found:\n\n", len(result.Differences))
	
	// Group by type
	added := []models.Difference{}
	removed := []models.Difference{}
	modified := []models.Difference{}
	
	for _, diff := range result.Differences {
		switch diff.Type {
		case models.Added:
			added = append(added, diff)
		case models.Removed:
			removed = append(removed, diff)
		case models.Modified:
			modified = append(modified, diff)
		}
	}
	
	// Output grouped differences
	if len(removed) > 0 {
		fmt.Println("Missing from current schema:")
		for _, diff := range removed {
			fmt.Printf("  - %s: %s\n", diff.ObjectType, diff.ObjectName)
		}
		fmt.Println()
	}
	
	if len(added) > 0 {
		fmt.Println("Extra in current schema:")
		for _, diff := range added {
			fmt.Printf("  + %s: %s\n", diff.ObjectType, diff.ObjectName)
		}
		fmt.Println()
	}
	
	if len(modified) > 0 {
		fmt.Println("Modified in current schema:")
		for _, diff := range modified {
			fmt.Printf("  ~ %s: %s\n", diff.ObjectType, diff.ObjectName)
		}
		fmt.Println()
	}
	
	// Return error with exit code for differences
	return NewExitErrorf(ExitCodeMismatch, "validation failed with %d differences", len(result.Differences))
}