# SQLON – SQL Object Notation

SQLON is a strict, SQL-shaped interchange format designed to normalise
semi-structured data into relational tables.

## Phase 1 Scope

Supported:
- Multiple tables per file
- @table <name>
- @cols <name:type,...>
- @pk <column>
- Positional rows: [1,"Matt",true]

Not supported (yet):
- Foreign keys
- Nested structures
- JSON / CSV / XML import
- Validation beyond parsing
