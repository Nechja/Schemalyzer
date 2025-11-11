package compare

import (
	"testing"

	"github.com/nechja/schemalyzer/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestComparer_Compare_IdenticalSchemas(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "users",
				Schema: "test",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false, Position: 1},
					{Name: "name", DataType: "varchar(100)", IsNullable: true, Position: 2},
				},
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "users",
				Schema: "test",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false, Position: 1},
					{Name: "name", DataType: "varchar(100)", IsNullable: true, Position: 2},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 0, len(result.Differences))
}

func TestComparer_Compare_AddedTable(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables:       []models.Table{},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "users",
				Schema: "test",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Added, result.Differences[0].Type)
	assert.Equal(t, "Table", result.Differences[0].ObjectType)
	assert.Equal(t, "users", result.Differences[0].ObjectName)
}

func TestComparer_Compare_RemovedTable(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "users",
				Schema: "test",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
				},
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables:       []models.Table{},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Removed, result.Differences[0].Type)
	assert.Equal(t, "Table", result.Differences[0].ObjectType)
	assert.Equal(t, "users", result.Differences[0].ObjectName)
}

func TestComparer_Compare_ModifiedColumn(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "users",
				Schema: "test",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
					{Name: "name", DataType: "varchar(50)", IsNullable: true},
				},
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "users",
				Schema: "test",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
					{Name: "name", DataType: "varchar(100)", IsNullable: true},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Modified, result.Differences[0].Type)
	assert.Equal(t, "Column", result.Differences[0].ObjectType)
	assert.Equal(t, "users.name", result.Differences[0].ObjectName)
}

func TestComparer_Compare_AutoIncrementChange(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.MySQL,
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "int", IsPrimaryKey: true, IsAutoIncrement: true},
				},
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.MySQL,
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "int", IsPrimaryKey: true, IsAutoIncrement: false},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)
	if assert.Equal(t, 1, len(result.Differences)) {
		assert.Equal(t, models.Modified, result.Differences[0].Type)
		assert.Equal(t, "Column", result.Differences[0].ObjectType)
		assert.Equal(t, "users.id", result.Differences[0].ObjectName)
	}
}

func TestComparer_Compare_AddedConstraint(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:        "users",
				Schema:      "test",
				Columns:     []models.Column{{Name: "id", DataType: "integer"}},
				Constraints: []models.Constraint{},
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:    "users",
				Schema:  "test",
				Columns: []models.Column{{Name: "id", DataType: "integer"}},
				Constraints: []models.Constraint{
					{
						Name:    "users_pkey",
						Type:    models.PrimaryKey,
						Columns: []string{"id"},
					},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Added, result.Differences[0].Type)
	assert.Equal(t, "Constraint", result.Differences[0].ObjectType)
	assert.Equal(t, "users.users_pkey", result.Differences[0].ObjectName)
}

func TestComparer_Compare_ModifiedIndex(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:    "users",
				Schema:  "test",
				Columns: []models.Column{{Name: "name", DataType: "varchar"}},
				Indexes: []models.Index{
					{
						Name:      "idx_users_name",
						TableName: "users",
						Columns:   []string{"name"},
						IsUnique:  false,
						Type:      "btree",
					},
				},
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:    "users",
				Schema:  "test",
				Columns: []models.Column{{Name: "name", DataType: "varchar"}},
				Indexes: []models.Index{
					{
						Name:      "idx_users_name",
						TableName: "users",
						Columns:   []string{"name"},
						IsUnique:  true, // Changed to unique
						Type:      "btree",
					},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Modified, result.Differences[0].Type)
	assert.Equal(t, "Index", result.Differences[0].ObjectType)
}

func TestComparer_Compare_Views(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Views: []models.View{
			{
				Name:       "user_summary",
				Schema:     "test",
				Definition: "SELECT id, name FROM users",
			},
		},
	}

	schema2 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Views: []models.View{
			{
				Name:       "user_summary",
				Schema:     "test",
				Definition: "SELECT id, name, email FROM users",
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Modified, result.Differences[0].Type)
	assert.Equal(t, "View", result.Differences[0].ObjectType)
	assert.Equal(t, "user_summary", result.Differences[0].ObjectName)
}
