package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/nechja/schemalyzer/pkg/models"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestFormatter_FormatJSON(t *testing.T) {
	formatter := NewFormatter(FormatJSON)

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
		Differences: []models.Difference{
			{
				Type:        models.Added,
				ObjectType:  "Table",
				ObjectName:  "users",
				Description: "Table exists in target but not in source",
			},
		},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	var jsonOutput map[string]interface{}
	err = json.Unmarshal(output, &jsonOutput)
	assert.NoError(t, err)

	assert.Equal(t, "postgresql://test", jsonOutput["source_database"])
	assert.Equal(t, "mysql://test", jsonOutput["target_database"])
	assert.Equal(t, float64(1), jsonOutput["total_differences"])

	summary := jsonOutput["summary"].(map[string]interface{})
	assert.Equal(t, float64(1), summary["added"])
}

func TestFormatter_FormatYAML(t *testing.T) {
	formatter := NewFormatter(FormatYAML)

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
		Differences: []models.Difference{
			{
				Type:        models.Removed,
				ObjectType:  "Column",
				ObjectName:  "users.email",
				Description: "Column removed from table",
			},
		},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	var yamlOutput map[string]interface{}
	err = yaml.Unmarshal(output, &yamlOutput)
	assert.NoError(t, err)

	assert.Equal(t, "postgresql://test", yamlOutput["source_database"])
	assert.Equal(t, "mysql://test", yamlOutput["target_database"])
	assert.Equal(t, 1, yamlOutput["total_differences"])
}

func TestFormatter_FormatText(t *testing.T) {
	formatter := NewFormatter(FormatText)

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
		Differences: []models.Difference{
			{
				Type:        models.Added,
				ObjectType:  "Table",
				ObjectName:  "users",
				Description: "Table exists in target but not in source",
			},
			{
				Type:        models.Removed,
				ObjectType:  "View",
				ObjectName:  "user_summary",
				Description: "View exists in source but not in target",
			},
			{
				Type:        models.Modified,
				ObjectType:  "Column",
				ObjectName:  "products.price",
				Description: "Column definition changed",
			},
			{
				Type:        models.Modified,
				ObjectType:  "Constraint",
				ObjectName:  "orders.orders_user_id_fkey",
				Description: "Constraint definition changed",
				Source: &models.Constraint{
					Type:     models.ForeignKey,
					OnUpdate: "NO ACTION",
					OnDelete: "CASCADE",
				},
				Target: &models.Constraint{
					Type:     models.ForeignKey,
					OnUpdate: "CASCADE",
					OnDelete: "SET NULL",
				},
			},
		},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	text := string(output)
	assert.Contains(t, text, "Schema Comparison Report")
	assert.Contains(t, text, "Total Differences: 4")
	assert.Contains(t, text, "Added Objects")
	assert.Contains(t, text, "+ Table: users")
	assert.Contains(t, text, "Removed Objects")
	assert.Contains(t, text, "- View: user_summary")
	assert.Contains(t, text, "Modified Objects")
	assert.Contains(t, text, "~ Column: products.price")
	assert.Contains(t, text, "FK Actions: OnUpdate NO ACTION -> CASCADE, OnDelete CASCADE -> SET NULL")
}

func TestFormatter_FormatText_NoDifferences(t *testing.T) {
	formatter := NewFormatter(FormatText)

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Now(),
		Differences:    []models.Difference{},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	text := string(output)
	assert.Contains(t, text, "Total Differences: 0")
	assert.Contains(t, text, "No differences found.")
}

func TestFormatter_FormatSummary(t *testing.T) {
	formatter := NewFormatter(FormatSummary)

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
		Differences: []models.Difference{
			{Type: models.Added, ObjectType: "Table", ObjectName: "users"},
			{Type: models.Added, ObjectType: "Index", ObjectName: "idx_users_email"},
			{Type: models.Removed, ObjectType: "View", ObjectName: "user_summary"},
			{Type: models.Modified, ObjectType: "Column", ObjectName: "products.price"},
			{Type: models.Modified, ObjectType: "Constraint", ObjectName: "fk_orders_users"},
		},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	text := string(output)
	assert.Contains(t, text, "Schema Comparison Summary")
	assert.Contains(t, text, "Source: postgresql://test")
	assert.Contains(t, text, "Target: mysql://test")
	assert.Contains(t, text, "added: 2")
	assert.Contains(t, text, "removed: 1")
	assert.Contains(t, text, "modified: 2")
	assert.Contains(t, text, "Total Differences: 5")
}

func TestFormatter_FormatSummary_NoDifferences(t *testing.T) {
	formatter := NewFormatter(FormatSummary)

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Now(),
		Differences:    []models.Difference{},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	text := string(output)
	assert.Contains(t, text, "Result: SCHEMAS ARE IDENTICAL")
}

func TestFormatter_UnsupportedFormat(t *testing.T) {
	formatter := NewFormatter("invalid")

	result := &models.ComparisonResult{}

	_, err := formatter.Format(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestFormatter_GenerateSummary(t *testing.T) {
	formatter := NewFormatter(FormatJSON)

	result := &models.ComparisonResult{
		Differences: []models.Difference{
			{Type: models.Added},
			{Type: models.Added},
			{Type: models.Removed},
			{Type: models.Modified},
			{Type: models.Modified},
			{Type: models.Modified},
		},
	}

	summary := formatter.generateSummary(result)

	assert.Equal(t, 2, summary["added"])
	assert.Equal(t, 1, summary["removed"])
	assert.Equal(t, 3, summary["modified"])
}

func TestFormatter_JSONStructure(t *testing.T) {
	formatter := NewFormatter(FormatJSON)

	table := models.Table{
		Name:   "users",
		Schema: "public",
		Columns: []models.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
	}

	result := &models.ComparisonResult{
		SourceDatabase: "postgresql://test",
		TargetDatabase: "mysql://test",
		ComparisonTime: time.Now(),
		Differences: []models.Difference{
			{
				Type:        models.Added,
				ObjectType:  "Table",
				ObjectName:  "users",
				Target:      table,
				Description: "Table added",
			},
		},
	}

	output, err := formatter.Format(result)
	assert.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(output, &parsed)
	assert.NoError(t, err)

	// Check structure
	assert.Contains(t, parsed, "source_database")
	assert.Contains(t, parsed, "target_database")
	assert.Contains(t, parsed, "comparison_time")
	assert.Contains(t, parsed, "total_differences")
	assert.Contains(t, parsed, "summary")
	assert.Contains(t, parsed, "differences")

	// Verify indentation (should be pretty-printed)
	assert.True(t, strings.Contains(string(output), "\n  "))
}
