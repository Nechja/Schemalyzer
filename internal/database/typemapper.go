package database

import (
	"strings"
	"regexp"
	"github.com/nechja/schemalyzer/pkg/models"
)

type TypeMapper struct {
	mappings map[string]map[string]TypeEquivalence
}

type TypeEquivalence struct {
	TargetType   string
	IsCompatible bool
	Warning      string
}

func NewTypeMapper() *TypeMapper {
	tm := &TypeMapper{
		mappings: make(map[string]map[string]TypeEquivalence),
	}
	tm.initializeMappings()
	return tm
}

func (tm *TypeMapper) initializeMappings() {
	// PostgreSQL to MySQL mappings
	tm.mappings["postgresql-mysql"] = map[string]TypeEquivalence{
		// Numeric types
		"smallint":       {"smallint", true, ""},
		"integer":        {"int", true, ""},
		"bigint":         {"bigint", true, ""},
		"decimal":        {"decimal", true, ""},
		"numeric":        {"decimal", true, ""},
		"real":           {"float", true, ""},
		"double precision": {"double", true, ""},
		"serial":         {"int AUTO_INCREMENT", true, ""},
		"bigserial":      {"bigint AUTO_INCREMENT", true, ""},
		
		// String types
		"varchar":        {"varchar", true, ""},
		"char":           {"char", true, ""},
		"text":           {"text", true, ""},
		
		// Date/Time types
		"timestamp":      {"timestamp", true, ""},
		"date":           {"date", true, ""},
		"time":           {"time", true, ""},
		"interval":       {"varchar(50)", false, "INTERVAL type not supported in MySQL, using VARCHAR"},
		
		// Boolean
		"boolean":        {"tinyint(1)", true, "Boolean mapped to tinyint(1)"},
		
		// Binary
		"bytea":          {"blob", true, ""},
		
		// JSON
		"json":           {"json", true, ""},
		"jsonb":          {"json", true, "JSONB performance benefits lost in MySQL"},
		
		// UUID
		"uuid":           {"varchar(36)", true, "UUID stored as VARCHAR(36)"},
		
		// Arrays
		"array":          {"json", false, "Array types not supported in MySQL, consider using JSON"},
		
		// Network types
		"inet":           {"varchar(45)", false, "Network types not supported, using VARCHAR"},
		"cidr":           {"varchar(45)", false, "Network types not supported, using VARCHAR"},
	}
	
	// PostgreSQL to Oracle mappings  
	tm.mappings["postgresql-oracle"] = map[string]TypeEquivalence{
		// Numeric types
		"smallint":       {"NUMBER(5)", true, ""},
		"integer":        {"NUMBER(10)", true, ""},
		"bigint":         {"NUMBER(19)", true, ""},
		"decimal":        {"NUMBER", true, ""},
		"numeric":        {"NUMBER", true, ""},
		"real":           {"BINARY_FLOAT", true, ""},
		"double precision": {"BINARY_DOUBLE", true, ""},
		"serial":         {"NUMBER", false, "Use SEQUENCE with trigger for auto-increment"},
		"bigserial":      {"NUMBER", false, "Use SEQUENCE with trigger for auto-increment"},
		
		// String types
		"varchar":        {"VARCHAR2", true, ""},
		"char":           {"CHAR", true, ""},
		"text":           {"CLOB", true, "Large text stored as CLOB"},
		
		// Date/Time types
		"timestamp":      {"TIMESTAMP", true, ""},
		"date":           {"DATE", true, ""},
		"time":           {"TIMESTAMP", true, "TIME stored as TIMESTAMP"},
		"interval":       {"INTERVAL DAY TO SECOND", true, ""},
		
		// Boolean
		"boolean":        {"NUMBER(1)", true, "Boolean mapped to NUMBER(1)"},
		
		// Binary
		"bytea":          {"BLOB", true, ""},
		
		// JSON
		"json":           {"CLOB", false, "JSON stored as CLOB, use JSON functions"},
		"jsonb":          {"CLOB", false, "JSONB stored as CLOB, use JSON functions"},
		
		// UUID
		"uuid":           {"VARCHAR2(36)", true, "UUID stored as VARCHAR2(36)"},
		
		// Arrays
		"array":          {"", false, "Array types not supported in Oracle, use nested tables"},
	}
	
	// MySQL to PostgreSQL mappings
	tm.mappings["mysql-postgresql"] = map[string]TypeEquivalence{
		// Numeric types
		"tinyint":        {"smallint", true, ""},
		"smallint":       {"smallint", true, ""},
		"mediumint":      {"integer", true, ""},
		"int":            {"integer", true, ""},
		"bigint":         {"bigint", true, ""},
		"decimal":        {"decimal", true, ""},
		"float":          {"real", true, ""},
		"double":         {"double precision", true, ""},
		
		// String types
		"varchar":        {"varchar", true, ""},
		"char":           {"char", true, ""},
		"tinytext":       {"text", true, ""},
		"text":           {"text", true, ""},
		"mediumtext":     {"text", true, ""},
		"longtext":       {"text", true, ""},
		
		// Date/Time types
		"datetime":       {"timestamp", true, ""},
		"timestamp":      {"timestamp", true, ""},
		"date":           {"date", true, ""},
		"time":           {"time", true, ""},
		"year":           {"integer", false, "YEAR type mapped to INTEGER"},
		
		// Binary
		"tinyblob":       {"bytea", true, ""},
		"blob":           {"bytea", true, ""},
		"mediumblob":     {"bytea", true, ""},
		"longblob":       {"bytea", true, ""},
		"binary":         {"bytea", true, ""},
		"varbinary":      {"bytea", true, ""},
		
		// JSON
		"json":           {"json", true, ""},
		
		// Special
		"enum":           {"varchar", false, "ENUM mapped to VARCHAR, consider CHECK constraint"},
		"set":            {"varchar", false, "SET mapped to VARCHAR, consider array or separate table"},
	}
	
	// MySQL to Oracle mappings
	tm.mappings["mysql-oracle"] = map[string]TypeEquivalence{
		// Numeric types
		"tinyint":        {"NUMBER(3)", true, ""},
		"smallint":       {"NUMBER(5)", true, ""},
		"mediumint":      {"NUMBER(7)", true, ""},
		"int":            {"NUMBER(10)", true, ""},
		"bigint":         {"NUMBER(19)", true, ""},
		"decimal":        {"NUMBER", true, ""},
		"float":          {"BINARY_FLOAT", true, ""},
		"double":         {"BINARY_DOUBLE", true, ""},
		
		// String types
		"varchar":        {"VARCHAR2", true, ""},
		"char":           {"CHAR", true, ""},
		"tinytext":       {"VARCHAR2(4000)", true, ""},
		"text":           {"CLOB", true, ""},
		"mediumtext":     {"CLOB", true, ""},
		"longtext":       {"CLOB", true, ""},
		
		// Date/Time types
		"datetime":       {"TIMESTAMP", true, ""},
		"timestamp":      {"TIMESTAMP", true, ""},
		"date":           {"DATE", true, ""},
		"time":           {"TIMESTAMP", true, "TIME stored as TIMESTAMP"},
		"year":           {"NUMBER(4)", false, "YEAR type mapped to NUMBER(4)"},
		
		// Binary
		"tinyblob":       {"BLOB", true, ""},
		"blob":           {"BLOB", true, ""},
		"mediumblob":     {"BLOB", true, ""},
		"longblob":       {"BLOB", true, ""},
		"binary":         {"RAW", true, ""},
		"varbinary":      {"RAW", true, ""},
		
		// JSON
		"json":           {"CLOB", false, "JSON stored as CLOB, use JSON functions"},
		
		// Special
		"enum":           {"VARCHAR2", false, "ENUM mapped to VARCHAR2, add CHECK constraint"},
		"set":            {"VARCHAR2", false, "SET mapped to VARCHAR2, consider separate table"},
	}
	
	// Oracle to PostgreSQL mappings
	tm.mappings["oracle-postgresql"] = map[string]TypeEquivalence{
		// Numeric types
		"NUMBER":         {"numeric", true, ""},
		"BINARY_FLOAT":   {"real", true, ""},
		"BINARY_DOUBLE":  {"double precision", true, ""},
		
		// String types
		"VARCHAR2":       {"varchar", true, ""},
		"CHAR":           {"char", true, ""},
		"CLOB":           {"text", true, ""},
		"NCHAR":          {"char", true, ""},
		"NVARCHAR2":      {"varchar", true, ""},
		"NCLOB":          {"text", true, ""},
		
		// Date/Time types
		"DATE":           {"timestamp", true, "Oracle DATE includes time"},
		"TIMESTAMP":      {"timestamp", true, ""},
		"INTERVAL":       {"interval", true, ""},
		
		// Binary
		"BLOB":           {"bytea", true, ""},
		"RAW":            {"bytea", true, ""},
		"LONG RAW":       {"bytea", true, ""},
		
		// Special
		"ROWID":          {"varchar", false, "ROWID has no direct equivalent"},
		"XMLTYPE":        {"xml", true, ""},
	}
	
	// Oracle to MySQL mappings
	tm.mappings["oracle-mysql"] = map[string]TypeEquivalence{
		// Numeric types
		"NUMBER":         {"decimal", true, ""},
		"BINARY_FLOAT":   {"float", true, ""},
		"BINARY_DOUBLE":  {"double", true, ""},
		
		// String types
		"VARCHAR2":       {"varchar", true, ""},
		"CHAR":           {"char", true, ""},
		"CLOB":           {"longtext", true, ""},
		"NCHAR":          {"char", true, ""},
		"NVARCHAR2":      {"varchar", true, ""},
		"NCLOB":          {"longtext", true, ""},
		
		// Date/Time types
		"DATE":           {"datetime", true, "Oracle DATE includes time"},
		"TIMESTAMP":      {"timestamp", true, ""},
		"INTERVAL":       {"varchar(50)", false, "INTERVAL stored as VARCHAR"},
		
		// Binary
		"BLOB":           {"longblob", true, ""},
		"RAW":            {"varbinary", true, ""},
		"LONG RAW":       {"longblob", true, ""},
		
		// Special
		"ROWID":          {"varchar(18)", false, "ROWID has no direct equivalent"},
		"XMLTYPE":        {"text", false, "XMLTYPE stored as TEXT"},
	}
}

