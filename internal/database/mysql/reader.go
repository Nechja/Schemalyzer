package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nechja/schemalyzer/pkg/models"
	"strings"
)

type MySQLReader struct {
	db *sql.DB
}

func NewMySQLReader() *MySQLReader {
	return &MySQLReader{}
}

func (r *MySQLReader) Connect(ctx context.Context, connectionString string) error {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to mysql: %w", err)
	}
	
	// Configure connection pool for large databases
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No timeout
	
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}
	
	r.db = db
	return nil
}

func (r *MySQLReader) GetSchema(ctx context.Context, schemaName string) (*models.Schema, error) {
	schema := &models.Schema{
		Name:         schemaName,
		DatabaseType: models.MySQL,
	}
	
	// Use parallel fetching for better performance on large databases
	type result struct {
		tables     []models.Table
		views      []models.View
		sequences  []models.Sequence
		functions  []models.Function
		procedures []models.Procedure
		triggers   []models.Trigger
		err        error
	}
	
	var wg sync.WaitGroup
	res := &result{}
	
	// Fetch schema objects in parallel
	wg.Add(5) // MySQL doesn't have sequences traditionally
	
	go func() {
		defer wg.Done()
		tables, err := r.getTables(ctx, schemaName)
		if err != nil {
			res.err = fmt.Errorf("failed to get tables: %w", err)
			return
		}
		res.tables = tables
	}()
	
	go func() {
		defer wg.Done()
		views, err := r.getViews(ctx, schemaName)
		if err != nil {
			res.err = fmt.Errorf("failed to get views: %w", err)
			return
		}
		res.views = views
	}()
	
	// MySQL doesn't have sequences traditionally - just set empty
	res.sequences = []models.Sequence{}
	
	go func() {
		defer wg.Done()
		functions, err := r.getFunctions(ctx, schemaName)
		if err != nil {
			res.err = fmt.Errorf("failed to get functions: %w", err)
			return
		}
		res.functions = functions
	}()
	
	go func() {
		defer wg.Done()
		procedures, err := r.getProcedures(ctx, schemaName)
		if err != nil {
			res.err = fmt.Errorf("failed to get procedures: %w", err)
			return
		}
		res.procedures = procedures
	}()
	
	go func() {
		defer wg.Done()
		triggers, err := r.getTriggers(ctx, schemaName)
		if err != nil {
			res.err = fmt.Errorf("failed to get triggers: %w", err)
			return
		}
		res.triggers = triggers
	}()
	
	wg.Wait()
	
	if res.err != nil {
		return nil, res.err
	}
	
	schema.Tables = res.tables
	schema.Views = res.views
	schema.Sequences = res.sequences
	schema.Functions = res.functions
	schema.Procedures = res.procedures
	schema.Triggers = res.triggers
	
	return schema, nil
}

func (r *MySQLReader) ListSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
		ORDER BY schema_name`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()
	
	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}
	
	return schemas, nil
}

func (r *MySQLReader) getTables(ctx context.Context, schemaName string) ([]models.Table, error) {
	query := `
		SELECT table_name, table_comment
		FROM information_schema.tables
		WHERE table_schema = ? AND table_type = 'BASE TABLE'
		ORDER BY table_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tables []models.Table
	for rows.Next() {
		var table models.Table
		
		if err := rows.Scan(&table.Name, &table.Comment); err != nil {
			return nil, err
		}
		
		table.Schema = schemaName
		
		columns, err := r.getColumns(ctx, schemaName, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table.Name, err)
		}
		table.Columns = columns
		
		constraints, err := r.getConstraints(ctx, schemaName, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get constraints for table %s: %w", table.Name, err)
		}
		table.Constraints = constraints
		
		indexes, err := r.getIndexes(ctx, schemaName, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", table.Name, err)
		}
		table.Indexes = indexes
		
		tables = append(tables, table)
	}
	
	return tables, nil
}

