package database

import (
	"context"
	"github.com/nechja/schemalyzer/pkg/models"
)

type SchemaReader interface {
	Connect(ctx context.Context, connectionString string) error
	GetSchema(ctx context.Context, schemaName string) (*models.Schema, error)
	ListSchemas(ctx context.Context) ([]string, error)
	Close() error
}

type SchemaComparer interface {
	Compare(source, target *models.Schema) *models.ComparisonResult
}

type OutputFormatter interface {
	Format(result *models.ComparisonResult) ([]byte, error)
}