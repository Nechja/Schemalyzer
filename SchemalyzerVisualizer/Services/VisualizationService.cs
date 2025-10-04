using Z.Blazor.Diagrams;
using Z.Blazor.Diagrams.Core.Geometry;
using Z.Blazor.Diagrams.Core.Models;
using SchemalyzerVisualizer.Models;

namespace SchemalyzerVisualizer.Services;

public class VisualizationService
{
    public Diagram CreateDiagram(Schema schema)
    {
        var diagram = new Diagram();
        diagram.Options.Zoom.Enabled = true;
        diagram.Options.Links.EnableSnapping = true;
        diagram.Options.GridSize = 20;

        if (schema.Tables == null || !schema.Tables.Any())
            return diagram;

        var tableNodes = new Dictionary<string, NodeModel>();
        var positions = CalculateNodePositions(schema.Tables.Count);

        // Create nodes for each table
        for (int i = 0; i < schema.Tables.Count; i++)
        {
            var table = schema.Tables[i];
            var position = positions[i];

            var node = diagram.Nodes.Add(new NodeModel(position)
            {
                Title = table.Name
            });

            // Add ports for foreign key connections
            node.AddPort(PortAlignment.Top);
            node.AddPort(PortAlignment.Bottom);
            node.AddPort(PortAlignment.Left);
            node.AddPort(PortAlignment.Right);

            tableNodes[table.Name] = node;
        }

        // Create links for foreign key relationships
        foreach (var table in schema.Tables)
        {
            if (table.Constraints == null) continue;

            foreach (var constraint in table.Constraints.Where(c => c.Type == "FOREIGN_KEY"))
            {
                if (string.IsNullOrEmpty(constraint.ReferencedTable)) continue;

                if (tableNodes.TryGetValue(table.Name, out var sourceNode) &&
                    tableNodes.TryGetValue(constraint.ReferencedTable, out var targetNode))
                {
                    var sourcePort = sourceNode.Ports.First();
                    var targetPort = targetNode.Ports.First();

                    diagram.Links.Add(new LinkModel(sourcePort, targetPort)
                    {
                        Labels = { new LinkLabelModel { Content = constraint.Name } }
                    });
                }
            }
        }

        return diagram;
    }

    private List<Point> CalculateNodePositions(int nodeCount)
    {
        var positions = new List<Point>();
        var columns = (int)Math.Ceiling(Math.Sqrt(nodeCount));
        var rows = (int)Math.Ceiling((double)nodeCount / columns);

        const double spacing = 200;
        const double startX = 50;
        const double startY = 50;

        for (int i = 0; i < nodeCount; i++)
        {
            var col = i % columns;
            var row = i / columns;

            positions.Add(new Point(
                startX + col * spacing,
                startY + row * spacing
            ));
        }

        return positions;
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
                        uml.AppendLine($"\"{table.Name}\" }|--|| \"{constraint.ReferencedTable}\" : {constraint.Name}");
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