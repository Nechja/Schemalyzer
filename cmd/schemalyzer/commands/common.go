package commands

import (
	"fmt"
	
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