# Schemalyzer Visualizer

A Blazor WebAssembly application for visualizing database schemas exported from Schemalyzer.

## Features

- **Schema Upload**: Drag-and-drop or browse to upload YAML/JSON schema files
- **Visual Statistics**: View table counts, column counts, and other metrics at a glance
- **Table Explorer**: Browse tables with their columns and data types
- **Row Counts**: See row counts for each table (when exported with `--with-row-count`)
- **Export Options**: Generate PlantUML and Mermaid diagrams from your schema
- **Dark Theme**: GitHub-inspired dark theme for comfortable viewing
- **Fully Client-Side**: No server required - runs entirely in your browser

## Live Demo

Once deployed, the visualizer will be available at:
`https://[username].github.io/Schemalyzer/SchemalyzerVisualizer/`

## Using the Visualizer

1. **Export your schema** using Schemalyzer with optional statistics:
   ```bash
   schemalyzer export --type postgresql \
     --conn "postgres://user:pass@localhost/db" \
     --schema public \
     --output schema.yaml \
     --with-stats \
     --with-row-count
   ```

2. **Open the visualizer** in your browser

3. **Upload your schema file** by dragging it onto the upload area or clicking "Browse Files"

4. **Explore your schema**:
   - View statistics in the dashboard
   - Browse tables and their structures
   - Export as PlantUML or Mermaid diagrams

## Development

### Prerequisites

- .NET 9.0 SDK or later
- A modern web browser

### Running Locally

1. Navigate to the SchemalyzerVisualizer directory:
   ```bash
   cd SchemalyzerVisualizer
   ```

2. Run the development server:
   ```bash
   dotnet run
   ```

3. Open your browser to `https://localhost:5001` (or the port shown in the console)

### Building for Production

```bash
dotnet publish -c Release
```

The output will be in `bin/Release/net9.0/publish/wwwroot/`

## Deployment

The project includes a GitHub Actions workflow that automatically deploys to GitHub Pages when changes are pushed to the main branch.

### Manual Deployment

1. Build the project:
   ```bash
   dotnet publish -c Release -o dist
   ```

2. Update the base href in `dist/wwwroot/index.html` and `dist/wwwroot/404.html`:
   ```html
   <base href="/Schemalyzer/SchemalyzerVisualizer/" />
   ```

3. Deploy the contents of `dist/wwwroot/` to your web server

## Supported Schema Formats

The visualizer supports schemas exported from Schemalyzer in:
- YAML format (`.yaml`, `.yml`)
- JSON format (`.json`)

## Sample Schema

A sample blog schema is included in `wwwroot/sample-schemas/` for testing.

## Technology Stack

- **Blazor WebAssembly**: .NET running in the browser via WebAssembly
- **YamlDotNet**: For parsing YAML schema files
- **GitHub Pages**: For hosting the static site

## License

Part of the Schemalyzer project - see the main repository for license information.