func (tm *TypeMapper) MapType(sourceDB, targetDB models.DatabaseType, sourceType string) (string, bool, string) {
	key := string(sourceDB) + "-" + string(targetDB)
	
	if mappings, ok := tm.mappings[key]; ok {
		// Normalize the source type
		normalizedType := tm.normalizeType(sourceType)
		
		// First try exact match
		if equiv, ok := mappings[normalizedType]; ok {
			return equiv.TargetType, equiv.IsCompatible, equiv.Warning
		}
		
		// Try to match base type (remove size specifications)
		baseType := tm.extractBaseType(normalizedType)
		if equiv, ok := mappings[baseType]; ok {
			// Append size if present in original
			targetType := equiv.TargetType
			if size := tm.extractSize(sourceType); size != "" && tm.supportsSize(equiv.TargetType) {
				targetType = equiv.TargetType + "(" + size + ")"
			}
			return targetType, equiv.IsCompatible, equiv.Warning
		}
		
		// Check for array types
		if strings.Contains(normalizedType, "[]") || strings.Contains(normalizedType, " array") {
			if equiv, ok := mappings["array"]; ok {
				return equiv.TargetType, equiv.IsCompatible, equiv.Warning
			}
		}
	}
	
	// No mapping found, return as-is with warning
	return sourceType, false, "No type mapping found"
}

