using YamlDotNet.Serialization;

namespace SchemalyzerVisualizer.Models;

public class Schema
{
    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "databasetype")]
    public string DatabaseType { get; set; } = string.Empty;

    [YamlMember(Alias = "tables")]
    public List<Table> Tables { get; set; } = new();

    [YamlMember(Alias = "views")]
    public List<View> Views { get; set; } = new();

    [YamlMember(Alias = "indexes")]
    public List<Index> Indexes { get; set; } = new();

    [YamlMember(Alias = "sequences")]
    public List<Sequence> Sequences { get; set; } = new();

    [YamlMember(Alias = "procedures")]
    public List<Procedure> Procedures { get; set; } = new();

    [YamlMember(Alias = "functions")]
    public List<Function> Functions { get; set; } = new();

    [YamlMember(Alias = "triggers")]
    public List<Trigger> Triggers { get; set; } = new();
}

public class Table
{
    [YamlMember(Alias = "schema")]
    public string Schema { get; set; } = string.Empty;

    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "columns")]
    public List<Column> Columns { get; set; } = new();

    [YamlMember(Alias = "constraints")]
    public List<Constraint> Constraints { get; set; } = new();

    [YamlMember(Alias = "indexes")]
    public List<Index> Indexes { get; set; } = new();

    [YamlMember(Alias = "comment")]
    public string Comment { get; set; } = string.Empty;
}

public class Column
{
    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "datatype")]
    public string DataType { get; set; } = string.Empty;

    [YamlMember(Alias = "isnullable")]
    public bool IsNullable { get; set; }

    [YamlMember(Alias = "defaultvalue")]
    public string? DefaultValue { get; set; }

    [YamlMember(Alias = "isprimarykey")]
    public bool IsPrimaryKey { get; set; }

    [YamlMember(Alias = "isunique")]
    public bool IsUnique { get; set; }

    [YamlMember(Alias = "comment")]
    public string Comment { get; set; } = string.Empty;

    [YamlMember(Alias = "position")]
    public int Position { get; set; }
}

public class Constraint
{
    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "type")]
    public string Type { get; set; } = string.Empty;

    [YamlMember(Alias = "columns")]
    public List<string> Columns { get; set; } = new();

    [YamlMember(Alias = "referencedtable")]
    public string ReferencedTable { get; set; } = string.Empty;

    [YamlMember(Alias = "referencedcolumn")]
    public List<string> ReferencedColumn { get; set; } = new();

    [YamlMember(Alias = "checkexpression")]
    public string CheckExpression { get; set; } = string.Empty;
}

public class Index
{
    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "tablename")]
    public string TableName { get; set; } = string.Empty;

    [YamlMember(Alias = "columns")]
    public List<string> Columns { get; set; } = new();

    [YamlMember(Alias = "isunique")]
    public bool IsUnique { get; set; }

    [YamlMember(Alias = "type")]
    public string Type { get; set; } = string.Empty;
}

public class View
{
    [YamlMember(Alias = "schema")]
    public string Schema { get; set; } = string.Empty;

    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "definition")]
    public string Definition { get; set; } = string.Empty;

    [YamlMember(Alias = "columns")]
    public List<Column> Columns { get; set; } = new();
}

public class Sequence
{
    [YamlMember(Alias = "schema")]
    public string Schema { get; set; } = string.Empty;

    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "startvalue")]
    public long StartValue { get; set; }

    [YamlMember(Alias = "increment")]
    public long Increment { get; set; }

    [YamlMember(Alias = "minvalue")]
    public long MinValue { get; set; }

    [YamlMember(Alias = "maxvalue")]
    public long MaxValue { get; set; }

    [YamlMember(Alias = "iscyclic")]
    public bool IsCyclic { get; set; }

    [YamlMember(Alias = "currentvalue")]
    public long CurrentValue { get; set; }
}

public class Procedure
{
    [YamlMember(Alias = "schema")]
    public string Schema { get; set; } = string.Empty;

    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "parameters")]
    public List<Parameter> Parameters { get; set; } = new();

    [YamlMember(Alias = "body")]
    public string Body { get; set; } = string.Empty;
}

public class Function
{
    [YamlMember(Alias = "schema")]
    public string Schema { get; set; } = string.Empty;

    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "parameters")]
    public List<Parameter> Parameters { get; set; } = new();

    [YamlMember(Alias = "returntype")]
    public string ReturnType { get; set; } = string.Empty;

    [YamlMember(Alias = "body")]
    public string Body { get; set; } = string.Empty;
}

public class Parameter
{
    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "datatype")]
    public string DataType { get; set; } = string.Empty;

    [YamlMember(Alias = "direction")]
    public string Direction { get; set; } = string.Empty;
}

public class Trigger
{
    [YamlMember(Alias = "schema")]
    public string Schema { get; set; } = string.Empty;

    [YamlMember(Alias = "name")]
    public string Name { get; set; } = string.Empty;

    [YamlMember(Alias = "tablename")]
    public string TableName { get; set; } = string.Empty;

    [YamlMember(Alias = "event")]
    public string Event { get; set; } = string.Empty;

    [YamlMember(Alias = "timing")]
    public string Timing { get; set; } = string.Empty;

    [YamlMember(Alias = "body")]
    public string Body { get; set; } = string.Empty;
}