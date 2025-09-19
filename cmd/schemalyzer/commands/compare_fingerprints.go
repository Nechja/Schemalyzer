package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nechja/schemalyzer/internal/fingerprint"
	"github.com/spf13/cobra"
)

var (
	sourceFingerprint string
	targetFingerprint string
	cfSourceType      string
	cfSourceConn      string
	cfSourceSchema    string
	cfTargetType      string
	cfTargetConn      string
	cfTargetSchema    string
	cfJSON            bool
	cfTablesOnly      bool
)

var compareFingerprintsCmd = &cobra.Command{
	Use:   "compare-fingerprints",
	Short: "Compare fingerprints of two database schemas",
	Long:  `Quickly compare two database schemas by generating and comparing their SHA256 fingerprints`,
	RunE:  runCompareFingerprints,
}

func init() {
	compareFingerprintsCmd.Flags().StringVar(&sourceFingerprint, "source-hash", "", "Pre-computed source fingerprint (optional)")
	compareFingerprintsCmd.Flags().StringVar(&targetFingerprint, "target-hash", "", "Pre-computed target fingerprint (optional)")
	compareFingerprintsCmd.Flags().StringVar(&cfSourceType, "source-type", "", "Source database type")
	compareFingerprintsCmd.Flags().StringVar(&cfSourceConn, "source-conn", "", "Source database connection")
	compareFingerprintsCmd.Flags().StringVar(&cfSourceSchema, "source-schema", "", "Source schema name")
	compareFingerprintsCmd.Flags().StringVar(&cfTargetType, "target-type", "", "Target database type")
	compareFingerprintsCmd.Flags().StringVar(&cfTargetConn, "target-conn", "", "Target database connection")
	compareFingerprintsCmd.Flags().StringVar(&cfTargetSchema, "target-schema", "", "Target schema name")
	compareFingerprintsCmd.Flags().BoolVar(&cfJSON, "json", false, "Output in JSON format")
	compareFingerprintsCmd.Flags().BoolVar(&cfTablesOnly, "tables-only", false, "Include only tables in fingerprints")
}

func runCompareFingerprints(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	var sourceHash, targetHash string
	var err error
	
	if sourceFingerprint != "" {
		sourceHash = sourceFingerprint
	} else {
		if cfSourceType == "" || cfSourceConn == "" || cfSourceSchema == "" {
			return fmt.Errorf("source database connection details required when source-hash not provided")
		}
		sourceHash, err = generateFingerprint(ctx, cfSourceType, cfSourceConn, cfSourceSchema, cfTablesOnly)
		if err != nil {
			return fmt.Errorf("failed to generate source fingerprint: %w", err)
		}
	}
	
	if targetFingerprint != "" {
		targetHash = targetFingerprint
	} else {
		if cfTargetType == "" || cfTargetConn == "" || cfTargetSchema == "" {
			return fmt.Errorf("target database connection details required when target-hash not provided")
		}
		targetHash, err = generateFingerprint(ctx, cfTargetType, cfTargetConn, cfTargetSchema, cfTablesOnly)
		if err != nil {
			return fmt.Errorf("failed to generate target fingerprint: %w", err)
		}
	}
	
	match := sourceHash == targetHash
	
	if cfJSON {
		output := struct {
			SourceFingerprint string    `json:"source_fingerprint"`
			TargetFingerprint string    `json:"target_fingerprint"`
			Match             bool      `json:"match"`
			Timestamp         time.Time `json:"timestamp"`
			SourceSchema      string    `json:"source_schema,omitempty"`
			TargetSchema      string    `json:"target_schema,omitempty"`
		}{
			SourceFingerprint: sourceHash,
			TargetFingerprint: targetHash,
			Match:             match,
			Timestamp:         time.Now(),
		}
		
		if cfSourceSchema != "" {
			output.SourceSchema = fmt.Sprintf("%s://%s", cfSourceType, cfSourceSchema)
		}
		if cfTargetSchema != "" {
			output.TargetSchema = fmt.Sprintf("%s://%s", cfTargetType, cfTargetSchema)
		}
		
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("Source Fingerprint: %s\n", sourceHash)
		fmt.Printf("Target Fingerprint: %s\n", targetHash)
		if match {
			fmt.Println("\n✓ Schemas match!")
		} else {
			fmt.Println("\n✗ Schemas differ")
			fmt.Println("\nRun 'schemalyzer compare' for detailed differences")
		}
	}
	
	if !match {
		// Return error to trigger exit code 2
		return NewExitError(ExitCodeMismatch, "schemas do not match")
	}
	
	return nil
}

func generateFingerprint(ctx context.Context, dbType, conn, schema string, tablesOnly bool) (string, error) {
	reader, err := createReader(dbType)
	if err != nil {
		return "", fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	if err := reader.Connect(ctx, conn); err != nil {
		return "", fmt.Errorf("failed to connect to database: %w", err)
	}
	
	schemaData, err := reader.GetSchema(ctx, schema)
	if err != nil {
		return "", fmt.Errorf("failed to read schema: %w", err)
	}
	
	if tablesOnly {
		schemaData = filterTablesOnly(schemaData)
	}
	
	hasher := fingerprint.NewHasher()
	return hasher.GenerateFingerprint(schemaData)
}