func (r *MySQLReader) getColumns(ctx context.Context, schemaName, tableName string) ([]models.Column, error) {
	query := `
		SELECT 
			column_name,
			column_type,
			is_nullable,
			column_default,
			ordinal_position,
			column_comment,
			column_key
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var isNullable, columnKey string
		var defaultValue sql.NullString
		
		if err := rows.Scan(&col.Name, &col.DataType, &isNullable, &defaultValue, 
			&col.Position, &col.Comment, &columnKey); err != nil {
			return nil, err
		}
		
		col.IsNullable = isNullable == "YES"
		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}
		
		if columnKey == "PRI" {
			col.IsPrimaryKey = true
		} else if columnKey == "UNI" {
			col.IsUnique = true
		}
		
		columns = append(columns, col)
	}
	
	return columns, nil
}

func (r *MySQLReader) getConstraints(ctx context.Context, schemaName, tableName string) ([]models.Constraint, error) {
	// Primary Key and Unique constraints
	uniqueQuery := `
		SELECT 
			constraint_name,
			constraint_type
		FROM information_schema.table_constraints
		WHERE table_schema = ? AND table_name = ?
		AND constraint_type IN ('PRIMARY KEY', 'UNIQUE')`
	
	rows, err := r.db.QueryContext(ctx, uniqueQuery, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var constraints []models.Constraint
	constraintMap := make(map[string]*models.Constraint)
	
	for rows.Next() {
		var constraint models.Constraint
		var constraintType string
		
		if err := rows.Scan(&constraint.Name, &constraintType); err != nil {
			return nil, err
		}
		
		switch constraintType {
		case "PRIMARY KEY":
			constraint.Type = models.PrimaryKey
		case "UNIQUE":
			constraint.Type = models.Unique
		}
		
		constraintMap[constraint.Name] = &constraint
	}
	
	// Get columns for each constraint
	for name, constraint := range constraintMap {
		columnQuery := `
			SELECT column_name
			FROM information_schema.key_column_usage
			WHERE table_schema = ? AND table_name = ? AND constraint_name = ?
			ORDER BY ordinal_position`
		
		colRows, err := r.db.QueryContext(ctx, columnQuery, schemaName, tableName, name)
		if err != nil {
			return nil, err
		}
		
		for colRows.Next() {
			var columnName string
			if err := colRows.Scan(&columnName); err != nil {
				colRows.Close()
				return nil, err
			}
			constraint.Columns = append(constraint.Columns, columnName)
		}
		colRows.Close()
		
		constraints = append(constraints, *constraint)
	}
	
	// Foreign Key constraints
	fkQuery := `
		SELECT 
			kcu.constraint_name,
			kcu.column_name,
			kcu.referenced_table_name,
			kcu.referenced_column_name
		FROM information_schema.key_column_usage kcu
		WHERE kcu.table_schema = ? 
		AND kcu.table_name = ?
		AND kcu.referenced_table_name IS NOT NULL
		ORDER BY kcu.constraint_name, kcu.ordinal_position`
	
	fkRows, err := r.db.QueryContext(ctx, fkQuery, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()
	
	fkMap := make(map[string]*models.Constraint)
	
	for fkRows.Next() {
		var constraintName, columnName, refTable, refColumn string
		
		if err := fkRows.Scan(&constraintName, &columnName, &refTable, &refColumn); err != nil {
			return nil, err
		}
		
		if _, exists := fkMap[constraintName]; !exists {
			fkMap[constraintName] = &models.Constraint{
				Name:             constraintName,
				Type:             models.ForeignKey,
				Columns:          []string{},
				ReferencedTable:  refTable,
				ReferencedColumn: []string{},
			}
		}
		
		// Check if column already exists to avoid duplicates
		columnExists := false
		for _, col := range fkMap[constraintName].Columns {
			if col == columnName {
				columnExists = true
				break
			}
		}
		if !columnExists {
			fkMap[constraintName].Columns = append(fkMap[constraintName].Columns, columnName)
		}

		// Check if referenced column already exists to avoid duplicates
		refColumnExists := false
		for _, col := range fkMap[constraintName].ReferencedColumn {
			if col == refColumn {
				refColumnExists = true
				break
			}
		}
		if !refColumnExists {
			fkMap[constraintName].ReferencedColumn = append(fkMap[constraintName].ReferencedColumn, refColumn)
		}
	}
	
	for _, fk := range fkMap {
		constraints = append(constraints, *fk)
	}
	
	// Check constraints (MySQL 8.0.16+)
	checkQuery := `
		SELECT 
			tc.constraint_name,
			cc.check_clause
		FROM information_schema.table_constraints tc
		JOIN information_schema.check_constraints cc 
			ON tc.constraint_name = cc.constraint_name 
			AND tc.constraint_schema = cc.constraint_schema
		WHERE tc.table_schema = ? AND tc.table_name = ?
		AND tc.constraint_type = 'CHECK'`
	
	checkRows, err := r.db.QueryContext(ctx, checkQuery, schemaName, tableName)
	if err == nil {
		defer checkRows.Close()
		
		for checkRows.Next() {
			var constraint models.Constraint
			
			if err := checkRows.Scan(&constraint.Name, &constraint.CheckExpression); err != nil {
				continue
			}
			
			constraint.Type = models.Check
			constraints = append(constraints, constraint)
		}
	}
	
	return constraints, nil
}

func (r *MySQLReader) getIndexes(ctx context.Context, schemaName, tableName string) ([]models.Index, error) {
	query := `
		SELECT DISTINCT
			index_name,
			non_unique,
			index_type
		FROM information_schema.statistics
		WHERE table_schema = ? AND table_name = ?
		AND index_name != 'PRIMARY'
		ORDER BY index_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	indexMap := make(map[string]*models.Index)
	
	for rows.Next() {
		var indexName, indexType string
		var nonUnique int
		
		if err := rows.Scan(&indexName, &nonUnique, &indexType); err != nil {
			return nil, err
		}
		
		indexMap[indexName] = &models.Index{
			Name:      indexName,
			TableName: tableName,
			IsUnique:  nonUnique == 0,
			Type:      indexType,
			Columns:   []string{},
		}
	}
	
	// Get columns for each index
	for name, index := range indexMap {
		colQuery := `
			SELECT column_name
			FROM information_schema.statistics
			WHERE table_schema = ? AND table_name = ? AND index_name = ?
			ORDER BY seq_in_index`
		
		colRows, err := r.db.QueryContext(ctx, colQuery, schemaName, tableName, name)
		if err != nil {
			return nil, err
		}
		
		for colRows.Next() {
			var columnName string
			if err := colRows.Scan(&columnName); err != nil {
				colRows.Close()
				return nil, err
			}
			index.Columns = append(index.Columns, columnName)
		}
		colRows.Close()
	}
	
	var indexes []models.Index
	for _, index := range indexMap {
		indexes = append(indexes, *index)
	}
	
	return indexes, nil
}

