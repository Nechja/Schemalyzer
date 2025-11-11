package compare

import (
	"github.com/nechja/schemalyzer/pkg/models"
	"reflect"
	"time"
)

type Comparer struct {
	ignoreConfig *models.IgnoreConfig
}

func NewComparer() *Comparer {
	return &Comparer{}
}

func NewComparerWithIgnore(ignoreConfig *models.IgnoreConfig) *Comparer {
	return &Comparer{
		ignoreConfig: ignoreConfig,
	}
}

func (c *Comparer) Compare(source, target *models.Schema) *models.ComparisonResult {
	result := &models.ComparisonResult{
		SourceSchema:   source,
		TargetSchema:   target,
		Differences:    []models.Difference{},
		ComparisonTime: time.Now(),
	}

	// Compare tables
	result.Differences = append(result.Differences, c.compareTables(source.Tables, target.Tables)...)

	// Compare views
	result.Differences = append(result.Differences, c.compareViews(source.Views, target.Views)...)

	// Compare indexes
	result.Differences = append(result.Differences, c.compareIndexes(source.Indexes, target.Indexes)...)

	// Compare sequences
	result.Differences = append(result.Differences, c.compareSequences(source.Sequences, target.Sequences)...)

	// Compare procedures
	result.Differences = append(result.Differences, c.compareProcedures(source.Procedures, target.Procedures)...)

	// Compare functions
	result.Differences = append(result.Differences, c.compareFunctions(source.Functions, target.Functions)...)

	// Compare triggers
	result.Differences = append(result.Differences, c.compareTriggers(source.Triggers, target.Triggers)...)

	return result
}

func (c *Comparer) compareTables(source, target []models.Table) []models.Difference {
	var differences []models.Difference

	// Filter out ignored tables
	if c.ignoreConfig != nil {
		source = c.filterTables(source)
		target = c.filterTables(target)
	}

	sourceMap := make(map[string]*models.Table)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Table)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed tables
	for name, table := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Table",
				ObjectName:  name,
				Source:      table,
				Description: "Table exists in source but not in target",
			})
		}
	}

	// Check for added tables
	for name, table := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Table",
				ObjectName:  name,
				Target:      table,
				Description: "Table exists in target but not in source",
			})
		}
	}

	// Check for modified tables
	for name, sourceTable := range sourceMap {
		if targetTable, exists := targetMap[name]; exists {
			tableDiffs := c.compareTable(sourceTable, targetTable)
			differences = append(differences, tableDiffs...)
		}
	}

	return differences
}

func (c *Comparer) compareTable(source, target *models.Table) []models.Difference {
	var differences []models.Difference

	// Compare columns
	columnDiffs := c.compareColumns(source.Name, source.Columns, target.Columns)
	differences = append(differences, columnDiffs...)

	// Compare constraints
	constraintDiffs := c.compareConstraints(source.Name, source.Constraints, target.Constraints)
	differences = append(differences, constraintDiffs...)

	// Compare indexes
	indexDiffs := c.compareTableIndexes(source.Name, source.Indexes, target.Indexes)
	differences = append(differences, indexDiffs...)

	// Compare comment
	if source.Comment != target.Comment {
		differences = append(differences, models.Difference{
			Type:        models.Modified,
			ObjectType:  "Table Comment",
			ObjectName:  source.Name,
			Source:      source.Comment,
			Target:      target.Comment,
			Description: "Table comment changed",
		})
	}

	return differences
}

