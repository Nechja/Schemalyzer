package output

import (
	"encoding/json"
	"fmt"
	"github.com/nechja/schemalyzer/pkg/models"
	"gopkg.in/yaml.v3"
	"strings"
)

type OutputFormat string

const (
	FormatJSON    OutputFormat = "json"
	FormatYAML    OutputFormat = "yaml"
	FormatText    OutputFormat = "text"
	FormatSummary OutputFormat = "summary"
)

type Formatter struct {
	format OutputFormat
}

func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

func (f *Formatter) Format(result *models.ComparisonResult) ([]byte, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(result)
	case FormatYAML:
		return f.formatYAML(result)
	case FormatText:
		return f.formatText(result)
	case FormatSummary:
		return f.formatSummary(result)
	default:
		return nil, fmt.Errorf("unsupported format: %s", f.format)
	}
}

func (f *Formatter) formatJSON(result *models.ComparisonResult) ([]byte, error) {
	output := struct {
		SourceDatabase   string              `json:"source_database"`
		TargetDatabase   string              `json:"target_database"`
		ComparisonTime   string              `json:"comparison_time"`
		TotalDifferences int                 `json:"total_differences"`
		Summary          map[string]int      `json:"summary"`
		Differences      []models.Difference `json:"differences"`
	}{
		SourceDatabase:   result.SourceDatabase,
		TargetDatabase:   result.TargetDatabase,
		ComparisonTime:   result.ComparisonTime.Format("2006-01-02T15:04:05Z"),
		TotalDifferences: len(result.Differences),
		Summary:          f.generateSummary(result),
		Differences:      result.Differences,
	}

	return json.MarshalIndent(output, "", "  ")
}

func (f *Formatter) formatYAML(result *models.ComparisonResult) ([]byte, error) {
	output := struct {
		SourceDatabase   string              `yaml:"source_database"`
		TargetDatabase   string              `yaml:"target_database"`
		ComparisonTime   string              `yaml:"comparison_time"`
		TotalDifferences int                 `yaml:"total_differences"`
		Summary          map[string]int      `yaml:"summary"`
		Differences      []models.Difference `yaml:"differences"`
	}{
		SourceDatabase:   result.SourceDatabase,
		TargetDatabase:   result.TargetDatabase,
		ComparisonTime:   result.ComparisonTime.Format("2006-01-02T15:04:05Z"),
		TotalDifferences: len(result.Differences),
		Summary:          f.generateSummary(result),
		Differences:      result.Differences,
	}

	return yaml.Marshal(output)
}

