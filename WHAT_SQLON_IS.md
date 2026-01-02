# SQLON — SQL Object Notation

> *"What if JSON… but harder?"*

SQLON is a bold attempt to answer a question nobody asked:  
**what if we stopped pretending our data wasn't relational?**

It converts friendly, flexible, human-written documents into a rigid, table-shaped format that looks suspiciously like something a database would enjoy.

It is not faster.
It is not simpler.
It is, however, extremely explicit — and now your data has opinions about where it belongs.

---

## What is SQLON?

SQLON (SQL Object Notation) is a **strict, SQL-shaped interchange format** designed to act as a canonical internal representation for structured data.

It can be compiled *to* and *from* formats such as:

- JSON  
- CSV  
- XML  
- SQL (SQLite first)

Nested data is normalised into multiple tables with foreign keys, because eventually that's where it was going anyway.

---

## The Sales Pitch (Unreasonably Honest Edition) 💼

Tired of JSON letting you get away with things?

Sick of objects casually nesting other objects without declaring their intent, schema, or long-term commitment?

Introducing **SQLON** — the data format that asks:

> *"But what table is this really?"*

With SQLON, every innocent-looking blob of JSON:

- Is flattened  
- Normalised  
- Given foreign keys  
- And forced to explain itself  

That simple:

```json
"appearanceTools": true
```

Becomes:

A row.
In a table.
With a schema.
And opinions.

Because ambiguity is a luxury — and SQLON is here to take it away.

### Example

**Input (JSON):**

```json
{
    "$schema": "https://schemas.wp.org/trunk/theme.json",
    "version": 2,
    "customTemplates": [],
    "settings": {
        "appearanceTools": true
    }
}
```

**Output (SQLON):**

```sql
@table _root
@cols $schema:text,version:int
["https://schemas.wp.org/trunk/theme.json",2]

@table customTemplates
@cols value:text

@table settings
@cols _id:int,appearanceTools:bool
[1,true]
```

A document has been politely but firmly turned into tables.

---

## Why You Would Use SQLON ✅

You might want SQLON if:

- You want a canonical internal format that doesn't change shape on a whim
- You care about round-trip safety between formats
- You're tired of nested JSON pretending it isn't relational
- You want explicit schemas, types, and structure
- You're building tooling, not content:
  - converters
  - validators
  - compilers
  - data pipelines

SQLON is not trying to replace JSON.

It is trying to discipline it.

---

## Why You Absolutely Shouldn't ❌

You should not use SQLON if:

- You like JSON because it's quick and forgiving
- You enjoy changing data shapes at 2am without consequences
- You don't want to think about schemas
- You believe nested objects should remain free and wild
- You value happiness

SQLON will ask follow-up questions.
SQLON will remember your mistakes.
SQLON will not let you "just add a field".

---

## Philosophy 🧠

> "Your scientists were so preoccupied with whether or not they could, they didn't stop to think if they should."

SQLON is built on the belief that:

- Just because data can be nested doesn't mean it should be
- Flexibility without structure is technical debt with better PR
- Every document is one refactor away from becoming a schema
- If something survives long enough, it will demand constraints

In practice, this means:

- SQLON is relational-first
- SQLON fails early and loudly
- SQLON prefers correctness over convenience
- SQLON assumes your data will end up in a database anyway

SQLON is what happens when you stop asking "can we?" and start asking "what happens when we do?"

---

## In Summary

SQLON is not a new database.

It's a confession about what your data already is.

> Also I was bored between jobs.