func (tm *TypeMapper) normalizeType(dataType string) string {
	// Convert to lowercase and trim
	normalized := strings.ToLower(strings.TrimSpace(dataType))
	
	// Handle PostgreSQL special cases
	normalized = strings.Replace(normalized, "character varying", "varchar", 1)
	normalized = strings.Replace(normalized, "character", "char", 1)
	normalized = strings.Replace(normalized, "int4", "integer", 1)
	normalized = strings.Replace(normalized, "int8", "bigint", 1)
	normalized = strings.Replace(normalized, "float8", "double precision", 1)
	normalized = strings.Replace(normalized, "bool", "boolean", 1)
	
	// Handle Oracle special cases
	normalized = strings.Replace(normalized, "varchar2", "varchar", 1)
	
	return normalized
}

func (tm *TypeMapper) extractBaseType(dataType string) string {
	// Remove size specifications and array indicators
	re := regexp.MustCompile(`^([a-zA-Z_]+)`)
	matches := re.FindStringSubmatch(dataType)
	if len(matches) > 1 {
		return matches[1]
	}
	return dataType
}

func (tm *TypeMapper) extractSize(dataType string) string {
	// Extract size specification from type
	re := regexp.MustCompile(`\(([^)]+)\)`)
	matches := re.FindStringSubmatch(dataType)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (tm *TypeMapper) supportsSize(targetType string) bool {
	// Check if target type supports size specification
	sizableTypes := []string{"varchar", "char", "decimal", "numeric", "varchar2"}
	baseType := strings.ToLower(tm.extractBaseType(targetType))
	
	for _, st := range sizableTypes {
		if baseType == st {
			return true
		}
	}
	return false
}

func (tm *TypeMapper) GetCompatibilityIssues(source, target *models.Schema) []models.CompatibilityIssue {
	var issues []models.CompatibilityIssue
	
	sourceDB := source.DatabaseType
	targetDB := target.DatabaseType
	
	// Check for database-specific features
	if sourceDB == models.PostgreSQL && targetDB != models.PostgreSQL {
		// PostgreSQL-specific features
		for _, table := range source.Tables {
			for _, col := range table.Columns {
				if strings.Contains(strings.ToLower(col.DataType), "array") ||
					strings.Contains(col.DataType, "[]") {
					issues = append(issues, models.CompatibilityIssue{
						Level:        models.Incompatible,
						ObjectType:   "Column",
						ObjectName:   table.Name + "." + col.Name,
						SourceDetail: col.DataType,
						Description:  "PostgreSQL array types are not supported in " + string(targetDB),
						Suggestion:   "Consider using JSON type or a separate junction table",
					})
				}
			}
		}
		
		// Check for sequences (PostgreSQL SERIAL)
		if len(source.Sequences) > 0 && targetDB == models.Oracle {
			issues = append(issues, models.CompatibilityIssue{
				Level:        models.Warning,
				ObjectType:   "Sequences",
				ObjectName:   "Multiple sequences",
				Description:  "PostgreSQL SERIAL columns use sequences, Oracle requires explicit sequence + trigger",
				Suggestion:   "Create sequences and triggers for auto-increment functionality",
			})
		}
	}
	
	// Check for Oracle-specific features
	if sourceDB == models.Oracle && targetDB != models.Oracle {
		// Oracle packages, procedures
		if len(source.Procedures) > 0 || len(source.Functions) > 0 {
			issues = append(issues, models.CompatibilityIssue{
				Level:        models.Warning,
				ObjectType:   "Stored Procedures",
				ObjectName:   "Multiple procedures/functions",
				Description:  "Oracle PL/SQL procedures may need syntax adjustments",
				Suggestion:   "Review and adjust procedure syntax for target database",
			})
		}
	}
	
	// Check for MySQL-specific features
	if sourceDB == models.MySQL && targetDB != models.MySQL {
		for _, table := range source.Tables {
			for _, col := range table.Columns {
				dataType := strings.ToLower(col.DataType)
				if strings.Contains(dataType, "enum") {
					issues = append(issues, models.CompatibilityIssue{
						Level:        models.Warning,
						ObjectType:   "Column",
						ObjectName:   table.Name + "." + col.Name,
						SourceDetail: col.DataType,
						Description:  "MySQL ENUM type not directly supported in " + string(targetDB),
						Suggestion:   "Use CHECK constraint or lookup table",
					})
				}
				if strings.Contains(dataType, "set") {
					issues = append(issues, models.CompatibilityIssue{
						Level:        models.Incompatible,
						ObjectType:   "Column",
						ObjectName:   table.Name + "." + col.Name,
						SourceDetail: col.DataType,
						Description:  "MySQL SET type not supported in " + string(targetDB),
						Suggestion:   "Use separate junction table for many-to-many relationship",
					})
				}
			}
		}
	}
	
	return issues
}