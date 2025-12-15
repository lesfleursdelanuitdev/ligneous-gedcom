# Export Formats

The GEDCOM Go implementation supports multiple export formats for maximum flexibility.

## Supported Formats

### 1. GEDCOM Format (`.ged`)
The native GEDCOM 5.5.5 format.

**Exporter**: `GedcomExporter`

**Features**:
- Full GEDCOM 5.5.5 compliance
- Automatic header management (version, encoding, date, time, app info)
- CONC/CONT line splitting for long values (255 char limit)
- Preserves all record types and hierarchy

**Usage**:
```go
errorManager := gedcom.NewErrorManager()
exporter := exporter.NewGedcomExporter(errorManager, "MyApp", "1.0.0")
err := exporter.ExportToFile(tree, "output.ged")
```

### 2. JSON Format (`.json`)
Structured JSON format for API and web applications.

**Exporter**: `JsonExporter`

**Features**:
- Clean, structured JSON output
- All record types included
- Metadata and export information
- Easy to parse and process

**Usage**:
```go
errorManager := gedcom.NewErrorManager()
exporter := exporter.NewJsonExporter(errorManager)
err := exporter.ExportToFile(tree, "output.json")
```

### 3. XML Format (`.xml`)
XML format for structured data exchange.

**Exporter**: `XMLExporter`

**Features**:
- Well-formed XML with proper structure
- XML schema-compliant
- All record types included
- Metadata included

**Usage**:
```go
errorManager := gedcom.NewErrorManager()
exporter := exporter.NewXMLExporter(errorManager)
err := exporter.ExportToFile(tree, "output.xml")
```

### 4. YAML Format (`.yaml` or `.yml`)
YAML format for human-readable configuration.

**Exporter**: `YAMLExporter`

**Features**:
- Human-readable format
- Easy to edit manually
- All record types included
- Metadata included

**Usage**:
```go
errorManager := gedcom.NewErrorManager()
exporter := exporter.NewYAMLExporter(errorManager)
err := exporter.ExportToFile(tree, "output.yaml")
```

## Common Interface

All exporters implement the `Exporter` interface:

```go
type Exporter interface {
    ExportToFile(tree *gedcom.GedcomTree, filePath string) error
    ExportToString(tree *gedcom.GedcomTree) (string, error)
}
```

## Header Management

The `GedcomExporter` automatically updates the header with:
- **GEDC.VERS**: GEDCOM version (5.5.5)
- **CHAR**: Character encoding (UTF-8)
- **SOUR**: Application name
- **SOUR.VERS**: Application version
- **DATE**: Export date (DD MMM YYYY format)
- **TIME**: Export time (HH:MM:SS format)
- **FILE**: Output filename

## Round-Trip Compatibility

All formats support round-trip conversion:
- Parse GEDCOM → Export to format → Parse exported file
- Structure and data are preserved across formats

## Error Handling

All exporters use the `ErrorManager` for centralized error reporting:
- File I/O errors
- Format-specific errors
- Validation errors

## Dependencies

- **GEDCOM**: No external dependencies
- **JSON**: Standard library (`encoding/json`)
- **XML**: Standard library (`encoding/xml`)
- **YAML**: `gopkg.in/yaml.v3`

## Examples

### Export to Multiple Formats

```go
tree := parseGEDCOM("input.ged")

// GEDCOM
gedExporter := exporter.NewGedcomExporter(errorManager, "MyApp", "1.0.0")
gedExporter.ExportToFile(tree, "output.ged")

// JSON
jsonExporter := exporter.NewJsonExporter(errorManager)
jsonExporter.ExportToFile(tree, "output.json")

// XML
xmlExporter := exporter.NewXMLExporter(errorManager)
xmlExporter.ExportToFile(tree, "output.xml")

// YAML
yamlExporter := exporter.NewYAMLExporter(errorManager)
yamlExporter.ExportToFile(tree, "output.yaml")
```

### Export to String

```go
jsonExporter := exporter.NewJsonExporter(errorManager)
jsonStr, err := jsonExporter.ExportToString(tree)
if err != nil {
    log.Fatal(err)
}
fmt.Println(jsonStr)
```

## Format Comparison

| Format | Human Readable | Machine Parseable | File Size | Use Case |
|--------|---------------|-------------------|-----------|----------|
| GEDCOM | ⚠️ | ✅ | Small | Standard exchange |
| JSON   | ✅ | ✅ | Medium | APIs, web apps |
| XML    | ⚠️ | ✅ | Large | Enterprise systems |
| YAML   | ✅ | ✅ | Medium | Configuration, editing |

## Future Enhancements

Potential additional formats:
- CSV (for spreadsheet import)
- SQL (for database import)
- GraphQL (for API queries)
- Protocol Buffers (for efficient serialization)

