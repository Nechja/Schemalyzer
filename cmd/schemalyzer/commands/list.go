package commands

import (
	"context"
	"fmt"
	
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List schemas in a database",
	Long:  `List all available schemas in the specified database`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&sourceType, "type", "", "Database type (postgresql, mysql, oracle)")
	listCmd.Flags().StringVar(&sourceConn, "conn", "", "Database connection string")
	listCmd.MarkFlagRequired("type")
	listCmd.MarkFlagRequired("conn")
}

func runList(cmd *cobra.Command, args []string) error {
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
	
	// List schemas
	schemas, err := reader.ListSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}
	
	fmt.Println("Available schemas:")
	for _, schema := range schemas {
		fmt.Printf("  - %s\n", schema)
	}
	
	return nil
}