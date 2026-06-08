package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nechja/schemalyzer/pkg/models"
)

type Hasher struct {
	includeComments bool
	verbose         bool
}

func NewHasher() *Hasher {
	return &Hasher{
		includeComments: false,
		verbose:         false,
	}
}

func (h *Hasher) WithVerbose(verbose bool) *Hasher {
	h.verbose = verbose
	return h
}

func (h *Hasher) GenerateFingerprint(schema *models.Schema) (string, error) {
	normalized := h.normalizeSchema(schema)

	data, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (h *Hasher) normalizeSchema(schema *models.Schema) map[string]interface{} {
	result := make(map[string]interface{})

	result["tables"] = h.normalizeTables(schema.Tables)
	result["views"] = h.normalizeViews(schema.Views)
	result["indexes"] = h.normalizeIndexes(schema.Indexes)
	result["sequences"] = h.normalizeSequences(schema.Sequences)
	result["procedures"] = h.normalizeProcedures(schema.Procedures)
	result["functions"] = h.normalizeFunctions(schema.Functions)
	result["triggers"] = h.normalizeTriggers(schema.Triggers)

	return result
}

func (h *Hasher) normalizeTables(tables []models.Table) []map[string]interface{} {
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	var result []map[string]interface{}
	for _, table := range tables {
		normalized := map[string]interface{}{
			"name":        table.Name,
			"columns":     h.normalizeColumns(table.Columns),
			"constraints": h.normalizeConstraints(table.Constraints),
			"indexes":     h.normalizeTableIndexes(table.Indexes),
		}

		if h.includeComments && table.Comment != "" {
			normalized["comment"] = table.Comment
		}

		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeColumns(columns []models.Column) []map[string]interface{} {
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Name < columns[j].Name
	})

	var result []map[string]interface{}
	for _, col := range columns {
		normalized := map[string]interface{}{
			"name":        col.Name,
			"type":        col.DataType,
			"nullable":    col.IsNullable,
			"position":    col.Position,
			"primary_key": col.IsPrimaryKey,
			"unique":      col.IsUnique,
		}

		if col.DefaultValue != nil {
			normalized["default"] = *col.DefaultValue
		}

		if col.IsAutoIncrement {
			normalized["auto_increment"] = true
		}

		if h.includeComments && col.Comment != "" {
			normalized["comment"] = col.Comment
		}

		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeConstraints(constraints []models.Constraint) []map[string]interface{} {
	// Hash constraints by their semantic content rather than name. Constraint
	// names are frequently system-generated (PostgreSQL OID-based, Oracle SYS_C*)
	// and differ between structurally-identical schemas, so including them would
	// make a schema's fingerprint unstable across environments. The trade-off is
	// that a pure rename (same definition, new name) is not treated as drift.
	type entry struct {
		sig  string
		data map[string]interface{}
	}

	entries := make([]entry, 0, len(constraints))
	for _, c := range constraints {
		normalized := map[string]interface{}{
			"type": c.Type,
		}

		cols := make([]string, len(c.Columns))
		copy(cols, c.Columns)
		sort.Strings(cols)
		if len(cols) > 0 {
			normalized["columns"] = cols
		}

		if c.ReferencedTable != "" {
			normalized["ref_table"] = c.ReferencedTable
		}

		refCols := make([]string, len(c.ReferencedColumn))
		copy(refCols, c.ReferencedColumn)
		sort.Strings(refCols)
		if len(refCols) > 0 {
			normalized["ref_columns"] = refCols
		}

		if c.OnUpdate != "" {
			normalized["on_update"] = c.OnUpdate
		}
		if c.OnDelete != "" {
			normalized["on_delete"] = c.OnDelete
		}

		if c.CheckExpression != "" {
			normalized["check_expr"] = c.CheckExpression
		}

		sig := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
			c.Type, strings.Join(cols, ","), c.ReferencedTable,
			strings.Join(refCols, ","), c.OnUpdate, c.OnDelete, c.CheckExpression)
		entries = append(entries, entry{sig: sig, data: normalized})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].sig < entries[j].sig
	})

	result := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		result = append(result, e.data)
	}

	return result
}

