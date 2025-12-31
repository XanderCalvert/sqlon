# SQLON Roadmap

This document tracks the implementation status of planned features for SQLON.

## Phase 2: Basic Format Conversions ✅ COMPLETE

- [x] **Import JSON (flat array of objects) → single table SQLON**
  - ✅ Implemented in `internal/format/json/import.go`
  - Supports arrays of objects, primitive arrays, and nested structures
  
- [x] **Export SQLON → JSON**
  - ✅ Implemented in `internal/format/json/export.go`
  - Produces JSON with table names as keys and arrays of row objects as values

- [x] **Convert: json ↔ sqlon**
  - ✅ Both directions working
  - Used in roundtrip pipeline: `JSON → SQLON → SQL → SQLON → JSON`

## Phase 3: Advanced Normalization & CSV Support 🚧 IN PROGRESS

- [x] **Normaliser for nested JSON → multi-table SQLON**
  - ✅ Partially implemented - nested JSON creates multiple tables
  - ✅ Tables are extracted from nested structures (e.g., `settings_color_palette`, `settings_spacing_spacingSizes`)
  - ⚠️ Foreign keys not yet explicitly tracked/managed
  - 🔄 Future work: Explicit FK relationships in `internal/normalise/normalise.go`

- [ ] **Export each table to CSV (folder output)**
  - 🔲 Not yet implemented - `internal/format/csv/export.go` is placeholder
  - Plan: Output each table as a separate CSV file in a directory

- [ ] **Convert: sqlon ↔ csv (multi-table aware)**
  - 🔲 Not yet implemented - `internal/format/csv/import.go` and `export.go` are placeholders
  - Plan: Support reading/writing multiple CSV files for a SQLON database

## Phase 4: XML & Advanced Features 🔲 PLANNED

- [ ] **XML import/export using fixed convention**
  - 🔲 Not yet implemented - `internal/format/xml/import.go` and `export.go` are placeholders
  - Plan: Convert between XML and SQLON formats

- [ ] **Optional: minifier `.sqlon.min`**
  - 🔲 Not yet implemented
  - Plan: Create a compact/minified version of SQLON format

## Additional Features Implemented

Beyond the original phases, the following have been completed:

- [x] **SQLite SQL export/import**
  - ✅ Export SQLON → SQLite SQL (`internal/format/sql/export.go`)
  - ✅ Import SQLite SQL → SQLON (`internal/format/sql/parse.go`)
  
- [x] **Roundtrip pipeline**
  - ✅ Complete pipeline: JSON → SQLON → SQL → SQLON → JSON
  - ✅ Pipeline logging and artifact management
  
- [x] **CLI tool**
  - ✅ `sqlon to-sql <file.sqlon>` - Convert SQLON to SQL
  - ✅ `sqlon roundtrip <file.json>` - Run complete roundtrip pipeline
  
- [x] **CI/CD**
  - ✅ GitHub Actions workflow for roundtrip testing
  - ✅ Automated regression testing

## Current Status Summary

- **Phase 2**: ✅ Complete
- **Phase 3**: 🚧 50% (normalization done, CSV pending)
- **Phase 4**: 🔲 Not started

## Next Steps

1. **Complete Phase 3**: Implement CSV import/export
2. **Enhance normalization**: Add explicit foreign key tracking
3. **Begin Phase 4**: Start XML format support
4. **Consider**: Minifier implementation based on use cases

