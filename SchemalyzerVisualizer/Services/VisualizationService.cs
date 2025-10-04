using SchemalyzerVisualizer.Models;

namespace SchemalyzerVisualizer.Services;

public class VisualizationService
{
    public Dictionary<string, List<string>> GetTableRelationships(Schema schema)
    {
        var relationships = new Dictionary<string, List<string>>();

        if (schema.Tables == null)
            return relationships;

        foreach (var table in schema.Tables)
        {
            var relatedTables = new List<string>();

            if (table.Constraints != null)
            {
                foreach (var constraint in table.Constraints.Where(c => c.Type == "FOREIGN_KEY"))
                {
                    if (!string.IsNullOrEmpty(constraint.ReferencedTable))
                    {
                        relatedTables.Add(constraint.ReferencedTable);
                    }
                }
            }

            if (relatedTables.Any())
            {
                relationships[table.Name] = relatedTables.Distinct().ToList();
            }
        }

        return relationships;
    }

    public string GeneratePlantUml(Schema schema)
    {
        var uml = new System.Text.StringBuilder();
        uml.AppendLine("@startuml");
        uml.AppendLine($"title {schema.Name} Database Schema");
        uml.AppendLine();

        if (schema.Tables != null)
        {
            foreach (var table in schema.Tables)
            {
                uml.AppendLine($"entity \"{table.Name}\" {{");

                if (table.Columns != null)
                {
                    foreach (var column in table.Columns.OrderBy(c => c.Position))
                    {
                        var prefix = column.IsPrimaryKey ? "*" : "";
                        var nullable = column.IsNullable ? "NULL" : "NOT NULL";
                        uml.AppendLine($"  {prefix}{column.Name} : {column.DataType} [{nullable}]");
                    }
                }

                uml.AppendLine("}");
                uml.AppendLine();
            }

            // Add relationships
            foreach (var table in schema.Tables)
            {
                if (table.Constraints == null) continue;

                foreach (var constraint in table.Constraints.Where(c => c.Type == "FOREIGN_KEY"))
                {
                    if (!string.IsNullOrEmpty(constraint.ReferencedTable))
                    {
                        uml.AppendLine($"\"{table.Name}\" }}|--|| \"{constraint.ReferencedTable}\" : {constraint.Name}");
                    }
                }
            }
        }

        uml.AppendLine("@enduml");
        return uml.ToString();
    }

    public string GenerateMermaid(Schema schema)
    {
        var mermaid = new System.Text.StringBuilder();
        mermaid.AppendLine("erDiagram");

        if (schema.Tables != null)
        {
            foreach (var table in schema.Tables)
            {
                mermaid.AppendLine($"    {table.Name} {{");

                if (table.Columns != null)
                {
                    foreach (var column in table.Columns.OrderBy(c => c.Position))
                    {
                        var type = column.DataType.Replace(" ", "_");
                        var key = column.IsPrimaryKey ? "PK" : "";
                        mermaid.AppendLine($"        {type} {column.Name} {key}");
                    }
                }

                mermaid.AppendLine("    }");
            }

            // Add relationships
            foreach (var table in schema.Tables)
            {
                if (table.Constraints == null) continue;

                foreach (var constraint in table.Constraints.Where(c => c.Type == "FOREIGN_KEY"))
                {
                    if (!string.IsNullOrEmpty(constraint.ReferencedTable))
                    {
                        mermaid.AppendLine($"    {table.Name} ||--o{{ {constraint.ReferencedTable} : \"{constraint.Name}\"");
                    }
                }
            }
        }

        return mermaid.ToString();
    }
}