func (r *MySQLReader) getViews(ctx context.Context, schemaName string) ([]models.View, error) {
	query := `
		SELECT table_name, view_definition
		FROM information_schema.views
		WHERE table_schema = ?
		ORDER BY table_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var views []models.View
	for rows.Next() {
		var view models.View
		if err := rows.Scan(&view.Name, &view.Definition); err != nil {
			return nil, err
		}
		view.Schema = schemaName
		
		columns, err := r.getColumns(ctx, schemaName, view.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for view %s: %w", view.Name, err)
		}
		view.Columns = columns
		
		views = append(views, view)
	}
	
	return views, nil
}

func (r *MySQLReader) getFunctions(ctx context.Context, schemaName string) ([]models.Function, error) {
	query := `
		SELECT 
			routine_name,
			data_type,
			routine_definition
		FROM information_schema.routines
		WHERE routine_schema = ? AND routine_type = 'FUNCTION'
		ORDER BY routine_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var functions []models.Function
	for rows.Next() {
		var fn models.Function
		var body sql.NullString
		
		if err := rows.Scan(&fn.Name, &fn.ReturnType, &body); err != nil {
			return nil, err
		}
		
		fn.Schema = schemaName
		if body.Valid {
			fn.Body = body.String
		}
		
		functions = append(functions, fn)
	}
	
	return functions, nil
}

func (r *MySQLReader) getProcedures(ctx context.Context, schemaName string) ([]models.Procedure, error) {
	query := `
		SELECT 
			routine_name,
			routine_definition
		FROM information_schema.routines
		WHERE routine_schema = ? AND routine_type = 'PROCEDURE'
		ORDER BY routine_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var procedures []models.Procedure
	for rows.Next() {
		var proc models.Procedure
		var body sql.NullString
		
		if err := rows.Scan(&proc.Name, &body); err != nil {
			return nil, err
		}
		
		proc.Schema = schemaName
		if body.Valid {
			proc.Body = body.String
		}
		
		procedures = append(procedures, proc)
	}
	
	return procedures, nil
}

func (r *MySQLReader) getTriggers(ctx context.Context, schemaName string) ([]models.Trigger, error) {
	query := `
		SELECT 
			trigger_name,
			event_object_table,
			action_timing,
			event_manipulation,
			action_statement
		FROM information_schema.triggers
		WHERE trigger_schema = ?
		ORDER BY trigger_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var triggers []models.Trigger
	for rows.Next() {
		var trigger models.Trigger
		var timing, event string
		
		if err := rows.Scan(&trigger.Name, &trigger.TableName, &timing, &event, &trigger.Body); err != nil {
			return nil, err
		}
		
		trigger.Schema = schemaName
		
		switch strings.ToUpper(timing) {
		case "BEFORE":
			trigger.Timing = models.Before
		case "AFTER":
			trigger.Timing = models.After
		}
		
		switch strings.ToUpper(event) {
		case "INSERT":
			trigger.Event = models.Insert
		case "UPDATE":
			trigger.Event = models.Update
		case "DELETE":
			trigger.Event = models.Delete
		}
		
		triggers = append(triggers, trigger)
	}
	
	return triggers, nil
}

func (r *MySQLReader) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}