package docs

import (
	"fmt"
	"strings"
	"github.com/nechja/schemalyzer/pkg/models"
)

type DocumentGenerator interface {
	Generate(schema *models.Schema) (string, error)
}

type PlantUMLGenerator struct{}

func NewPlantUMLGenerator() *PlantUMLGenerator {
	return &PlantUMLGenerator{}
}

func (g *PlantUMLGenerator) Generate(schema *models.Schema) (string, error) {
	var sb strings.Builder
	
	// PlantUML header
	sb.WriteString("@startuml\n")
	sb.WriteString("!theme plain\n")
	sb.WriteString("skinparam linetype ortho\n")
	sb.WriteString("skinparam classAttributeIconSize 0\n")
	sb.WriteString("skinparam classFontSize 12\n")
	sb.WriteString("skinparam classFontName Arial\n\n")
	
	sb.WriteString(fmt.Sprintf("title %s Database Schema\n\n", schema.Name))
	
	// Generate entities for tables
	for _, table := range schema.Tables {
		g.generateTable(&sb, table)
	}
	
	// Generate relationships from foreign keys
	for _, table := range schema.Tables {
		for _, constraint := range table.Constraints {
			if constraint.Type == models.ForeignKey {
				g.generateRelationship(&sb, table.Name, constraint)
			}
		}
	}
	
	sb.WriteString("\n@enduml\n")
	
	return sb.String(), nil
}