func (c *Comparer) compareColumns(tableName string, source, target []models.Column) []models.Difference {
	var differences []models.Difference

	sourceMap := make(map[string]*models.Column)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Column)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed columns
	for name, column := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Column",
				ObjectName:  tableName + "." + name,
				Source:      column,
				Description: "Column removed from table",
			})
		}
	}

	// Check for added columns
	for name, column := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Column",
				ObjectName:  tableName + "." + name,
				Target:      column,
				Description: "Column added to table",
			})
		}
	}

	// Check for modified columns
	for name, sourceCol := range sourceMap {
		if targetCol, exists := targetMap[name]; exists {
			if !c.columnsEqual(sourceCol, targetCol) {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Column",
					ObjectName:  tableName + "." + name,
					Source:      sourceCol,
					Target:      targetCol,
					Description: "Column definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) columnsEqual(source, target *models.Column) bool {
	if source.DataType != target.DataType {
		return false
	}
	if source.IsNullable != target.IsNullable {
		return false
	}
	if !c.stringPointersEqual(source.DefaultValue, target.DefaultValue) {
		return false
	}
	if source.IsPrimaryKey != target.IsPrimaryKey {
		return false
	}
	if source.IsUnique != target.IsUnique {
		return false
	}
	if source.IsAutoIncrement != target.IsAutoIncrement {
		return false
	}
	if source.Comment != target.Comment {
		return false
	}
	return true
}

func (c *Comparer) stringPointersEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (c *Comparer) compareConstraints(tableName string, source, target []models.Constraint) []models.Difference {
	var differences []models.Difference

	sourceMap := make(map[string]*models.Constraint)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Constraint)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed constraints
	for name, constraint := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Constraint",
				ObjectName:  tableName + "." + name,
				Source:      constraint,
				Description: "Constraint removed from table",
			})
		}
	}

	// Check for added constraints
	for name, constraint := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Constraint",
				ObjectName:  tableName + "." + name,
				Target:      constraint,
				Description: "Constraint added to table",
			})
		}
	}

	// Check for modified constraints
	for name, sourceConstraint := range sourceMap {
		if targetConstraint, exists := targetMap[name]; exists {
			if !c.constraintsEqual(sourceConstraint, targetConstraint) {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Constraint",
					ObjectName:  tableName + "." + name,
					Source:      sourceConstraint,
					Target:      targetConstraint,
					Description: "Constraint definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) constraintsEqual(source, target *models.Constraint) bool {
	if source.Type != target.Type {
		return false
	}
	// Compare columns as sets, not ordered lists
	if !c.stringSlicesEqualAsSet(source.Columns, target.Columns) {
		return false
	}
	if source.ReferencedTable != target.ReferencedTable {
		return false
	}
	// Compare referenced columns as sets
	if !c.stringSlicesEqualAsSet(source.ReferencedColumn, target.ReferencedColumn) {
		return false
	}
	if source.CheckExpression != target.CheckExpression {
		return false
	}
	return true
}

func (c *Comparer) stringSlicesEqualAsSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Count occurrences in each slice
	aCount := make(map[string]int)
	for _, s := range a {
		aCount[s]++
	}

	bCount := make(map[string]int)
	for _, s := range b {
		bCount[s]++
	}

	// Compare counts
	for k, v := range aCount {
		if bCount[k] != v {
			return false
		}
	}

	return true
}

func (c *Comparer) compareTableIndexes(tableName string, source, target []models.Index) []models.Difference {
	var differences []models.Difference

	sourceMap := make(map[string]*models.Index)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Index)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed indexes
	for name, index := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Index",
				ObjectName:  tableName + "." + name,
				Source:      index,
				Description: "Index removed from table",
			})
		}
	}

	// Check for added indexes
	for name, index := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Index",
				ObjectName:  tableName + "." + name,
				Target:      index,
				Description: "Index added to table",
			})
		}
	}

	// Check for modified indexes
	for name, sourceIndex := range sourceMap {
		if targetIndex, exists := targetMap[name]; exists {
			if !c.indexesEqual(sourceIndex, targetIndex) {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Index",
					ObjectName:  tableName + "." + name,
					Source:      sourceIndex,
					Target:      targetIndex,
					Description: "Index definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) indexesEqual(source, target *models.Index) bool {
	if source.IsUnique != target.IsUnique {
		return false
	}
	if source.Type != target.Type {
		return false
	}
	// For indexes, column order matters, so we still use ordered comparison
	if !reflect.DeepEqual(source.Columns, target.Columns) {
		return false
	}
	return true
}

func (c *Comparer) compareViews(source, target []models.View) []models.Difference {
	var differences []models.Difference

	// Filter out ignored views
	if c.ignoreConfig != nil {
		source = c.filterViews(source)
		target = c.filterViews(target)
	}

	sourceMap := make(map[string]*models.View)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.View)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed views
	for name, view := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "View",
				ObjectName:  name,
				Source:      view,
				Description: "View exists in source but not in target",
			})
		}
	}

	// Check for added views
	for name, view := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "View",
				ObjectName:  name,
				Target:      view,
				Description: "View exists in target but not in source",
			})
		}
	}

	// Check for modified views
	for name, sourceView := range sourceMap {
		if targetView, exists := targetMap[name]; exists {
			if sourceView.Definition != targetView.Definition {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "View",
					ObjectName:  name,
					Source:      sourceView,
					Target:      targetView,
					Description: "View definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) compareIndexes(source, target []models.Index) []models.Difference {
	var differences []models.Difference

	// Filter out ignored indexes
	if c.ignoreConfig != nil {
		source = c.filterIndexes(source)
		target = c.filterIndexes(target)
	}

	sourceMap := make(map[string]*models.Index)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Index)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed indexes
	for name, index := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Index",
				ObjectName:  name,
				Source:      index,
				Description: "Index exists in source but not in target",
			})
		}
	}

	// Check for added indexes
	for name, index := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Index",
				ObjectName:  name,
				Target:      index,
				Description: "Index exists in target but not in source",
			})
		}
	}

	// Check for modified indexes
	for name, sourceIndex := range sourceMap {
		if targetIndex, exists := targetMap[name]; exists {
			if !c.indexesEqual(sourceIndex, targetIndex) {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Index",
					ObjectName:  name,
					Source:      sourceIndex,
					Target:      targetIndex,
					Description: "Index definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) compareSequences(source, target []models.Sequence) []models.Difference {
	var differences []models.Difference

	// Filter out ignored sequences
	if c.ignoreConfig != nil {
		source = c.filterSequences(source)
		target = c.filterSequences(target)
	}

	sourceMap := make(map[string]*models.Sequence)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Sequence)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed sequences
	for name, sequence := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Sequence",
				ObjectName:  name,
				Source:      sequence,
				Description: "Sequence exists in source but not in target",
			})
		}
	}

	// Check for added sequences
	for name, sequence := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Sequence",
				ObjectName:  name,
				Target:      sequence,
				Description: "Sequence exists in target but not in source",
			})
		}
	}

	// Check for modified sequences
	for name, sourceSeq := range sourceMap {
		if targetSeq, exists := targetMap[name]; exists {
			if !c.sequencesEqual(sourceSeq, targetSeq) {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Sequence",
					ObjectName:  name,
					Source:      sourceSeq,
					Target:      targetSeq,
					Description: "Sequence definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) sequencesEqual(source, target *models.Sequence) bool {
	return source.StartValue == target.StartValue &&
		source.Increment == target.Increment &&
		source.MinValue == target.MinValue &&
		source.MaxValue == target.MaxValue &&
		source.IsCyclic == target.IsCyclic
}

func (c *Comparer) compareProcedures(source, target []models.Procedure) []models.Difference {
	var differences []models.Difference

	// Filter out ignored procedures
	if c.ignoreConfig != nil {
		source = c.filterProcedures(source)
		target = c.filterProcedures(target)
	}

	sourceMap := make(map[string]*models.Procedure)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Procedure)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed procedures
	for name, procedure := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Procedure",
				ObjectName:  name,
				Source:      procedure,
				Description: "Procedure exists in source but not in target",
			})
		}
	}

	// Check for added procedures
	for name, procedure := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Procedure",
				ObjectName:  name,
				Target:      procedure,
				Description: "Procedure exists in target but not in source",
			})
		}
	}

	// Check for modified procedures
	for name, sourceProc := range sourceMap {
		if targetProc, exists := targetMap[name]; exists {
			if sourceProc.Body != targetProc.Body {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Procedure",
					ObjectName:  name,
					Source:      sourceProc,
					Target:      targetProc,
					Description: "Procedure definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) compareFunctions(source, target []models.Function) []models.Difference {
	var differences []models.Difference

	// Filter out ignored functions
	if c.ignoreConfig != nil {
		source = c.filterFunctions(source)
		target = c.filterFunctions(target)
	}

	sourceMap := make(map[string]*models.Function)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Function)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed functions
	for name, function := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Function",
				ObjectName:  name,
				Source:      function,
				Description: "Function exists in source but not in target",
			})
		}
	}

	// Check for added functions
	for name, function := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Function",
				ObjectName:  name,
				Target:      function,
				Description: "Function exists in target but not in source",
			})
		}
	}

	// Check for modified functions
	for name, sourceFunc := range sourceMap {
		if targetFunc, exists := targetMap[name]; exists {
			if sourceFunc.Body != targetFunc.Body || sourceFunc.ReturnType != targetFunc.ReturnType {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Function",
					ObjectName:  name,
					Source:      sourceFunc,
					Target:      targetFunc,
					Description: "Function definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) compareTriggers(source, target []models.Trigger) []models.Difference {
	var differences []models.Difference

	// Filter out ignored triggers
	if c.ignoreConfig != nil {
		source = c.filterTriggers(source)
		target = c.filterTriggers(target)
	}

	sourceMap := make(map[string]*models.Trigger)
	for i := range source {
		sourceMap[source[i].Name] = &source[i]
	}

	targetMap := make(map[string]*models.Trigger)
	for i := range target {
		targetMap[target[i].Name] = &target[i]
	}

	// Check for removed triggers
	for name, trigger := range sourceMap {
		if _, exists := targetMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Removed,
				ObjectType:  "Trigger",
				ObjectName:  name,
				Source:      trigger,
				Description: "Trigger exists in source but not in target",
			})
		}
	}

	// Check for added triggers
	for name, trigger := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			differences = append(differences, models.Difference{
				Type:        models.Added,
				ObjectType:  "Trigger",
				ObjectName:  name,
				Target:      trigger,
				Description: "Trigger exists in target but not in source",
			})
		}
	}

	// Check for modified triggers
	for name, sourceTrigger := range sourceMap {
		if targetTrigger, exists := targetMap[name]; exists {
			if !c.triggersEqual(sourceTrigger, targetTrigger) {
				differences = append(differences, models.Difference{
					Type:        models.Modified,
					ObjectType:  "Trigger",
					ObjectName:  name,
					Source:      sourceTrigger,
					Target:      targetTrigger,
					Description: "Trigger definition changed",
				})
			}
		}
	}

	return differences
}

