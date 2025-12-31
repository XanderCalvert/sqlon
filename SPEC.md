# SQLON — SQL Object Notation

SQLON (SQL Object Notation) is a strict, SQL-shaped interchange format designed to
represent structured data in a relational form.

SQLON acts as a **canonical internal representation** that can be compiled into
real SQL and converted to and from formats such as JSON, CSV, and XML.

Unlike schema-less formats, SQLON is intentionally explicit and opinionated.

---

## Design Principles

- **Relational-first**  
  SQLON mirrors relational database concepts directly.

- **Strict, not permissive**  
  Ambiguous or malformed input must fail early.

- **Deterministic**  
  Identical input always produces identical output.

- **Readable and diff-friendly**  
  SQLON files are designed to be human-readable and version-control friendly.

- **Canonical representation**  
  SQLON is the internal normalised form; other formats are imported/exported.

---

## File Structure

A SQLON file may contain **one or more tables**, defined sequentially.

Each table consists of:
1. A table declaration
2. Column definitions
3. Optional constraints
4. Zero or more data rows

Example:

```sqlon
@table people
@cols id:int, name:text, active:bool
@pk id

[1,"Matt",true]
[2,"Calvert",false]