func (g *PlantUMLGenerator) generateTable(sb *strings.Builder, table models.Table) {
	sb.WriteString(fmt.Sprintf("entity \"%s\" as %s {\n", table.Name, sanitizeName(table.Name)))
	
	// Primary key columns first
	for _, col := range table.Columns {
		if col.IsPrimaryKey {
			sb.WriteString(fmt.Sprintf("  * **%s** : %s <<PK>>\n", col.Name, col.DataType))
		}
	}
	
	sb.WriteString("  --\n")
	
	// Other columns
	for _, col := range table.Columns {
		if !col.IsPrimaryKey {
			nullable := ""
			if col.IsNullable {
				nullable = " ?"
			}
			
			unique := ""
			if col.IsUnique {
				unique = " <<unique>>"
			}
			
			// Check if it's a foreign key
			fk := ""
			for _, constraint := range table.Constraints {
				if constraint.Type == models.ForeignKey {
					for _, fkCol := range constraint.Columns {
						if fkCol == col.Name {
							fk = " <<FK>>"
							break
						}
					}
				}
			}
			
			sb.WriteString(fmt.Sprintf("  %s : %s%s%s%s\n", col.Name, col.DataType, nullable, unique, fk))
		}
	}
	
	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateRelationship(sb *strings.Builder, tableName string, constraint models.Constraint) {
	// Determine cardinality based on constraint
	cardinality := "}o--||"  // Many to one (most common for FK)
	
	sb.WriteString(fmt.Sprintf("%s %s %s : %s\n", 
		sanitizeName(tableName), 
		cardinality, 
		sanitizeName(constraint.ReferencedTable),
		constraint.Name))
}

func sanitizeName(name string) string {
	// PlantUML doesn't like certain characters in entity names
	return strings.ReplaceAll(name, " ", "_")
}

// MermaidGenerator generates Mermaid ER diagrams
type MermaidGenerator struct{}

func NewMermaidGenerator() *MermaidGenerator {
	return &MermaidGenerator{}
}

func (g *MermaidGenerator) Generate(schema *models.Schema) (string, error) {
	var sb strings.Builder
	
	sb.WriteString("```mermaid\nerDiagram\n")
	
	// Generate entities
	for _, table := range schema.Tables {
		g.generateMermaidTable(&sb, table)
	}
	
	// Generate relationships
	for _, table := range schema.Tables {
		for _, constraint := range table.Constraints {
			if constraint.Type == models.ForeignKey {
				g.generateMermaidRelationship(&sb, table.Name, constraint)
			}
		}
	}
	
	sb.WriteString("```\n")
	
	return sb.String(), nil
}

func (g *MermaidGenerator) generateMermaidTable(sb *strings.Builder, table models.Table) {
	sb.WriteString(fmt.Sprintf("    %s {\n", sanitizeMermaidName(table.Name)))
	
	for _, col := range table.Columns {
		dataType := simplifyDataType(col.DataType)
		keys := []string{}
		
		if col.IsPrimaryKey {
			keys = append(keys, "PK")
		}
		
		// Check if foreign key
		for _, constraint := range table.Constraints {
			if constraint.Type == models.ForeignKey {
				for _, fkCol := range constraint.Columns {
					if fkCol == col.Name {
						keys = append(keys, "FK")
						break
					}
				}
			}
		}
		
		keyStr := ""
		if len(keys) > 0 {
			keyStr = " \"" + strings.Join(keys, ",") + "\""
		}
		
		sb.WriteString(fmt.Sprintf("        %s %s%s\n", dataType, sanitizeMermaidName(col.Name), keyStr))
	}
	
	sb.WriteString("    }\n")
}

func (g *MermaidGenerator) generateMermaidRelationship(sb *strings.Builder, tableName string, constraint models.Constraint) {
	// Mermaid relationship syntax: ||--o{ (one to many), ||--|| (one to one), etc.
	sb.WriteString(fmt.Sprintf("    %s ||--o{ %s : \"%s\"\n",
		sanitizeMermaidName(constraint.ReferencedTable),
		sanitizeMermaidName(tableName),
		"has"))
}

func sanitizeMermaidName(name string) string {
	// Mermaid is more forgiving but still sanitize
	return strings.ReplaceAll(name, " ", "_")
}

func simplifyDataType(dataType string) string {
	// Simplify data types for diagram clarity
	if strings.HasPrefix(strings.ToUpper(dataType), "VARCHAR") {
		return "string"
	}
	if strings.HasPrefix(strings.ToUpper(dataType), "NUMBER") || strings.HasPrefix(strings.ToUpper(dataType), "INT") {
		return "int"
	}
	if strings.HasPrefix(strings.ToUpper(dataType), "DATE") || strings.HasPrefix(strings.ToUpper(dataType), "TIMESTAMP") {
		return "date"
	}
	if strings.HasPrefix(strings.ToUpper(dataType), "BOOL") {
		return "bool"
	}
	return strings.ToLower(dataType)
}

// MarkdownDocGenerator generates comprehensive markdown documentation
type MarkdownDocGenerator struct{}

func NewMarkdownDocGenerator() *MarkdownDocGenerator {
	return &MarkdownDocGenerator{}
}

func (g *MarkdownDocGenerator) Generate(schema *models.Schema) (string, error) {
	var sb strings.Builder
	
	// Header
	sb.WriteString(fmt.Sprintf("# %s Database Schema Documentation\n\n", schema.Name))
	sb.WriteString(fmt.Sprintf("**Database Type**: %s\n\n", schema.DatabaseType))
	
	// Table of Contents
	sb.WriteString("## Table of Contents\n\n")
	sb.WriteString("- [Tables](#tables)\n")
	if len(schema.Views) > 0 {
		sb.WriteString("- [Views](#views)\n")
	}
	if len(schema.Sequences) > 0 {
		sb.WriteString("- [Sequences](#sequences)\n")
	}
	if len(schema.Functions) > 0 {
		sb.WriteString("- [Functions](#functions)\n")
	}
	if len(schema.Procedures) > 0 {
		sb.WriteString("- [Procedures](#procedures)\n")
	}
	if len(schema.Triggers) > 0 {
		sb.WriteString("- [Triggers](#triggers)\n")
	}
	sb.WriteString("\n")
	
	// Tables section
	sb.WriteString("## Tables\n\n")
	for _, table := range schema.Tables {
		g.generateMarkdownTable(&sb, table)
	}
	
	// Views section
	if len(schema.Views) > 0 {
		sb.WriteString("## Views\n\n")
		for _, view := range schema.Views {
			g.generateMarkdownView(&sb, view)
		}
	}
	
	// Sequences section
	if len(schema.Sequences) > 0 {
		sb.WriteString("## Sequences\n\n")
		for _, seq := range schema.Sequences {
			g.generateMarkdownSequence(&sb, seq)
		}
	}
	
	// Functions section
	if len(schema.Functions) > 0 {
		sb.WriteString("## Functions\n\n")
		for _, fn := range schema.Functions {
			sb.WriteString(fmt.Sprintf("### %s\n\n", fn.Name))
			sb.WriteString("```sql\n")
			sb.WriteString(fn.Body)
			sb.WriteString("\n```\n\n")
		}
	}
	
	// Procedures section
	if len(schema.Procedures) > 0 {
		sb.WriteString("## Procedures\n\n")
		for _, proc := range schema.Procedures {
			sb.WriteString(fmt.Sprintf("### %s\n\n", proc.Name))
			sb.WriteString("```sql\n")
			sb.WriteString(proc.Body)
			sb.WriteString("\n```\n\n")
		}
	}
	
	// Triggers section
	if len(schema.Triggers) > 0 {
		sb.WriteString("## Triggers\n\n")
		for _, trigger := range schema.Triggers {
			g.generateMarkdownTrigger(&sb, trigger)
		}
	}
	
	return sb.String(), nil
}

func (g *MarkdownDocGenerator) generateMarkdownTable(sb *strings.Builder, table models.Table) {
	sb.WriteString(fmt.Sprintf("### %s\n\n", table.Name))
	
	if table.Comment != "" {
		sb.WriteString(fmt.Sprintf("_%s_\n\n", table.Comment))
	}
	
	// Columns table
	sb.WriteString("**Columns:**\n\n")
	sb.WriteString("| Column | Type | Nullable | Default | Keys | Description |\n")
	sb.WriteString("|--------|------|----------|---------|------|-------------|\n")
	
	for _, col := range table.Columns {
		nullable := "No"
		if col.IsNullable {
			nullable = "Yes"
		}
		
		defaultVal := "-"
		if col.DefaultValue != nil {
			defaultVal = *col.DefaultValue
		}
		
		keys := []string{}
		if col.IsPrimaryKey {
			keys = append(keys, "PK")
		}
		if col.IsUnique {
			keys = append(keys, "UNIQUE")
		}
		
		// Check if foreign key
		for _, constraint := range table.Constraints {
			if constraint.Type == models.ForeignKey {
				for _, fkCol := range constraint.Columns {
					if fkCol == col.Name {
						keys = append(keys, "FK")
						break
					}
				}
			}
		}
		
		keyStr := strings.Join(keys, ", ")
		if keyStr == "" {
			keyStr = "-"
		}
		
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			col.Name, col.DataType, nullable, defaultVal, keyStr, col.Comment))
	}
	
	// Constraints section
	if len(table.Constraints) > 0 {
		sb.WriteString("\n**Constraints:**\n\n")
		for _, constraint := range table.Constraints {
			switch constraint.Type {
			case models.PrimaryKey:
				sb.WriteString(fmt.Sprintf("- **Primary Key** (%s): %s\n", constraint.Name, strings.Join(constraint.Columns, ", ")))
			case models.ForeignKey:
				sb.WriteString(fmt.Sprintf("- **Foreign Key** (%s): %s → %s.%s\n", 
					constraint.Name, 
					strings.Join(constraint.Columns, ", "),
					constraint.ReferencedTable,
					strings.Join(constraint.ReferencedColumn, ", ")))
			case models.Unique:
				sb.WriteString(fmt.Sprintf("- **Unique** (%s): %s\n", constraint.Name, strings.Join(constraint.Columns, ", ")))
			case models.Check:
				sb.WriteString(fmt.Sprintf("- **Check** (%s): %s\n", constraint.Name, constraint.CheckExpression))
			}
		}
	}
	
	// Indexes section
	if len(table.Indexes) > 0 {
		sb.WriteString("\n**Indexes:**\n\n")
		for _, index := range table.Indexes {
			indexType := "INDEX"
			if index.IsUnique {
				indexType = "UNIQUE INDEX"
			}
			sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", indexType, index.Name, strings.Join(index.Columns, ", ")))
		}
	}
	
	sb.WriteString("\n")
}