func (c *Comparer) triggersEqual(source, target *models.Trigger) bool {
	return source.TableName == target.TableName &&
		source.Event == target.Event &&
		source.Timing == target.Timing &&
		source.Body == target.Body
}

// Filter methods for ignore patterns
func (c *Comparer) filterTables(tables []models.Table) []models.Table {
	var filtered []models.Table
	for _, table := range tables {
		if !c.ignoreConfig.ShouldIgnore("table", table.Name) {
			// Also filter columns, constraints, and indexes within the table
			table.Columns = c.filterColumns(table.Columns)
			table.Constraints = c.filterConstraints(table.Constraints)
			table.Indexes = c.filterIndexes(table.Indexes)
			filtered = append(filtered, table)
		}
	}
	return filtered
}

func (c *Comparer) filterColumns(columns []models.Column) []models.Column {
	var filtered []models.Column
	for _, column := range columns {
		if !c.ignoreConfig.ShouldIgnore("column", column.Name) {
			filtered = append(filtered, column)
		}
	}
	return filtered
}

func (c *Comparer) filterConstraints(constraints []models.Constraint) []models.Constraint {
	var filtered []models.Constraint
	for _, constraint := range constraints {
		if !c.ignoreConfig.ShouldIgnore("constraint", constraint.Name) {
			filtered = append(filtered, constraint)
		}
	}
	return filtered
}

