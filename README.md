# Schemalyzer

[![Build and Test](https://github.com/nechja/schemalyzer/actions/workflows/build.yml/badge.svg)](https://github.com/nechja/schemalyzer/actions/workflows/build.yml)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Schemalyzer is a database schema comparison and documentation tool supporting PostgreSQL, MySQL, and Oracle. It enables cross-database schema comparison, generates ERD diagrams in formats such as PlantUML, GraphViz, Mermaid, and D2, and is optimized for performance with parallel schema reading and connection pooling. Designed for CI/CD pipelines, it provides exit codes and minimal output modes for integration, allows flexible filtering to exclude system or temporary objects, supports schema export and import in YAML or JSON for version control, and operates in read-only mode for safety in production environments.

## Installation

### Build from Source

```bash
git clone https://github.com/nechja/schemalyzer.git
cd schemalyzer
go build -o schemalyzer cmd/schemalyzer/main.go
```

### Download Binary

Pre-built binaries can be found here. [releases page](https://github.com/nechja/schemalyzer/releases) 

Once available, you can download and install:
```bash
# Linux/macOS
chmod +x schemalyzer-linux-amd64
sudo mv schemalyzer-linux-amd64 /usr/local/bin/schemalyzer

# Windows
# Add schemalyzer-windows-amd64.exe to your PATH
```

## Quick Start

### List Available Schemas

```bash
schemalyzer list --type postgresql --conn "postgres://user:pass@localhost/dbname?sslmode=disable"
```

### Compare Two Schemas

```bash
schemalyzer compare \
  --source-type postgresql \
  --source-conn "postgres://user:pass@localhost/prod_db?sslmode=disable" \
  --source-schema public \
  --target-type postgresql \
  --target-conn "postgres://user:pass@localhost/dev_db?sslmode=disable" \
  --target-schema public \
  --format summary
```

### Export Schema to File

```bash
schemalyzer export \
  --type mysql \
  --conn "user:pass@tcp(localhost:3306)/database" \
  --schema myschema \
  --output schema.yaml
```

### Generate Visual Documentation

```bash
# Generate GraphViz diagram
schemalyzer document \
  --type oracle \
  --conn "oracle://user:pass@localhost:1521/ORCL" \
  --schema MYSCHEMA \
  --format graphviz \
  --output schema.dot

# Convert to PNG
dot -Tpng schema.dot -o schema.png
```

### Validate Schema in CI/CD Pipeline

```bash
schemalyzer validate \
  --type postgresql \
  --conn "$DATABASE_URL" \
  --schema public \
  --golden expected-schema.yaml \
  --pipeline
```

## Connection Strings

### PostgreSQL
```
postgres://username:password@host:port/database?sslmode=disable
postgresql://username:password@host:port/database?sslmode=disable (if you need sslmode disabled)
```

### MySQL
```
username:password@tcp(host:port)/database
```

### Oracle
```
oracle://username:password@host:port/service_name
```

## Commands

### `compare` - Compare two database schemas

```bash
schemalyzer compare [flags]

Flags:
  --source-type string     Source database type (postgresql, mysql, oracle)
  --source-conn string     Source database connection string
  --source-schema string   Source schema name
  --target-type string     Target database type (postgresql, mysql, oracle)
  --target-conn string     Target database connection string
  --target-schema string   Target schema name
  --format string          Output format (json, yaml, text, summary) (default "text")
  --output string          Output file path (default: stdout)
  --ignore strings         Ignore patterns (e.g., 'table:temp_*', 'constraint:SYS_*')
  --tables-only            Compare only tables and their structure (no procedures, functions, triggers)
```

### `validate` - Validate schema against a golden file

Perfect for CI/CD pipelines. Returns exit code 0 if schemas match, 2 if they differ.

```bash
schemalyzer validate [flags]

Flags:
  --type string      Database type (postgresql, mysql, oracle)
  --conn string      Database connection string
  --schema string    Schema name to validate
  --golden string    Golden schema file (JSON or YAML)
  --pipeline         Pipeline mode: minimal output, only exit codes
  --ignore strings   Ignore patterns
```

### `export` - Export schema to file

```bash
schemalyzer export [flags]

Flags:
  --type string      Database type (postgresql, mysql, oracle)
  --conn string      Database connection string
  --schema string    Schema name to export
  --output string    Output file path (required)
  --tables-only      Export only tables and their structure (no procedures, functions, triggers)
```

### `document` - Generate visual documentation

```bash
schemalyzer document [flags]

Flags:
  --type string      Database type (postgresql, mysql, oracle)
  --conn string      Database connection string
  --schema string    Schema name to document
  --format string    Documentation format (markdown, plantuml, mermaid, graphviz, d2)
  --output string    Output file path (required)
  --tables-only      Document only tables and their structure (no procedures, functions, triggers)
```

### `list` - List available schemas

```bash
schemalyzer list [flags]

Flags:
  --type string    Database type (postgresql, mysql, oracle)
  --conn string    Database connection string
```

## Ignore Patterns

Use ignore patterns to exclude specific database objects from comparison:

```bash
# Ignore all tables starting with 'temp_'
--ignore "table:temp_*"

# Ignore system-generated constraints
--ignore "constraint:SYS_*"

# Ignore all audit-related objects
--ignore "*_audit"

# Multiple patterns
--ignore "table:temp_*" --ignore "constraint:SYS_*" --ignore "index:idx_temp_*"
```

Pattern format: `[object_type:]pattern`

Object types: `table`, `column`, `constraint`, `index`, `view`, `sequence`, `procedure`, `function`, `trigger`, or `*` for all

## Tables Only Mode

The `--tables-only` flag allows you to focus exclusively on the core data schema, excluding stored procedures, functions, triggers, and sequences. This is useful when:

- You only care about data structure changes
- Comparing schemas across different database vendors with incompatible procedural code
- Generating cleaner ERD diagrams
- Creating minimal schema documentation

```bash
# Export only tables and views
schemalyzer export \
  --type postgresql \
  --conn "postgres://user:pass@localhost/db" \
  --schema public \
  --output schema.yaml \
  --tables-only

# Compare only table structures
schemalyzer compare \
  --source-type postgresql \
  --source-conn "postgres://user:pass@localhost/prod" \
  --source-schema public \
  --target-type mysql \
  --target-conn "user:pass@tcp(localhost:3306)/dev" \
  --target-schema dev \
  --tables-only

# Generate ERD with only tables
schemalyzer document \
  --type oracle \
  --conn "oracle://user:pass@localhost:1521/XE" \
  --schema MYSCHEMA \
  --format graphviz \
  --output tables.dot \
  --tables-only
```

## Output Formats

### Comparison Formats

- **json** - Structured JSON output for parsing
- **yaml** - YAML format for human readability
- **text** - Detailed text output with all differences
- **summary** - Concise summary of differences

### Documentation Formats

- **markdown** - Comprehensive documentation with tables
- **plantuml** - PlantUML ERD diagrams
- **mermaid** - Mermaid diagrams (GitHub/GitLab native)
- **graphviz** - DOT format for GraphViz (requires `graphviz` package)
- **d2** - Modern D2 diagramming language

#### GraphViz Example

```bash
# Generate GraphViz ERD
schemalyzer document \
  --type postgresql \
  --conn "postgres://user:pass@localhost/db" \
  --schema public \
  --format graphviz \
  --output schema.dot

# Convert to PNG (requires graphviz installed)
dot -Tpng schema.dot -o schema.png

# Convert to SVG
dot -Tsvg schema.dot -o schema.svg
```

To install GraphViz:
- Ubuntu/Debian: `sudo apt-get install graphviz`
- macOS: `brew install graphviz`
- Windows: Download from [graphviz.org](https://graphviz.org/download/)

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Schema Validation

on: [push, pull_request]

jobs:
  validate-schema:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download Schemalyzer
        run: |
          wget https://github.com/nechja/schemalyzer/releases/latest/download/schemalyzer-linux-amd64
          chmod +x schemalyzer-linux-amd64
          
      - name: Validate Database Schema
        run: |
          ./schemalyzer-linux-amd64 validate \
            --type postgresql \
            --conn "${{ secrets.DATABASE_URL }}" \
            --schema public \
            --golden schema/expected.yaml \
            --pipeline \
            --ignore "constraint:SYS_*"
```

### Azure DevOps Pipeline Example

```yaml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

steps:
  - checkout: self

  - script: |
      wget https://github.com/nechja/schemalyzer/releases/latest/download/schemalyzer-linux-amd64
      chmod +x schemalyzer-linux-amd64
    displayName: 'Download Schemalyzer'

  - script: |
      ./schemalyzer-linux-amd64 validate \
        --type postgresql \
        --conn "$(DATABASE_URL)" \
        --schema public \
        --golden schema/expected.yaml \
        --pipeline \
        --ignore "constraint:SYS_*"
    displayName: 'Validate Database Schema'
```

## Performance Features

- **Parallel Schema Reading** - Fetches tables, views, procedures, etc. concurrently
- **Connection Pooling** - Optimized for large databases with configurable pool settings
- **Streaming Output** - Efficient memory usage for large schemas

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE) file for details.