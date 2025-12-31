# SQLON — SQL Object Notation

SQLON (SQL Object Notation) is a strict, SQL-shaped interchange format designed to represent structured data in a relational form. It acts as a **canonical internal representation** that can be compiled into real SQL and converted to and from formats such as JSON, CSV, and XML.

## Features

- **Relational-first design** - Mirrors relational database concepts directly
- **Strict schema enforcement** - Explicit types and structure, fails early on malformed input
- **Format conversion** - Convert between JSON, SQL, and SQLON formats
- **Roundtrip testing** - Verify data integrity through format conversions
- **Human-readable** - Designed to be readable and version-control friendly

## Installation

### Build from source

```bash
git clone https://github.com/XanderCalvert/sqlon.git
cd sqlon
go build ./cmd/sqlon
```

This will create a `sqlon` executable (or `sqlon.exe` on Windows).

## Usage

### Convert SQLON to SQL

Convert a SQLON file to SQLite SQL:

```bash
sqlon to-sql example.sqlon
```

This outputs SQLite CREATE TABLE and INSERT statements to stdout.

### Roundtrip Pipeline

Run a complete roundtrip conversion pipeline (JSON → SQLON → SQL → SQLON → JSON):

```bash
sqlon roundtrip input.json
```

This generates files in the `out/` directory:
- `01.sqlon` - Initial SQLON conversion from JSON
- `02.sqlite.sql` - SQLite SQL output
- `03.roundtrip.sqlon` - SQLON after SQL roundtrip
- `04.json.out.json` - Final JSON output
- `pipeline.log.jsonl` - Pipeline execution log

## SQLON Format

A SQLON file contains one or more tables, defined sequentially. Each table consists of:

1. A table declaration (`@table <name>`)
2. Column definitions (`@cols <col1:type1,col2:type2,...>`)
3. Optional primary key (`@pk <column>`)
4. Zero or more data rows (arrays of values)

### Example

```sqlon
@table people
@cols id:int, name:text, active:bool
@pk id

[1,"Matt",true]
[2,"Calvert",false]
```

### Supported Types

- `int` - Integer
- `text` - Text/String
- `bool` - Boolean
- `decimal` - Decimal number
- `datetime` - DateTime
- `null` - Null type

### Comments

SQLON supports both `#` and `--` style comments:

```sqlon
# This is a comment
@table users
@cols id:int, name:text  -- Column definitions
@pk id
```

## Examples

See the `examples/` directory for example files:
- `input.json` - Sample JSON input
- `input.sqlon` - SQLON representation
- `input.sqlite.sql` - SQLite SQL output
- `input.roundtrip.sqlon` - SQLON after roundtrip
- `input.out.json` - Final JSON output

## Project Structure

```
sqlon/
├── cmd/sqlon/          # CLI application
├── internal/
│   ├── format/         # Format converters (json, sql, sqlon)
│   ├── model/          # Core data model (Database, Table, Column, Row)
│   ├── pipeline/       # Conversion pipeline
│   └── normalise/      # Normalization utilities
├── examples/           # Example files
└── .github/workflows/  # CI/CD workflows
```

## Design Principles

- **Relational-first** - SQLON mirrors relational database concepts directly
- **Strict, not permissive** - Ambiguous or malformed input must fail early
- **Deterministic** - Identical input always produces identical output
- **Readable and diff-friendly** - Designed for human readability and version control
- **Canonical representation** - SQLON is the internal normalized form

See [SPEC.md](SPEC.md) for the complete specification and [MANIFESTO.md](MANIFESTO.md) for the project philosophy.

## Status

This is an experimental project. Breaking changes are expected. The current MVP supports:

- ✅ SQLON parsing and formatting
- ✅ JSON import/export
- ✅ SQLite SQL export/import
- ✅ Roundtrip pipeline testing
- ✅ GitHub Actions CI/CD

## Related Projects

- [sqlon-vscode](https://github.com/XanderCalvert/sqlon-vscode) - VS Code syntax highlighting extension for SQLON

## License

[Add your license here]

## Contributing

Contributions are welcome! This is an exploratory project, so feel free to experiment and propose changes.