func (h *Hasher) normalizeTableIndexes(indexes []models.Index) []map[string]interface{} {
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i].Name < indexes[j].Name
	})

	var result []map[string]interface{}
	for _, idx := range indexes {
		// Sort columns for consistent hashing
		sortedColumns := make([]string, len(idx.Columns))
		copy(sortedColumns, idx.Columns)
		sort.Strings(sortedColumns)

		normalized := map[string]interface{}{
			"name":    idx.Name,
			"unique":  idx.IsUnique,
			"columns": sortedColumns,
		}

		if idx.Type != "" {
			normalized["type"] = idx.Type
		}

		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeIndexes(indexes []models.Index) []map[string]interface{} {
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i].Name < indexes[j].Name
	})

	var result []map[string]interface{}
	for _, idx := range indexes {
		// Sort columns for consistent hashing
		sortedColumns := make([]string, len(idx.Columns))
		copy(sortedColumns, idx.Columns)
		sort.Strings(sortedColumns)

		normalized := map[string]interface{}{
			"name":    idx.Name,
			"table":   idx.TableName,
			"unique":  idx.IsUnique,
			"columns": sortedColumns,
		}

		if idx.Type != "" {
			normalized["type"] = idx.Type
		}

		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeViews(views []models.View) []map[string]interface{} {
	sort.Slice(views, func(i, j int) bool {
		return views[i].Name < views[j].Name
	})

	var result []map[string]interface{}
	for _, view := range views {
		normalized := map[string]interface{}{
			"name":       view.Name,
			"definition": view.Definition,
		}
		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeSequences(sequences []models.Sequence) []map[string]interface{} {
	sort.Slice(sequences, func(i, j int) bool {
		return sequences[i].Name < sequences[j].Name
	})

	var result []map[string]interface{}
	for _, seq := range sequences {
		normalized := map[string]interface{}{
			"name":      seq.Name,
			"start":     seq.StartValue,
			"increment": seq.Increment,
			"min_value": seq.MinValue,
			"max_value": seq.MaxValue,
			"cyclic":    seq.IsCyclic,
		}
		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeProcedures(procedures []models.Procedure) []map[string]interface{} {
	sort.Slice(procedures, func(i, j int) bool {
		return procedures[i].Name < procedures[j].Name
	})

	var result []map[string]interface{}
	for _, proc := range procedures {
		normalized := map[string]interface{}{
			"name":       proc.Name,
			"parameters": h.normalizeParameters(proc.Parameters),
			"body":       proc.Body,
		}
		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeFunctions(functions []models.Function) []map[string]interface{} {
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	var result []map[string]interface{}
	for _, fn := range functions {
		normalized := map[string]interface{}{
			"name":        fn.Name,
			"parameters":  h.normalizeParameters(fn.Parameters),
			"return_type": fn.ReturnType,
			"body":        fn.Body,
		}
		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeTriggers(triggers []models.Trigger) []map[string]interface{} {
	sort.Slice(triggers, func(i, j int) bool {
		return triggers[i].Name < triggers[j].Name
	})

	var result []map[string]interface{}
	for _, trig := range triggers {
		normalized := map[string]interface{}{
			"name":   trig.Name,
			"table":  trig.TableName,
			"event":  trig.Event,
			"timing": trig.Timing,
			"body":   trig.Body,
		}
		result = append(result, normalized)
	}

	return result
}

func (h *Hasher) normalizeParameters(params []models.Parameter) []map[string]interface{} {
	// Sort parameters by name for consistent hashing
	sort.Slice(params, func(i, j int) bool {
		return params[i].Name < params[j].Name
	})

	var result []map[string]interface{}
	for _, p := range params {
		normalized := map[string]interface{}{
			"name":      p.Name,
			"type":      p.DataType,
			"direction": p.Direction,
		}
		result = append(result, normalized)
	}
	return result
}
