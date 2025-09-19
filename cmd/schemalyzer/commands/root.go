package commands

import (
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "schemalyzer",
	Short: "A schema comparison tool for PostgreSQL, MySQL, and Oracle databases",
	Long: `Schemalyzer is a database schema comparison tool that reads and compares
schemas between PostgreSQL, MySQL, and Oracle databases. It's designed
to be used in CI/CD pipelines for database change detection.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// Check if it's an ExitError with a specific exit code
		if exitErr, ok := err.(*ExitError); ok {
			// For ExitError, we've already printed output in the command
			// Just exit with the specified code
			os.Exit(exitErr.Code)
		}
		// For other errors, print and exit with 1
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(compareCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(exportCmd)
	RootCmd.AddCommand(validateCmd)
	RootCmd.AddCommand(documentCmd)
	RootCmd.AddCommand(fingerprintCmd)
	RootCmd.AddCommand(compareFingerprintsCmd)
}