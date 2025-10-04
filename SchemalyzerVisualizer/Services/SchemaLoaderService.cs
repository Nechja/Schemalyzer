using System.Text.Json;
using YamlDotNet.Serialization;
using YamlDotNet.Serialization.NamingConventions;
using SchemalyzerVisualizer.Models;

namespace SchemalyzerVisualizer.Services;

public class SchemaLoaderService
{
    private readonly IDeserializer _yamlDeserializer;
    private readonly JsonSerializerOptions _jsonOptions;
    private Schema? _currentSchema;

    public Schema? CurrentSchema => _currentSchema;

    public SchemaLoaderService()
    {
        _yamlDeserializer = new DeserializerBuilder()
            .WithNamingConvention(CamelCaseNamingConvention.Instance)
            .IgnoreUnmatchedProperties()
            .Build();

        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNameCaseInsensitive = true
        };
    }

    public async Task<Schema> LoadFromYamlAsync(Stream stream)
    {
        using var reader = new StreamReader(stream);
        var yaml = await reader.ReadToEndAsync();
        _currentSchema = _yamlDeserializer.Deserialize<Schema>(yaml);
        return _currentSchema;
    }

    public async Task<Schema> LoadFromJsonAsync(Stream stream)
    {
        _currentSchema = await JsonSerializer.DeserializeAsync<Schema>(stream, _jsonOptions)
            ?? throw new InvalidOperationException("Failed to deserialize JSON schema");
        return _currentSchema;
    }

    public async Task<Schema> LoadFromFileAsync(Stream stream, string fileName)
    {
        var extension = Path.GetExtension(fileName).ToLowerInvariant();
        return extension switch
        {
            ".yaml" or ".yml" => await LoadFromYamlAsync(stream),
            ".json" => await LoadFromJsonAsync(stream),
            _ => throw new NotSupportedException($"File type {extension} is not supported")
        };
    }

    public void ClearSchema()
    {
        _currentSchema = null;
    }

    public Dictionary<string, List<string>> GetTableRelationships()
    {
        var relationships = new Dictionary<string, List<string>>();

        if (_currentSchema?.Tables == null)
            return relationships;

        foreach (var table in _currentSchema.Tables)
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

    public int GetTableCount() => _currentSchema?.Tables?.Count ?? 0;

    public int GetViewCount() => _currentSchema?.Views?.Count ?? 0;

    public int GetIndexCount()
    {
        var schemaIndexes = _currentSchema?.Indexes?.Count ?? 0;
        var tableIndexes = _currentSchema?.Tables?.Sum(t => t.Indexes?.Count ?? 0) ?? 0;
        return schemaIndexes + tableIndexes;
    }

    public int GetTotalColumnCount() => _currentSchema?.Tables?.Sum(t => t.Columns?.Count ?? 0) ?? 0;
}