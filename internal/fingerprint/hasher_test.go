package fingerprint

import (
	"testing"

	"github.com/nechja/schemalyzer/pkg/models"
)

func TestGenerateFingerprint(t *testing.T) {
	hasher := NewHasher()
	
	schema1 := &models.Schema{
		Name: "test",
		DatabaseType: models.PostgreSQL,
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", IsNullable: false, IsPrimaryKey: true},
					{Name: "name", DataType: "varchar(100)", IsNullable: false},
					{Name: "email", DataType: "varchar(255)", IsNullable: false, IsUnique: true},
				},
				Constraints: []models.Constraint{
					{Name: "pk_users", Type: models.PrimaryKey, Columns: []string{"id"}},
					{Name: "uk_users_email", Type: models.Unique, Columns: []string{"email"}},
				},
			},
		},
	}
	
	hash1, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}
	
	if len(hash1) != 64 {
		t.Errorf("Expected SHA256 hash length of 64, got %d", len(hash1))
	}
	
	hash2, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate second fingerprint: %v", err)
	}
	
	if hash1 != hash2 {
		t.Error("Same schema should produce same fingerprint")
	}
}

func TestFingerprintDifference(t *testing.T) {
	hasher := NewHasher()
	
	schema1 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "integer"},
					{Name: "name", DataType: "varchar(100)"},
				},
			},
		},
	}
	
	schema2 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "integer"},
					{Name: "name", DataType: "varchar(200)"}, // Different size
				},
			},
		},
	}
	
	hash1, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema1: %v", err)
	}
	
	hash2, err := hasher.GenerateFingerprint(schema2)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema2: %v", err)
	}
	
	if hash1 == hash2 {
		t.Error("Different schemas should produce different fingerprints")
	}
}

func TestFingerprintIgnoresComments(t *testing.T) {
	hasher := NewHasher()
	
	schema1 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", Comment: "Primary key"},
				},
				Comment: "User table",
			},
		},
	}
	
	schema2 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "integer", Comment: "Different comment"},
				},
				Comment: "Different table comment",
			},
		},
	}
	
	hash1, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema1: %v", err)
	}
	
	hash2, err := hasher.GenerateFingerprint(schema2)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema2: %v", err)
	}
	
	if hash1 != hash2 {
		t.Error("Schemas with different comments only should produce same fingerprint")
	}
}

func TestFingerprintOrderIndependence(t *testing.T) {
	hasher := NewHasher()
	
	schema1 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "id", DataType: "integer"},
					{Name: "name", DataType: "varchar"},
				},
			},
			{
				Name: "posts",
				Columns: []models.Column{
					{Name: "id", DataType: "integer"},
					{Name: "title", DataType: "varchar"},
				},
			},
		},
	}
	
	schema2 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "posts", // Different order
				Columns: []models.Column{
					{Name: "title", DataType: "varchar"}, // Different order
					{Name: "id", DataType: "integer"},
				},
			},
			{
				Name: "users",
				Columns: []models.Column{
					{Name: "name", DataType: "varchar"}, // Different order
					{Name: "id", DataType: "integer"},
				},
			},
		},
	}
	
	hash1, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema1: %v", err)
	}
	
	hash2, err := hasher.GenerateFingerprint(schema2)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema2: %v", err)
	}
	
	if hash1 != hash2 {
		t.Error("Schemas with different ordering should produce same fingerprint")
	}
}

func TestFingerprintWithAllObjectTypes(t *testing.T) {
	hasher := NewHasher()
	
	schema := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{Name: "users", Columns: []models.Column{{Name: "id", DataType: "int"}}},
		},
		Views: []models.View{
			{Name: "user_view", Definition: "SELECT * FROM users"},
		},
		Indexes: []models.Index{
			{Name: "idx_users_id", TableName: "users", Columns: []string{"id"}},
		},
		Sequences: []models.Sequence{
			{Name: "user_seq", StartValue: 1, Increment: 1},
		},
		Procedures: []models.Procedure{
			{Name: "get_user", Body: "BEGIN SELECT * FROM users; END"},
		},
		Functions: []models.Function{
			{Name: "count_users", ReturnType: "integer", Body: "RETURN COUNT(*) FROM users"},
		},
		Triggers: []models.Trigger{
			{Name: "user_audit", TableName: "users", Event: models.Insert, Timing: models.After},
		},
	}
	
	hash, err := hasher.GenerateFingerprint(schema)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}
	
	if len(hash) != 64 {
		t.Errorf("Expected SHA256 hash length of 64, got %d", len(hash))
	}
}

func TestFingerprintIndexColumnOrderIndependence(t *testing.T) {
	hasher := NewHasher()
	
	schema1 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Indexes: []models.Index{
					{
						Name:     "idx_user_data",
						Columns:  []string{"email", "username", "created_at"},
						IsUnique: true,
					},
				},
			},
		},
		Indexes: []models.Index{
			{
				Name:      "global_idx",
				TableName: "users", 
				Columns:   []string{"id", "status", "type"},
				IsUnique:  false,
			},
		},
	}
	
	schema2 := &models.Schema{
		Name: "test",
		Tables: []models.Table{
			{
				Name: "users",
				Indexes: []models.Index{
					{
						Name:     "idx_user_data",
						Columns:  []string{"created_at", "email", "username"}, // Different order
						IsUnique: true,
					},
				},
			},
		},
		Indexes: []models.Index{
			{
				Name:      "global_idx",
				TableName: "users",
				Columns:   []string{"type", "id", "status"}, // Different order
				IsUnique:  false,
			},
		},
	}
	
	hash1, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema1: %v", err)
	}
	
	hash2, err := hasher.GenerateFingerprint(schema2)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema2: %v", err)
	}
	
	if hash1 != hash2 {
		t.Error("Indexes with different column order should produce same fingerprint")
	}
}

func TestFingerprintParameterOrderIndependence(t *testing.T) {
	hasher := NewHasher()
	
	schema1 := &models.Schema{
		Name: "test",
		Functions: []models.Function{
			{
				Name: "test_func",
				Parameters: []models.Parameter{
					{Name: "param_a", DataType: "integer", Direction: models.In},
					{Name: "param_b", DataType: "varchar", Direction: models.In},
					{Name: "param_c", DataType: "boolean", Direction: models.Out},
				},
				ReturnType: "integer",
				Body: "BEGIN RETURN 1; END",
			},
		},
	}
	
	schema2 := &models.Schema{
		Name: "test",
		Functions: []models.Function{
			{
				Name: "test_func",
				Parameters: []models.Parameter{
					{Name: "param_c", DataType: "boolean", Direction: models.Out}, // Different order
					{Name: "param_a", DataType: "integer", Direction: models.In},
					{Name: "param_b", DataType: "varchar", Direction: models.In},
				},
				ReturnType: "integer",
				Body: "BEGIN RETURN 1; END",
			},
		},
	}
	
	hash1, err := hasher.GenerateFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema1: %v", err)
	}
	
	hash2, err := hasher.GenerateFingerprint(schema2)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint for schema2: %v", err)
	}
	
	if hash1 != hash2 {
		t.Error("Functions with different parameter order should produce same fingerprint")
	}
}