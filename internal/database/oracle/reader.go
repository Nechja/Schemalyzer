package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	_ "github.com/sijms/go-ora/v2"
	"github.com/nechja/schemalyzer/pkg/models"
	"strings"
)

type OracleReader struct {
	db *sql.DB
}

func NewOracleReader() *OracleReader {
	return &OracleReader{}
}

func (r *OracleReader) Connect(ctx context.Context, connectionString string) error {
	db, err := sql.Open("oracle", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to oracle: %w", err)
	}
	
	// Configure connection pool for large databases
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No timeout
	
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping oracle: %w", err)
	}
	
	r.db = db
	return nil
}

func (r *OracleReader) GetSchema(ctx context.Context, schemaName string) (*models.Schema, error) {
	schema := &models.Schema{
		Name:         schemaName,
		DatabaseType: models.Oracle,
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

func (r *OracleReader) ListSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT username 
		FROM all_users
		WHERE username NOT IN ('SYS', 'SYSTEM', 'DBSNMP', 'SYSMAN', 'OUTLN', 'FLOWS_FILES', 'MDSYS', 
			'ORDSYS', 'EXFSYS', 'WMSYS', 'APPQOSSYS', 'APEX_030200', 'OWBSYS_AUDIT', 
			'ORDDATA', 'CTXSYS', 'ANONYMOUS', 'XDB', 'ORDPLUGINS', 'SI_INFORMTN_SCHEMA',
			'OLAPSYS', 'ORACLE_OCM', 'XS$NULL', 'BI', 'PM', 'MDDATA', 'IX', 'SH', 'DIP')
		ORDER BY username`
	
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

func (r *OracleReader) getTables(ctx context.Context, schemaName string) ([]models.Table, error) {
	query := `
		SELECT t.table_name, c.comments
		FROM all_tables t
		LEFT JOIN all_tab_comments c ON t.owner = c.owner AND t.table_name = c.table_name
		WHERE t.owner = :1 AND t.temporary = 'N'
		ORDER BY t.table_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName))
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

func (r *OracleReader) getColumns(ctx context.Context, schemaName, tableName string) ([]models.Column, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type || 
			CASE 
				WHEN c.data_type IN ('VARCHAR2', 'CHAR', 'NVARCHAR2', 'NCHAR') THEN '(' || c.char_length || ')'
				WHEN c.data_type = 'NUMBER' AND c.data_precision IS NOT NULL THEN 
					'(' || c.data_precision || 
					CASE WHEN c.data_scale > 0 THEN ',' || c.data_scale ELSE '' END || ')'
				ELSE ''
			END AS data_type,
			c.nullable,
			c.data_default,
			c.column_id,
			cc.comments
		FROM all_tab_columns c
		LEFT JOIN all_col_comments cc ON c.owner = cc.owner AND c.table_name = cc.table_name AND c.column_name = cc.column_name
		WHERE c.owner = :1 AND c.table_name = :2
		ORDER BY c.column_id`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName), strings.ToUpper(tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var nullable string
		var defaultValue, comment sql.NullString
		
		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &defaultValue, &col.Position, &comment); err != nil {
			return nil, err
		}
		
		col.IsNullable = nullable == "Y"
		if defaultValue.Valid {
			def := strings.TrimSpace(defaultValue.String)
			col.DefaultValue = &def
		}
		if comment.Valid {
			col.Comment = comment.String
		}
		
		columns = append(columns, col)
	}
	
	return columns, nil
}

func (r *OracleReader) getConstraints(ctx context.Context, schemaName, tableName string) ([]models.Constraint, error) {
	query := `
		SELECT 
			c.constraint_name,
			c.constraint_type,
			c.search_condition
		FROM all_constraints c
		WHERE c.owner = :1 AND c.table_name = :2
		AND c.constraint_type IN ('P', 'U', 'R', 'C')
		ORDER BY c.constraint_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName), strings.ToUpper(tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	constraintMap := make(map[string]*models.Constraint)
	
	for rows.Next() {
		var constraintName, constraintType string
		var searchCondition sql.NullString
		
		if err := rows.Scan(&constraintName, &constraintType, &searchCondition); err != nil {
			return nil, err
		}
		
		constraint := &models.Constraint{
			Name:    constraintName,
			Columns: []string{},
		}
		
		switch constraintType {
		case "P":
			constraint.Type = models.PrimaryKey
		case "U":
			constraint.Type = models.Unique
		case "R":
			constraint.Type = models.ForeignKey
		case "C":
			constraint.Type = models.Check
			if searchCondition.Valid {
				constraint.CheckExpression = searchCondition.String
			}
		}
		
		constraintMap[constraintName] = constraint
	}
	
	// Get columns for each constraint
	for name, constraint := range constraintMap {
		colQuery := `
			SELECT column_name
			FROM all_cons_columns
			WHERE owner = :1 AND constraint_name = :2
			ORDER BY position`
		
		colRows, err := r.db.QueryContext(ctx, colQuery, strings.ToUpper(schemaName), name)
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
	}
	
	// Get foreign key references
	for name, constraint := range constraintMap {
		if constraint.Type == models.ForeignKey {
			refQuery := `
				SELECT 
					r.table_name,
					rc.column_name
				FROM all_constraints c
				JOIN all_constraints r ON c.r_constraint_name = r.constraint_name AND c.r_owner = r.owner
				JOIN all_cons_columns rc ON r.constraint_name = rc.constraint_name AND r.owner = rc.owner
				WHERE c.owner = :1 AND c.constraint_name = :2
				ORDER BY rc.position`
			
			refRows, err := r.db.QueryContext(ctx, refQuery, strings.ToUpper(schemaName), name)
			if err != nil {
				return nil, err
			}
			
			var refTable string
			for refRows.Next() {
				var refColumn string
				if err := refRows.Scan(&refTable, &refColumn); err != nil {
					refRows.Close()
					return nil, err
				}
				if constraint.ReferencedTable == "" {
					constraint.ReferencedTable = refTable
				}
				constraint.ReferencedColumn = append(constraint.ReferencedColumn, refColumn)
			}
			refRows.Close()
		}
	}
	
	var constraints []models.Constraint
	for _, constraint := range constraintMap {
		constraints = append(constraints, *constraint)
	}
	
	return constraints, nil
}

