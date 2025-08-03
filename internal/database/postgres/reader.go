package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	_ "github.com/lib/pq"
	"github.com/nechja/schemalyzer/pkg/models"
)

type PostgresReader struct {
	db *sql.DB
}

func NewPostgresReader() *PostgresReader {
	return &PostgresReader{}
}

func (r *PostgresReader) Connect(ctx context.Context, connectionString string) error {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	
	// Configure connection pool for large databases
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No timeout
	
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}
	
	r.db = db
	return nil
}

func (r *PostgresReader) GetSchema(ctx context.Context, schemaName string) (*models.Schema, error) {
	schema := &models.Schema{
		Name:         schemaName,
		DatabaseType: models.PostgreSQL,
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
	
	// Fetch tables in parallel
	wg.Add(6)
	
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
	
	go func() {
		defer wg.Done()
		sequences, err := r.getSequences(ctx, schemaName)
		if err != nil {
			res.err = fmt.Errorf("failed to get sequences: %w", err)
			return
		}
		res.sequences = sequences
	}()
	
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

func (r *PostgresReader) ListSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
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

func (r *PostgresReader) getTables(ctx context.Context, schemaName string) ([]models.Table, error) {
	query := `
		SELECT table_name, obj_description(pgc.oid)
		FROM information_schema.tables t
		JOIN pg_class pgc ON pgc.relname = t.table_name
		JOIN pg_namespace pgn ON pgn.oid = pgc.relnamespace AND pgn.nspname = t.table_schema
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tables []models.Table
	for rows.Next() {
		var table models.Table
		var comment sql.NullString
		
		if err := rows.Scan(&table.Name, &comment); err != nil {
			return nil, err
		}
		
		table.Schema = schemaName
		if comment.Valid {
			table.Comment = comment.String
		}
		
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

func (r *PostgresReader) getColumns(ctx context.Context, schemaName, tableName string) ([]models.Column, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.ordinal_position,
			col_description(pgc.oid, c.ordinal_position::int)
		FROM information_schema.columns c
		JOIN pg_class pgc ON pgc.relname = c.table_name
		JOIN pg_namespace pgn ON pgn.oid = pgc.relnamespace AND pgn.nspname = c.table_schema
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var isNullable string
		var defaultValue, comment sql.NullString
		
		if err := rows.Scan(&col.Name, &col.DataType, &isNullable, &defaultValue, &col.Position, &comment); err != nil {
			return nil, err
		}
		
		col.IsNullable = isNullable == "YES"
		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}
		if comment.Valid {
			col.Comment = comment.String
		}
		
		columns = append(columns, col)
	}
	
	return columns, nil
}

func (r *PostgresReader) getConstraints(ctx context.Context, schemaName, tableName string) ([]models.Constraint, error) {
	query := `
		SELECT 
			tc.constraint_name,
			tc.constraint_type,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			cc.check_clause
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		LEFT JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		LEFT JOIN information_schema.check_constraints cc
			ON cc.constraint_name = tc.constraint_name
			AND cc.constraint_schema = tc.table_schema
		WHERE tc.table_schema = $1 AND tc.table_name = $2
		ORDER BY tc.constraint_name, kcu.ordinal_position`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	constraintMap := make(map[string]*models.Constraint)
	
	for rows.Next() {
		var constraintName, constraintType string
		var columnName, foreignTable, foreignColumn, checkClause sql.NullString
		
		if err := rows.Scan(&constraintName, &constraintType, &columnName, &foreignTable, &foreignColumn, &checkClause); err != nil {
			return nil, err
		}
		
		if _, exists := constraintMap[constraintName]; !exists {
			constraint := &models.Constraint{
				Name:    constraintName,
				Columns: []string{},
			}
			
			switch constraintType {
			case "PRIMARY KEY":
				constraint.Type = models.PrimaryKey
			case "FOREIGN KEY":
				constraint.Type = models.ForeignKey
				if foreignTable.Valid {
					constraint.ReferencedTable = foreignTable.String
				}
			case "UNIQUE":
				constraint.Type = models.Unique
			case "CHECK":
				constraint.Type = models.Check
				if checkClause.Valid {
					constraint.CheckExpression = checkClause.String
				}
			}
			
			constraintMap[constraintName] = constraint
		}
		
		if columnName.Valid {
			constraintMap[constraintName].Columns = append(constraintMap[constraintName].Columns, columnName.String)
		}
		
		if constraintMap[constraintName].Type == models.ForeignKey && foreignColumn.Valid {
			constraintMap[constraintName].ReferencedColumn = append(constraintMap[constraintName].ReferencedColumn, foreignColumn.String)
		}
	}
	
	var constraints []models.Constraint
	for _, constraint := range constraintMap {
		constraints = append(constraints, *constraint)
	}
	
	return constraints, nil
}

func (r *PostgresReader) getIndexes(ctx context.Context, schemaName, tableName string) ([]models.Index, error) {
	query := `
		SELECT 
			i.relname AS index_name,
			idx.indisunique,
			am.amname AS index_type,
			array_agg(a.attname ORDER BY array_position(idx.indkey, a.attnum)) AS column_names
		FROM pg_index idx
		JOIN pg_class t ON t.oid = idx.indrelid
		JOIN pg_class i ON i.oid = idx.indexrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_am am ON am.oid = i.relam
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(idx.indkey)
		WHERE n.nspname = $1 AND t.relname = $2 AND NOT idx.indisprimary
		GROUP BY i.relname, idx.indisunique, am.amname`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var indexes []models.Index
	for rows.Next() {
		var index models.Index
		var columnNames []string
		
		if err := rows.Scan(&index.Name, &index.IsUnique, &index.Type, &columnNames); err != nil {
			return nil, err
		}
		
		index.TableName = tableName
		index.Columns = columnNames
		indexes = append(indexes, index)
	}
	
	return indexes, nil
}

func (r *PostgresReader) getViews(ctx context.Context, schemaName string) ([]models.View, error) {
	query := `
		SELECT table_name, view_definition
		FROM information_schema.views
		WHERE table_schema = $1
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

func (r *PostgresReader) getSequences(ctx context.Context, schemaName string) ([]models.Sequence, error) {
	query := `
		SELECT 
			sequence_name,
			start_value,
			increment,
			minimum_value,
			maximum_value,
			cycle_option
		FROM information_schema.sequences
		WHERE sequence_schema = $1
		ORDER BY sequence_name`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sequences []models.Sequence
	for rows.Next() {
		var seq models.Sequence
		var cycleOption string
		
		if err := rows.Scan(&seq.Name, &seq.StartValue, &seq.Increment, 
			&seq.MinValue, &seq.MaxValue, &cycleOption); err != nil {
			return nil, err
		}
		
		seq.Schema = schemaName
		seq.IsCyclic = cycleOption == "YES"
		sequences = append(sequences, seq)
	}
	
	return sequences, nil
}

func (r *PostgresReader) getFunctions(ctx context.Context, schemaName string) ([]models.Function, error) {
	query := `
		SELECT 
			p.proname AS function_name,
			pg_get_function_result(p.oid) AS return_type,
			pg_get_functiondef(p.oid) AS function_body
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1 AND p.prokind = 'f'
		ORDER BY p.proname`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var functions []models.Function
	for rows.Next() {
		var fn models.Function
		if err := rows.Scan(&fn.Name, &fn.ReturnType, &fn.Body); err != nil {
			return nil, err
		}
		fn.Schema = schemaName
		functions = append(functions, fn)
	}
	
	return functions, nil
}

func (r *PostgresReader) getProcedures(ctx context.Context, schemaName string) ([]models.Procedure, error) {
	query := `
		SELECT 
			p.proname AS procedure_name,
			pg_get_functiondef(p.oid) AS procedure_body
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1 AND p.prokind = 'p'
		ORDER BY p.proname`
	
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var procedures []models.Procedure
	for rows.Next() {
		var proc models.Procedure
		if err := rows.Scan(&proc.Name, &proc.Body); err != nil {
			return nil, err
		}
		proc.Schema = schemaName
		procedures = append(procedures, proc)
	}
	
	return procedures, nil
}

func (r *PostgresReader) getTriggers(ctx context.Context, schemaName string) ([]models.Trigger, error) {
	query := `
		SELECT 
			t.tgname AS trigger_name,
			c.relname AS table_name,
			CASE 
				WHEN t.tgtype & 2 = 2 THEN 'BEFORE'
				ELSE 'AFTER'
			END AS timing,
			CASE 
				WHEN t.tgtype & 4 = 4 THEN 'INSERT'
				WHEN t.tgtype & 8 = 8 THEN 'UPDATE'
				WHEN t.tgtype & 16 = 16 THEN 'DELETE'
			END AS event,
			pg_get_triggerdef(t.oid) AS trigger_def
		FROM pg_trigger t
		JOIN pg_class c ON c.oid = t.tgrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND NOT t.tgisinternal
		ORDER BY t.tgname`
	
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
		
		switch timing {
		case "BEFORE":
			trigger.Timing = models.Before
		case "AFTER":
			trigger.Timing = models.After
		}
		
		switch event {
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

func (r *PostgresReader) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}