func (g *MarkdownDocGenerator) generateMarkdownView(sb *strings.Builder, view models.View) {
	sb.WriteString(fmt.Sprintf("### %s\n\n", view.Name))
	
	sb.WriteString("**Columns:**\n\n")
	sb.WriteString("| Column | Type | Nullable |\n")
	sb.WriteString("|--------|------|----------|\n")
	
	for _, col := range view.Columns {
		nullable := "No"
		if col.IsNullable {
			nullable = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", col.Name, col.DataType, nullable))
	}
	
	sb.WriteString("\n**Definition:**\n\n")
	sb.WriteString("```sql\n")
	sb.WriteString(view.Definition)
	sb.WriteString("\n```\n\n")
}

func (g *MarkdownDocGenerator) generateMarkdownSequence(sb *strings.Builder, seq models.Sequence) {
	sb.WriteString(fmt.Sprintf("### %s\n\n", seq.Name))
	sb.WriteString(fmt.Sprintf("- **Start Value**: %d\n", seq.StartValue))
	sb.WriteString(fmt.Sprintf("- **Increment**: %d\n", seq.Increment))
	sb.WriteString(fmt.Sprintf("- **Min Value**: %d\n", seq.MinValue))
	sb.WriteString(fmt.Sprintf("- **Max Value**: %d\n", seq.MaxValue))
	sb.WriteString(fmt.Sprintf("- **Cyclic**: %v\n", seq.IsCyclic))
	sb.WriteString(fmt.Sprintf("- **Current Value**: %d\n\n", seq.CurrentValue))
}

func (g *MarkdownDocGenerator) generateMarkdownTrigger(sb *strings.Builder, trigger models.Trigger) {
	sb.WriteString(fmt.Sprintf("### %s\n\n", trigger.Name))
	sb.WriteString(fmt.Sprintf("- **Table**: %s\n", trigger.TableName))
	sb.WriteString(fmt.Sprintf("- **Event**: %s\n", trigger.Event))
	sb.WriteString(fmt.Sprintf("- **Timing**: %s\n", trigger.Timing))
	sb.WriteString("\n**Body:**\n\n")
	sb.WriteString("```sql\n")
	sb.WriteString(trigger.Body)
	sb.WriteString("\n```\n\n")
}