func (c *Comparer) filterIndexes(indexes []models.Index) []models.Index {
	var filtered []models.Index
	for _, index := range indexes {
		if !c.ignoreConfig.ShouldIgnore("index", index.Name) {
			filtered = append(filtered, index)
		}
	}
	return filtered
}

func (c *Comparer) filterViews(views []models.View) []models.View {
	var filtered []models.View
	for _, view := range views {
		if !c.ignoreConfig.ShouldIgnore("view", view.Name) {
			filtered = append(filtered, view)
		}
	}
	return filtered
}

func (c *Comparer) filterSequences(sequences []models.Sequence) []models.Sequence {
	var filtered []models.Sequence
	for _, sequence := range sequences {
		if !c.ignoreConfig.ShouldIgnore("sequence", sequence.Name) {
			filtered = append(filtered, sequence)
		}
	}
	return filtered
}

func (c *Comparer) filterProcedures(procedures []models.Procedure) []models.Procedure {
	var filtered []models.Procedure
	for _, procedure := range procedures {
		if !c.ignoreConfig.ShouldIgnore("procedure", procedure.Name) {
			filtered = append(filtered, procedure)
		}
	}
	return filtered
}

func (c *Comparer) filterFunctions(functions []models.Function) []models.Function {
	var filtered []models.Function
	for _, function := range functions {
		if !c.ignoreConfig.ShouldIgnore("function", function.Name) {
			filtered = append(filtered, function)
		}
	}
	return filtered
}

func (c *Comparer) filterTriggers(triggers []models.Trigger) []models.Trigger {
	var filtered []models.Trigger
	for _, trigger := range triggers {
		if !c.ignoreConfig.ShouldIgnore("trigger", trigger.Name) {
			filtered = append(filtered, trigger)
		}
	}
	return filtered
}
