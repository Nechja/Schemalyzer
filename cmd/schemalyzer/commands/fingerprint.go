package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nechja/schemalyzer/internal/fingerprint"
	"github.com/spf13/cobra"
)

var (
	fingerprintType   string
	fingerprintConn   string
	fingerprintSchema string
	fingerprintVerbose bool
	fingerprintJSON   bool
	fingerprintTablesOnly bool
)

var fingerprintCmd = &cobra.Command{
	Use:   "fingerprint",
	Short: "Generate a fingerprint hash of a database schema",
	Long:  `Generate a SHA256 hash of a database schema structure for quick comparison`,
	RunE:  runFingerprint,
}

func init() {
	fingerprintCmd.Flags().StringVar(&fingerprintType, "type", "", "Database type (postgresql, mysql, oracle)")
	fingerprintCmd.Flags().StringVar(&fingerprintConn, "conn", "", "Database connection string")
	fingerprintCmd.Flags().StringVar(&fingerprintSchema, "schema", "", "Schema name")
	fingerprintCmd.Flags().BoolVar(&fingerprintVerbose, "verbose", false, "Show detailed information about what's included in the hash")
	fingerprintCmd.Flags().BoolVar(&fingerprintJSON, "json", false, "Output in JSON format with metadata")
	fingerprintCmd.Flags().BoolVar(&fingerprintTablesOnly, "tables-only", false, "Include only tables in the fingerprint (no procedures, functions, triggers)")
	
	_ = fingerprintCmd.MarkFlagRequired("type")
	_ = fingerprintCmd.MarkFlagRequired("conn")
	_ = fingerprintCmd.MarkFlagRequired("schema")
}

func runFingerprint(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	reader, err := createReader(fingerprintType)
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	if err := reader.Connect(ctx, fingerprintConn); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	if fingerprintVerbose {
		fmt.Fprintf(os.Stderr, "Reading schema: %s\n", fingerprintSchema)
	}
	
	schema, err := reader.GetSchema(ctx, fingerprintSchema)
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}
	
	if fingerprintTablesOnly {
		schema = filterTablesOnly(schema)
	}
	
	hasher := fingerprint.NewHasher().WithVerbose(fingerprintVerbose)
	hash, err := hasher.GenerateFingerprint(schema)
	if err != nil {
		return fmt.Errorf("failed to generate fingerprint: %w", err)
	}
	
	if fingerprintJSON {
		output := struct {
			Database     string    `json:"database"`
			DatabaseType string    `json:"database_type"`
			Schema       string    `json:"schema"`
			Fingerprint  string    `json:"fingerprint"`
			Algorithm    string    `json:"algorithm"`
			Timestamp    time.Time `json:"timestamp"`
			TablesOnly   bool      `json:"tables_only"`
			Statistics   struct {
				Tables     int `json:"tables"`
				Views      int `json:"views"`
				Indexes    int `json:"indexes"`
				Sequences  int `json:"sequences"`
				Procedures int `json:"procedures"`
				Functions  int `json:"functions"`
				Triggers   int `json:"triggers"`
			} `json:"statistics"`
		}{
			Database:     fingerprintConn,
			DatabaseType: fingerprintType,
			Schema:       fingerprintSchema,
			Fingerprint:  hash,
			Algorithm:    "SHA256",
			Timestamp:    time.Now(),
			TablesOnly:   fingerprintTablesOnly,
		}
		
		output.Statistics.Tables = len(schema.Tables)
		output.Statistics.Views = len(schema.Views)
		output.Statistics.Indexes = len(schema.Indexes)
		output.Statistics.Sequences = len(schema.Sequences)
		output.Statistics.Procedures = len(schema.Procedures)
		output.Statistics.Functions = len(schema.Functions)
		output.Statistics.Triggers = len(schema.Triggers)
		
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else if fingerprintVerbose {
		fmt.Printf("Database Type: %s\n", fingerprintType)
		fmt.Printf("Schema: %s\n", fingerprintSchema)
		fmt.Printf("Algorithm: SHA256\n")
		fmt.Printf("Tables: %d\n", len(schema.Tables))
		fmt.Printf("Views: %d\n", len(schema.Views))
		fmt.Printf("Indexes: %d\n", len(schema.Indexes))
		fmt.Printf("Sequences: %d\n", len(schema.Sequences))
		fmt.Printf("Procedures: %d\n", len(schema.Procedures))
		fmt.Printf("Functions: %d\n", len(schema.Functions))
		fmt.Printf("Triggers: %d\n", len(schema.Triggers))
		fmt.Printf("\nFingerprint: %s\n", hash)
	} else {
		fmt.Println(hash)
	}
	
	return nil
}