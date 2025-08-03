package docs

import (
	"fmt"
	"strings"
	"github.com/nechja/schemalyzer/pkg/models"
)

// GraphVizGenerator generates DOT format for GraphViz
type GraphVizGenerator struct {
	// Options for rendering
	RankDir        string // TB (top-bottom), LR (left-right), BT, RL
	NodeShape      string // record, box, ellipse, etc.
	ShowDataTypes  bool
	ShowNullable   bool
	ColorScheme    string // blues, greens, accent, etc.
}

func NewGraphVizGenerator() *GraphVizGenerator {
	return &GraphVizGenerator{
		RankDir:       "TB",
		NodeShape:     "record",
		ShowDataTypes: true,
		ShowNullable:  true,
		ColorScheme:   "blues",
	}
}

func (g *GraphVizGenerator) Generate(schema *models.Schema) (string, error) {
	var sb strings.Builder
	
	// DOT header
	sb.WriteString("digraph ERD {\n")
	sb.WriteString(fmt.Sprintf("  graph [rankdir=%s, bgcolor=white, splines=true, overlap=false];\n", g.RankDir))
	sb.WriteString("  node [shape=record, fontname=\"Arial\", fontsize=11];\n")
	sb.WriteString("  edge [fontname=\"Arial\", fontsize=10];\n\n")
	
	// Title
	sb.WriteString("  labelloc=\"t\";\n")
	sb.WriteString(fmt.Sprintf("  label=\"%s Database Schema\";\n", schema.Name))
	sb.WriteString("  fontsize=16;\n\n")
	
	// Color definitions based on table types/purposes
	colors := g.getColorScheme()
	
	// Generate nodes for tables
	for i, table := range schema.Tables {
		color := colors[i%len(colors)]
		g.generateGraphVizTable(&sb, table, color)
	}
	
	// Generate edges for relationships
	for _, table := range schema.Tables {
		for _, constraint := range table.Constraints {
			if constraint.Type == models.ForeignKey {
				g.generateGraphVizRelationship(&sb, table.Name, constraint)
			}
		}
	}
	
	sb.WriteString("}\n")
	
	return sb.String(), nil
}

func (g *GraphVizGenerator) generateGraphVizTable(sb *strings.Builder, table models.Table, color string) {
	tableName := sanitizeGraphVizName(table.Name)
	
	// Start table node
	sb.WriteString(fmt.Sprintf("  %s [\n", tableName))
	sb.WriteString(fmt.Sprintf("    fillcolor=\"%s\"\n", color))
	sb.WriteString("    style=\"filled\"\n")
	sb.WriteString("    label=<\n")
	
	// HTML-like label for better formatting
	sb.WriteString("      <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n")
	
	// Table header
	sb.WriteString(fmt.Sprintf("        <tr><td colspan=\"3\" bgcolor=\"%s\"><b>%s</b></td></tr>\n", 
		darkenColor(color), table.Name))
	
	// Column header
	if g.ShowDataTypes {
		sb.WriteString("        <tr><td bgcolor=\"#f0f0f0\"><b>Column</b></td><td bgcolor=\"#f0f0f0\"><b>Type</b></td><td bgcolor=\"#f0f0f0\"><b>Constraints</b></td></tr>\n")
	} else {
		sb.WriteString("        <tr><td bgcolor=\"#f0f0f0\"><b>Columns</b></td></tr>\n")
	}
	
	// Columns
	for _, col := range table.Columns {
		g.generateGraphVizColumn(sb, col, table.Constraints)
	}
	
	sb.WriteString("      </table>\n")
	sb.WriteString("    >\n")
	sb.WriteString("  ];\n\n")
}

