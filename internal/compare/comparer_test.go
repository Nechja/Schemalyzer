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

func TestComparer_Compare_ForeignKeyActionChange(t *testing.T) {
	comparer := NewComparer()

	schema1 := &models.Schema{
		Name:         "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name:   "orders",
				Schema: "test",
				Constraints: []models.Constraint{
					{
						Name:             "orders_user_id_fkey",
						Type:             models.ForeignKey,
						Columns:          []string{"user_id"},
						ReferencedTable:  "users",
						ReferencedColumn: []string{"id"},
						OnDelete:         "CASCADE",
						OnUpdate:         "NO ACTION",
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
				Name:   "orders",
				Schema: "test",
				Constraints: []models.Constraint{
					{
						Name:             "orders_user_id_fkey",
						Type:             models.ForeignKey,
						Columns:          []string{"user_id"},
						ReferencedTable:  "users",
						ReferencedColumn: []string{"id"},
						OnDelete:         "SET NULL",
						OnUpdate:         "CASCADE",
					},
				},
			},
		},
	}

	result := comparer.Compare(schema1, schema2)

	if assert.Equal(t, 1, len(result.Differences)) {
		assert.Equal(t, models.Modified, result.Differences[0].Type)
		assert.Equal(t, "Constraint", result.Differences[0].ObjectType)
		assert.Equal(t, "orders.orders_user_id_fkey", result.Differences[0].ObjectName)
	}
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

// helper: a table with the given constraints, single id column.
func tableWithConstraints(cs []models.Constraint) models.Table {
	return models.Table{
		Name:        "orders",
		Schema:      "test",
		Columns:     []models.Column{{Name: "id", DataType: "integer"}, {Name: "email", DataType: "varchar(255)"}},
		Constraints: cs,
	}
}

// System-generated PK/UNIQUE/FK names (Oracle SYS_C*, Postgres OID-based) differ
// between structurally-identical schemas. They must not show as remove+add.
func TestComparer_Compare_SystemNamedConstraintsNoNoise(t *testing.T) {
	comparer := NewComparer()

	source := &models.Schema{Name: "a", DatabaseType: models.Oracle, Tables: []models.Table{
		tableWithConstraints([]models.Constraint{
			{Name: "SYS_C001", Type: models.PrimaryKey, Columns: []string{"id"}},
			{Name: "SYS_C002", Type: models.Unique, Columns: []string{"email"}},
			{Name: "SYS_C003", Type: models.ForeignKey, Columns: []string{"cust_id"}, ReferencedTable: "customers", ReferencedColumn: []string{"id"}},
		}),
	}}
	target := &models.Schema{Name: "b", DatabaseType: models.Oracle, Tables: []models.Table{
		tableWithConstraints([]models.Constraint{
			{Name: "SYS_C999", Type: models.PrimaryKey, Columns: []string{"id"}},
			{Name: "SYS_C998", Type: models.Unique, Columns: []string{"email"}},
			{Name: "SYS_C997", Type: models.ForeignKey, Columns: []string{"cust_id"}, ReferencedTable: "customers", ReferencedColumn: []string{"id"}},
		}),
	}}

	result := comparer.Compare(source, target)
	assert.Equal(t, 0, len(result.Differences), "system-named PK/UNIQUE/FK with identical structure should produce no diff")
}

// A changed CHECK expression must remain a single MODIFIED, not remove+add.
func TestComparer_Compare_ModifiedCheckStaysSingle(t *testing.T) {
	comparer := NewComparer()

	source := &models.Schema{Name: "a", DatabaseType: models.PostgreSQL, Tables: []models.Table{
		tableWithConstraints([]models.Constraint{
			{Name: "orders_total_check", Type: models.Check, CheckExpression: "total >= 0"},
		}),
	}}
	target := &models.Schema{Name: "b", DatabaseType: models.PostgreSQL, Tables: []models.Table{
		tableWithConstraints([]models.Constraint{
			{Name: "orders_total_check", Type: models.Check, CheckExpression: "total > 0"},
		}),
	}}

	result := comparer.Compare(source, target)
	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Modified, result.Differences[0].Type)
	assert.Equal(t, "Constraint", result.Differences[0].ObjectType)
}

// FKs are matched structurally (system names ignored), but a real change such as
// a different ON DELETE rule must still surface as a single MODIFIED.
func TestComparer_Compare_ModifiedForeignKeyStaysSingle(t *testing.T) {
	comparer := NewComparer()

	fk := func(name, onDelete string) models.Constraint {
		return models.Constraint{
			Name: name, Type: models.ForeignKey, Columns: []string{"cust_id"},
			ReferencedTable: "customers", ReferencedColumn: []string{"id"}, OnDelete: onDelete,
		}
	}
	source := &models.Schema{Name: "a", DatabaseType: models.Oracle, Tables: []models.Table{
		tableWithConstraints([]models.Constraint{fk("SYS_C001", "NO ACTION")}),
	}}
	target := &models.Schema{Name: "b", DatabaseType: models.Oracle, Tables: []models.Table{
		tableWithConstraints([]models.Constraint{fk("SYS_C999", "CASCADE")}),
	}}

	result := comparer.Compare(source, target)
	assert.Equal(t, 1, len(result.Differences))
	assert.Equal(t, models.Modified, result.Differences[0].Type)
	assert.Equal(t, "Constraint", result.Differences[0].ObjectType)
}