func (r *OracleReader) getIndexes(ctx context.Context, schemaName, tableName string) ([]models.Index, error) {
	query := `
		SELECT DISTINCT
			i.index_name,
			i.uniqueness,
			i.index_type
		FROM all_indexes i
		WHERE i.owner = :1 AND i.table_name = :2
		AND NOT EXISTS (
			SELECT 1 FROM all_constraints c 
			WHERE c.owner = i.owner 
			AND c.constraint_name = i.index_name 
			AND c.constraint_type = 'P'
		)
		ORDER BY i.index_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName), strings.ToUpper(tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	indexMap := make(map[string]*models.Index)
	
	for rows.Next() {
		var indexName, uniqueness, indexType string
		
		if err := rows.Scan(&indexName, &uniqueness, &indexType); err != nil {
			return nil, err
		}
		
		indexMap[indexName] = &models.Index{
			Name:      indexName,
			TableName: tableName,
			IsUnique:  uniqueness == "UNIQUE",
			Type:      indexType,
			Columns:   []string{},
		}
	}
	
	// Get columns for each index
	for name, index := range indexMap {
		colQuery := `
			SELECT column_name
			FROM all_ind_columns
			WHERE index_owner = :1 AND index_name = :2
			ORDER BY column_position`
		
		colRows, err := r.db.QueryContext(ctx, colQuery, strings.ToUpper(schemaName), name)
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

func (r *OracleReader) getViews(ctx context.Context, schemaName string) ([]models.View, error) {
	query := `
		SELECT view_name, text
		FROM all_views
		WHERE owner = :1
		ORDER BY view_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName))
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
		
		columns, err := r.getViewColumns(ctx, schemaName, view.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for view %s: %w", view.Name, err)
		}
		view.Columns = columns
		
		views = append(views, view)
	}
	
	return views, nil
}

func (r *OracleReader) getViewColumns(ctx context.Context, schemaName, viewName string) ([]models.Column, error) {
	query := `
		SELECT 
			column_name,
			data_type || 
			CASE 
				WHEN data_type IN ('VARCHAR2', 'CHAR', 'NVARCHAR2', 'NCHAR') THEN '(' || data_length || ')'
				WHEN data_type = 'NUMBER' AND data_precision IS NOT NULL THEN 
					'(' || data_precision || 
					CASE WHEN data_scale > 0 THEN ',' || data_scale ELSE '' END || ')'
				ELSE ''
			END AS data_type,
			nullable,
			column_id
		FROM all_tab_columns
		WHERE owner = :1 AND table_name = :2
		ORDER BY column_id`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName), strings.ToUpper(viewName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var nullable string
		
		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &col.Position); err != nil {
			return nil, err
		}
		
		col.IsNullable = nullable == "Y"
		columns = append(columns, col)
	}
	
	return columns, nil
}

