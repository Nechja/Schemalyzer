package models

import "time"

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	Oracle     DatabaseType = "oracle"
)

type Schema struct {
	Name         string
	DatabaseType DatabaseType
	Tables       []Table
	Views        []View
	Indexes      []Index
	Sequences    []Sequence
	Procedures   []Procedure
	Functions    []Function
	Triggers     []Trigger
	Stats        *SchemaStats `yaml:"stats,omitempty" json:"stats,omitempty"`
}

type Table struct {
	Schema      string
	Name        string
	Columns     []Column
	Constraints []Constraint
	Indexes     []Index
	Comment     string
	RowCount    *int64 `yaml:"row_count,omitempty" json:"row_count,omitempty"`
}

type Column struct {
	Name            string
	DataType        string
	IsNullable      bool
	DefaultValue    *string
	IsPrimaryKey    bool
	IsUnique        bool
	IsAutoIncrement bool `yaml:"auto_increment,omitempty" json:"auto_increment,omitempty"`
	Comment         string
	Position        int
	Samples         []string `yaml:"samples,omitempty" json:"samples,omitempty"`
}

type Constraint struct {
	Name             string
	Type             ConstraintType
	Columns          []string
	ReferencedTable  string
	ReferencedColumn []string
	OnUpdate         string `yaml:"on_update,omitempty" json:"on_update,omitempty"`
	OnDelete         string `yaml:"on_delete,omitempty" json:"on_delete,omitempty"`
	CheckExpression  string
}

type ConstraintType string

const (
	PrimaryKey ConstraintType = "PRIMARY_KEY"
	ForeignKey ConstraintType = "FOREIGN_KEY"
	Unique     ConstraintType = "UNIQUE"
	Check      ConstraintType = "CHECK"
	NotNull    ConstraintType = "NOT_NULL"
)

type Index struct {
	Name      string
	TableName string
	Columns   []string
	IsUnique  bool
	Type      string
}

type View struct {
	Schema     string
	Name       string
	Definition string
	Columns    []Column
}

type Sequence struct {
	Schema       string
	Name         string
	StartValue   int64
	Increment    int64
	MinValue     int64
	MaxValue     int64
	IsCyclic     bool
	CurrentValue int64
}

type Procedure struct {
	Schema     string
	Name       string
	Parameters []Parameter
	Body       string
}

type Function struct {
	Schema     string
	Name       string
	Parameters []Parameter
	ReturnType string
	Body       string
}

type Parameter struct {
	Name      string
	DataType  string
	Direction ParameterDirection
}

type ParameterDirection string

const (
	In    ParameterDirection = "IN"
	Out   ParameterDirection = "OUT"
	InOut ParameterDirection = "INOUT"
)

type Trigger struct {
	Schema    string
	Name      string
	TableName string
	Event     TriggerEvent
	Timing    TriggerTiming
	Body      string
}

type TriggerEvent string

const (
	Insert TriggerEvent = "INSERT"
	Update TriggerEvent = "UPDATE"
	Delete TriggerEvent = "DELETE"
)

type TriggerTiming string

const (
	Before TriggerTiming = "BEFORE"
	After  TriggerTiming = "AFTER"
)

type Difference struct {
	Type        DifferenceType
	ObjectType  string
	ObjectName  string
	Source      interface{}
	Target      interface{}
	Description string
}

type DifferenceType string

const (
	Added    DifferenceType = "ADDED"
	Removed  DifferenceType = "REMOVED"
	Modified DifferenceType = "MODIFIED"
)

type ComparisonResult struct {
	SourceSchema   *Schema
	TargetSchema   *Schema
	Differences    []Difference
	ComparisonTime time.Time
	SourceDatabase string
	TargetDatabase string
}

type SchemaStats struct {
	TableCount   int       `yaml:"table_count" json:"table_count"`
	ViewCount    int       `yaml:"view_count" json:"view_count"`
	TotalColumns int       `yaml:"total_columns" json:"total_columns"`
	IndexCount   int       `yaml:"index_count" json:"index_count"`
	GeneratedAt  time.Time `yaml:"generated_at" json:"generated_at"`
}
