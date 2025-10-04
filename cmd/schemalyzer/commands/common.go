package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nechja/schemalyzer/internal/database"
	"github.com/nechja/schemalyzer/internal/database/mysql"
	"github.com/nechja/schemalyzer/internal/database/oracle"
	"github.com/nechja/schemalyzer/internal/database/postgres"
	"github.com/nechja/schemalyzer/pkg/models"
)

func createReader(dbType string) (database.SchemaReader, error) {
	switch models.DatabaseType(dbType) {
	case models.PostgreSQL:
		return postgres.NewPostgresReader(), nil
	case models.MySQL:
		return mysql.NewMySQLReader(), nil
	case models.Oracle:
		return oracle.NewOracleReader(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// filterTablesOnly returns a copy of the schema with only tables and views
func filterTablesOnly(schema *models.Schema) *models.Schema {
	filtered := &models.Schema{
		Name:         schema.Name,
		DatabaseType: schema.DatabaseType,
		Tables:       schema.Tables,
		Views:        schema.Views,
		// Exclude these when --tables-only is set:
		Sequences:  []models.Sequence{},
		Functions:  []models.Function{},
		Procedures: []models.Procedure{},
		Triggers:   []models.Trigger{},
	}
	return filtered
}

// collectStatistics collects additional statistics for the schema
func collectStatistics(ctx context.Context, reader database.SchemaReader, schema *models.Schema) error {
	// Add overall statistics if requested
	if withStats {
		totalColumns := 0
		for _, table := range schema.Tables {
			totalColumns += len(table.Columns)
		}

		schema.Stats = &models.SchemaStats{
			TableCount:   len(schema.Tables),
			ViewCount:    len(schema.Views),
			TotalColumns: totalColumns,
			IndexCount:   len(schema.Indexes),
			GeneratedAt:  time.Now(),
		}
	}

	// Collect row counts and samples if requested
	if withRowCount || withSamples {
		statsReader, ok := reader.(database.StatisticsReader)
		if !ok {
			return fmt.Errorf("statistics not supported for this database type")
		}

		for i := range schema.Tables {
			table := &schema.Tables[i]

			// Get row count
			if withRowCount {
				count, err := statsReader.GetTableRowCount(ctx, schema.Name, table.Name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to get row count for table %s: %v\n", table.Name, err)
				} else {
					table.RowCount = &count
				}
			}

			// Get sample values
			if withSamples {
				for j := range table.Columns {
					column := &table.Columns[j]
					samples, err := statsReader.GetColumnSamples(ctx, schema.Name, table.Name, column.Name, sampleSize)
					if err != nil {
						// Silently skip columns where we can't get samples
						continue
					}
					column.Samples = samples
				}
			}
		}
	}

	return nil
}