func (g *GraphVizGenerator) generateGraphVizColumn(sb *strings.Builder, col models.Column, constraints []models.Constraint) {
	// Build constraints info
	var constraintInfo []string
	
	if col.IsPrimaryKey {
		constraintInfo = append(constraintInfo, "PK")
	}
	
	// Check if foreign key
	for _, constraint := range constraints {
		if constraint.Type == models.ForeignKey {
			for _, fkCol := range constraint.Columns {
				if fkCol == col.Name {
					constraintInfo = append(constraintInfo, "FK")
					break
				}
			}
		}
	}
	
	if col.IsUnique {
		constraintInfo = append(constraintInfo, "UQ")
	}
	
	if !col.IsNullable {
		constraintInfo = append(constraintInfo, "NN")
	}
	
	constraintStr := strings.Join(constraintInfo, ",")
	
	if g.ShowDataTypes {
		// Port for foreign key connections
		port := ""
		if strings.Contains(constraintStr, "FK") {
			port = fmt.Sprintf(" port=\"%s\"", col.Name)
		}
		
		sb.WriteString(fmt.Sprintf("        <tr><td align=\"left\"%s>%s</td><td align=\"left\">%s</td><td>%s</td></tr>\n",
			port, col.Name, simplifyDataTypeForGraph(col.DataType), constraintStr))
	} else {
		columnDisplay := col.Name
		if constraintStr != "" {
			columnDisplay = fmt.Sprintf("%s [%s]", col.Name, constraintStr)
		}
		sb.WriteString(fmt.Sprintf("        <tr><td align=\"left\">%s</td></tr>\n", columnDisplay))
	}
}

func (g *GraphVizGenerator) generateGraphVizRelationship(sb *strings.Builder, tableName string, constraint models.Constraint) {
	fromTable := sanitizeGraphVizName(tableName)
	toTable := sanitizeGraphVizName(constraint.ReferencedTable)
	
	// Determine arrow style based on relationship
	arrowhead := "crow"  // Many side
	arrowtail := "tee"   // One side
	
	// Edge attributes
	sb.WriteString(fmt.Sprintf("  %s -> %s [\n", fromTable, toTable))
	sb.WriteString(fmt.Sprintf("    arrowhead=\"%s\"\n", arrowhead))
	sb.WriteString(fmt.Sprintf("    arrowtail=\"%s\"\n", arrowtail))
	sb.WriteString("    dir=\"both\"\n")
	sb.WriteString(fmt.Sprintf("    label=\"%s\"\n", constraint.Name))
	sb.WriteString("    fontsize=9\n")
	sb.WriteString("  ];\n")
}

func (g *GraphVizGenerator) getColorScheme() []string {
	switch g.ColorScheme {
	case "blues":
		return []string{"#e6f2ff", "#cce5ff", "#b3d9ff", "#99ccff", "#80bfff"}
	case "greens":
		return []string{"#e6ffe6", "#ccffcc", "#b3ffb3", "#99ff99", "#80ff80"}
	case "accent":
		return []string{"#ffe6e6", "#ffcce6", "#e6e6ff", "#e6ffff", "#ffffe6"}
	case "pastel":
		return []string{"#ffd4e5", "#d4e5ff", "#ffffd4", "#e5d4ff", "#d4ffd4"}
	default:
		return []string{"#f0f0f0", "#e0e0e0", "#d0d0d0", "#c0c0c0", "#b0b0b0"}
	}
}

func darkenColor(color string) string {
	// Simple darkening by reducing each RGB component
	if strings.HasPrefix(color, "#") && len(color) == 7 {
		// Parse hex color
		r := color[1:3]
		g := color[3:5]
		b := color[5:7]
		
		// Darken by 20%
		rInt := int(float64(hexToInt(r)) * 0.8)
		gInt := int(float64(hexToInt(g)) * 0.8)
		bInt := int(float64(hexToInt(b)) * 0.8)
		
		return fmt.Sprintf("#%02x%02x%02x", rInt, gInt, bInt)
	}
	return color
}

func hexToInt(hex string) int {
	var val int
	fmt.Sscanf(hex, "%x", &val)
	return val
}

