package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/nechja/schemalyzer/internal/docs"
	"github.com/spf13/cobra"
)

var docFormat string

var documentCmd = &cobra.Command{
	Use:   "document",
	Short: "Generate visual documentation for database schema",
	Long:  `Generate visual documentation in various formats including PlantUML, GraphViz, Mermaid, and Markdown.`,
	RunE:  runDocument,
}

func init() {
	documentCmd.Flags().StringVar(&sourceType, "type", "", "Database type (postgresql, mysql, oracle)")
	documentCmd.Flags().StringVar(&sourceConn, "conn", "", "Database connection string")
	documentCmd.Flags().StringVar(&sourceSchema, "schema", "", "Schema name to document")
	documentCmd.Flags().StringVar(&docFormat, "format", "markdown", "Documentation format (markdown, plantuml, mermaid, graphviz, d2)")
	documentCmd.Flags().StringVar(&outputFile, "output", "", "Output file path (required)")
	documentCmd.Flags().BoolVar(&tablesOnly, "tables-only", false, "Document only tables and their structure (no procedures, functions, triggers)")
	_ = documentCmd.MarkFlagRequired("type")
	_ = documentCmd.MarkFlagRequired("conn")
	_ = documentCmd.MarkFlagRequired("schema")
	_ = documentCmd.MarkFlagRequired("output")
}

func runDocument(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Create reader
	reader, err := createReader(sourceType)
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	// Connect to database
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
	
	// Generate documentation
	var generator docs.DocumentGenerator
	switch strings.ToLower(docFormat) {
	case "plantuml":
		generator = docs.NewPlantUMLGenerator()
	case "mermaid":
		generator = docs.NewMermaidGenerator()
	case "graphviz", "dot":
		generator = docs.NewGraphVizGenerator()
	case "d2":
		generator = docs.NewD2Generator()
	case "markdown", "md":
		generator = docs.NewMarkdownDocGenerator()
	default:
		return fmt.Errorf("unsupported documentation format: %s", docFormat)
	}
	
	fmt.Fprintf(os.Stderr, "Generating %s documentation...\n", docFormat)
	docContent, err := generator.Generate(schemaData)
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}
	
	// Write to file
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	if err := os.WriteFile(outputFile, []byte(docContent), 0644); err != nil {
		return fmt.Errorf("failed to write documentation: %w", err)
	}
	
	fmt.Fprintf(os.Stderr, "Documentation written to: %s\n", outputFile)
	
	// Provide usage instructions based on format
	switch strings.ToLower(docFormat) {
	case "plantuml":
		fmt.Fprintf(os.Stderr, "\nTo generate an image:\n")
		fmt.Fprintf(os.Stderr, "  plantuml %s\n", outputFile)
		fmt.Fprintf(os.Stderr, "  # or use online: http://www.plantuml.com/plantuml\n")
	case "graphviz", "dot":
		fmt.Fprintf(os.Stderr, "\nTo generate an image:\n")
		fmt.Fprintf(os.Stderr, "  dot -Tpng %s -o schema.png\n", outputFile)
		fmt.Fprintf(os.Stderr, "  dot -Tsvg %s -o schema.svg\n", outputFile)
	case "d2":
		fmt.Fprintf(os.Stderr, "\nTo generate an image:\n")
		fmt.Fprintf(os.Stderr, "  d2 %s schema.png\n", outputFile)
		fmt.Fprintf(os.Stderr, "  # Install: https://d2lang.com/tour/install\n")
	case "mermaid":
		fmt.Fprintf(os.Stderr, "\nView in GitHub/GitLab or use:\n")
		fmt.Fprintf(os.Stderr, "  mmdc -i %s -o schema.png\n", outputFile)
		fmt.Fprintf(os.Stderr, "  # Install: npm install -g @mermaid-js/mermaid-cli\n")
	}
	
	return nil
}