func (r *OracleReader) getSequences(ctx context.Context, schemaName string) ([]models.Sequence, error) {
	query := `
		SELECT 
			sequence_name,
			min_value,
			max_value,
			increment_by,
			cycle_flag,
			last_number
		FROM all_sequences
		WHERE sequence_owner = :1
		ORDER BY sequence_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sequences []models.Sequence
	for rows.Next() {
		var seq models.Sequence
		var cycleFlag string
		var minValueStr, maxValueStr string
		
		if err := rows.Scan(&seq.Name, &minValueStr, &maxValueStr, 
			&seq.Increment, &cycleFlag, &seq.CurrentValue); err != nil {
			return nil, err
		}
		
		// Handle Oracle's large numeric values
		// If the value is the Oracle default max (28 9's), set to 0 to indicate no limit
		if maxValueStr == "9999999999999999999999999999" {
			seq.MaxValue = 0
		} else {
			// Try to parse as int64
			if val, err := strconv.ParseInt(maxValueStr, 10, 64); err == nil {
				seq.MaxValue = val
			} else {
				seq.MaxValue = 0 // Too large for int64
			}
		}
		
		if val, err := strconv.ParseInt(minValueStr, 10, 64); err == nil {
			seq.MinValue = val
		} else {
			seq.MinValue = 0
		}
		
		seq.Schema = schemaName
		seq.IsCyclic = cycleFlag == "Y"
		sequences = append(sequences, seq)
	}
	
	return sequences, nil
}

func (r *OracleReader) getFunctions(ctx context.Context, schemaName string) ([]models.Function, error) {
	query := `
		SELECT 
			object_name,
			DBMS_METADATA.GET_DDL('FUNCTION', object_name, owner) AS ddl
		FROM all_objects
		WHERE owner = :1 AND object_type = 'FUNCTION'
		ORDER BY object_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var functions []models.Function
	for rows.Next() {
		var fn models.Function
		var ddl string
		
		if err := rows.Scan(&fn.Name, &ddl); err != nil {
			return nil, err
		}
		
		fn.Schema = schemaName
		fn.Body = ddl
		functions = append(functions, fn)
	}
	
	return functions, nil
}

func (r *OracleReader) getProcedures(ctx context.Context, schemaName string) ([]models.Procedure, error) {
	query := `
		SELECT 
			object_name,
			DBMS_METADATA.GET_DDL('PROCEDURE', object_name, owner) AS ddl
		FROM all_objects
		WHERE owner = :1 AND object_type = 'PROCEDURE'
		ORDER BY object_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var procedures []models.Procedure
	for rows.Next() {
		var proc models.Procedure
		var ddl string
		
		if err := rows.Scan(&proc.Name, &ddl); err != nil {
			return nil, err
		}
		
		proc.Schema = schemaName
		proc.Body = ddl
		procedures = append(procedures, proc)
	}
	
	return procedures, nil
}

func (r *OracleReader) getTriggers(ctx context.Context, schemaName string) ([]models.Trigger, error) {
	query := `
		SELECT 
			trigger_name,
			table_name,
			trigger_type,
			triggering_event,
			trigger_body
		FROM all_triggers
		WHERE owner = :1
		ORDER BY trigger_name`
	
	rows, err := r.db.QueryContext(ctx, query, strings.ToUpper(schemaName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var triggers []models.Trigger
	for rows.Next() {
		var trigger models.Trigger
		var triggerType, event string
		
		if err := rows.Scan(&trigger.Name, &trigger.TableName, &triggerType, &event, &trigger.Body); err != nil {
			return nil, err
		}
		
		trigger.Schema = schemaName
		
		if strings.Contains(triggerType, "BEFORE") {
			trigger.Timing = models.Before
		} else {
			trigger.Timing = models.After
		}
		
		if strings.Contains(event, "INSERT") {
			trigger.Event = models.Insert
		} else if strings.Contains(event, "UPDATE") {
			trigger.Event = models.Update
		} else if strings.Contains(event, "DELETE") {
			trigger.Event = models.Delete
		}
		
		triggers = append(triggers, trigger)
	}
	
	return triggers, nil
}

func (r *OracleReader) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}