func sanitizeGraphVizName(name string) string {
	// GraphViz node names should be alphanumeric
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

func simplifyDataTypeForGraph(dataType string) string {
	// Shorten common data types for cleaner diagrams
	replacements := map[string]string{
		"CHARACTER VARYING": "VARCHAR",
		"TIMESTAMP WITHOUT TIME ZONE": "TIMESTAMP",
		"TIMESTAMP WITH TIME ZONE": "TIMESTAMPTZ",
		"INTEGER": "INT",
		"NUMERIC": "DECIMAL",
		"BOOLEAN": "BOOL",
	}
	
	upperType := strings.ToUpper(dataType)
	for old, new := range replacements {
		if strings.Contains(upperType, old) {
			dataType = strings.Replace(dataType, old, new, -1)
			dataType = strings.Replace(dataType, strings.ToLower(old), new, -1)
		}
	}
	
	return dataType
}

// D2Generator generates D2 diagrams (modern diagramming language)
type D2Generator struct{}

func NewD2Generator() *D2Generator {
	return &D2Generator{}
}

func (g *D2Generator) Generate(schema *models.Schema) (string, error) {
	var sb strings.Builder
	
	// D2 header
	sb.WriteString(fmt.Sprintf("# %s Database Schema\n\n", schema.Name))
	sb.WriteString("direction: down\n\n")
	
	// Style definitions
	sb.WriteString("style: {\n")
	sb.WriteString("  fill: white\n")
	sb.WriteString("  stroke: black\n")
	sb.WriteString("  stroke-width: 2\n")
	sb.WriteString("}\n\n")
	
	// Generate tables
	for _, table := range schema.Tables {
		g.generateD2Table(&sb, table)
	}
	
	// Generate relationships
	for _, table := range schema.Tables {
		for _, constraint := range table.Constraints {
			if constraint.Type == models.ForeignKey {
				g.generateD2Relationship(&sb, table.Name, constraint)
			}
		}
	}
	
	return sb.String(), nil
}

func (g *D2Generator) generateD2Table(sb *strings.Builder, table models.Table) {
	tableName := sanitizeD2Name(table.Name)
	
	sb.WriteString(fmt.Sprintf("%s: {\n", tableName))
	sb.WriteString("  shape: sql_table\n")
	sb.WriteString(fmt.Sprintf("  label: %s\n", table.Name))
	
	// Primary keys
	var pks []string
	for _, col := range table.Columns {
		if col.IsPrimaryKey {
			pks = append(pks, fmt.Sprintf("%s: %s", col.Name, simplifyDataType(col.DataType)))
		}
	}
	
	if len(pks) > 0 {
		sb.WriteString("  constraint: [\n")
		for _, pk := range pks {
			sb.WriteString(fmt.Sprintf("    %s\n", pk))
		}
		sb.WriteString("  ]\n")
	}
	
	// Other columns
	for _, col := range table.Columns {
		if !col.IsPrimaryKey {
			nullable := ""
			if col.IsNullable {
				nullable = "?"
			}
			sb.WriteString(fmt.Sprintf("  %s: %s%s\n", col.Name, simplifyDataType(col.DataType), nullable))
		}
	}
	
	sb.WriteString("}\n\n")
}

func (g *D2Generator) generateD2Relationship(sb *strings.Builder, tableName string, constraint models.Constraint) {
	fromTable := sanitizeD2Name(tableName)
	toTable := sanitizeD2Name(constraint.ReferencedTable)
	
	sb.WriteString(fmt.Sprintf("%s -> %s: %s {\n", fromTable, toTable, constraint.Name))
	sb.WriteString("  source-arrowhead: cf-many\n")
	sb.WriteString("  target-arrowhead: cf-one\n")
	sb.WriteString("}\n")
}

func sanitizeD2Name(name string) string {
	// D2 names should avoid spaces and special chars
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}