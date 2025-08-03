package models

type CompatibilityLevel string

const (
	Compatible   CompatibilityLevel = "COMPATIBLE"
	Warning      CompatibilityLevel = "WARNING"
	Incompatible CompatibilityLevel = "INCOMPATIBLE"
)

type TypeMapping struct {
	SourceType   string
	TargetType   string
	IsCompatible bool
	Warning      string
}

type CompatibilityIssue struct {
	Level        CompatibilityLevel
	ObjectType   string
	ObjectName   string
	SourceDetail string
	TargetDetail string
	Description  string
	Suggestion   string
}

type CrossDatabaseResult struct {
	ComparisonResult
	CompatibilityIssues []CompatibilityIssue
	TypeMappings        []TypeMapping
}