func (f *Formatter) formatText(result *models.ComparisonResult) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString("Schema Comparison Report\n")
	sb.WriteString("========================\n\n")
	sb.WriteString(fmt.Sprintf("Source Database: %s\n", result.SourceDatabase))
	sb.WriteString(fmt.Sprintf("Target Database: %s\n", result.TargetDatabase))
	sb.WriteString(fmt.Sprintf("Comparison Time: %s\n", result.ComparisonTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Total Differences: %d\n\n", len(result.Differences)))

	if len(result.Differences) == 0 {
		sb.WriteString("No differences found.\n")
		return []byte(sb.String()), nil
	}

	// Group differences by type
	added := []models.Difference{}
	removed := []models.Difference{}
	modified := []models.Difference{}

	for _, diff := range result.Differences {
		switch diff.Type {
		case models.Added:
			added = append(added, diff)
		case models.Removed:
			removed = append(removed, diff)
		case models.Modified:
			modified = append(modified, diff)
		}
	}

	// Write added objects
	if len(added) > 0 {
		sb.WriteString("Added Objects\n")
		sb.WriteString("-------------\n")
		for _, diff := range added {
			sb.WriteString(fmt.Sprintf("+ %s: %s\n", diff.ObjectType, diff.ObjectName))
			sb.WriteString(fmt.Sprintf("  %s\n", diff.Description))
			if extra := formatConstraintActions(diff); extra != "" {
				sb.WriteString(extra)
			}
		}
		sb.WriteString("\n")
	}

	// Write removed objects
	if len(removed) > 0 {
		sb.WriteString("Removed Objects\n")
		sb.WriteString("---------------\n")
		for _, diff := range removed {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", diff.ObjectType, diff.ObjectName))
			sb.WriteString(fmt.Sprintf("  %s\n", diff.Description))
			if extra := formatConstraintActions(diff); extra != "" {
				sb.WriteString(extra)
			}
		}
		sb.WriteString("\n")
	}

	// Write modified objects
	if len(modified) > 0 {
		sb.WriteString("Modified Objects\n")
		sb.WriteString("----------------\n")
		for _, diff := range modified {
			sb.WriteString(fmt.Sprintf("~ %s: %s\n", diff.ObjectType, diff.ObjectName))
			sb.WriteString(fmt.Sprintf("  %s\n", diff.Description))
			if extra := formatConstraintActions(diff); extra != "" {
				sb.WriteString(extra)
			}
		}
		sb.WriteString("\n")
	}

	return []byte(sb.String()), nil
}

func (f *Formatter) formatSummary(result *models.ComparisonResult) ([]byte, error) {
	summary := f.generateSummary(result)

	var sb strings.Builder
	sb.WriteString("Schema Comparison Summary\n")
	sb.WriteString("=========================\n\n")
	sb.WriteString(fmt.Sprintf("Source: %s\n", result.SourceDatabase))
	sb.WriteString(fmt.Sprintf("Target: %s\n", result.TargetDatabase))
	sb.WriteString(fmt.Sprintf("Time: %s\n\n", result.ComparisonTime.Format("2006-01-02 15:04:05")))

	if len(result.Differences) == 0 {
		sb.WriteString("Result: SCHEMAS ARE IDENTICAL\n")
		return []byte(sb.String()), nil
	}

	sb.WriteString("Differences by Type:\n")
	sb.WriteString("-------------------\n")
	for diffType, count := range summary {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", diffType, count))
	}

	sb.WriteString("\nDifferences by Object:\n")
	sb.WriteString("---------------------\n")

	objectCounts := make(map[string]int)
	for _, diff := range result.Differences {
		key := fmt.Sprintf("%s %s", diff.Type, diff.ObjectType)
		objectCounts[key]++
	}

	for objType, count := range objectCounts {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", objType, count))
	}

	sb.WriteString(fmt.Sprintf("\nTotal Differences: %d\n", len(result.Differences)))

	return []byte(sb.String()), nil
}

func (f *Formatter) generateSummary(result *models.ComparisonResult) map[string]int {
	summary := make(map[string]int)

	for _, diff := range result.Differences {
		switch diff.Type {
		case models.Added:
			summary["added"]++
		case models.Removed:
			summary["removed"]++
		case models.Modified:
			summary["modified"]++
		}
	}

	return summary
}

func formatConstraintActions(diff models.Difference) string {
	source := constraintFromInterface(diff.Source)
	target := constraintFromInterface(diff.Target)
	if !isForeignKeyConstraint(source) && !isForeignKeyConstraint(target) {
		return ""
	}

	normalize := func(val string) string {
		return strings.ToUpper(strings.TrimSpace(val))
	}
	display := func(val string) string {
		if trimmed := normalize(val); trimmed != "" {
			return trimmed
		}
		return "NOT SPECIFIED"
	}

	sb := strings.Builder{}
	sb.WriteString("  FK Actions: ")

	if source != nil && target != nil && (normalize(source.OnUpdate) != normalize(target.OnUpdate) || normalize(source.OnDelete) != normalize(target.OnDelete)) {
		sb.WriteString(fmt.Sprintf("OnUpdate %s -> %s, OnDelete %s -> %s\n",
			display(source.OnUpdate),
			display(target.OnUpdate),
			display(source.OnDelete),
			display(target.OnDelete)))
		return sb.String()
	}

	active := target
	if active == nil {
		active = source
	}
	sb.WriteString(fmt.Sprintf("OnUpdate %s, OnDelete %s\n",
		display(active.OnUpdate),
		display(active.OnDelete)))
	return sb.String()
}

func constraintFromInterface(val interface{}) *models.Constraint {
	switch v := val.(type) {
	case *models.Constraint:
		return v
	case models.Constraint:
		return &v
	default:
		return nil
	}
}

func isForeignKeyConstraint(c *models.Constraint) bool {
	return c != nil && c.Type == models.ForeignKey
}
