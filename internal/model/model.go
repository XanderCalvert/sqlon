package model

import (
	"fmt"
)

type ColumnType string

const (
	ColumnTypeInt      ColumnType = "int"
	ColumnTypeText     ColumnType = "text"
	ColumnTypeBool     ColumnType = "bool"
	ColumnTypeDecimal  ColumnType = "decimal"
	ColumnTypeDatetime ColumnType = "datetime"
	ColumnTypeNull     ColumnType = "null"
)

func (t ColumnType) Valid() bool {
	switch t {
	case ColumnTypeInt, ColumnTypeText, ColumnTypeBool, ColumnTypeDecimal, ColumnTypeDatetime, ColumnTypeNull:
		return true
	default:
		return false
	}
}

type Database struct {
	Tables []*Table
}

func (db *Database) TableByName(name string) (*Table, bool) {
	for _, t := range db.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

type Table struct {
	Name        string
	Columns     []Column
	PK          string
	Rows        []Row
	ForeignKeys []ForeignKey
}

type ForeignKey struct {
	Name             string // Name of the FK column in this table
	ReferencedTable  string // Table this FK references
	ReferencedColumn string // Column this FK references (usually PK)
}

func (t *Table) ColumnNames() []string {
	out := make([]string, 0, len(t.Columns))
	for _, c := range t.Columns {
		out = append(out, c.Name)
	}
	return out
}

func (t *Table) ColumnIndex(name string) (int, bool) {
	for i, c := range t.Columns {
		if c.Name == name {
			return i, true
		}
	}
	return -1, false
}

type Column struct {
	Name string
	Type ColumnType
}

func (c Column) String() string {
	return fmt.Sprintf("%s:%s", c.Name, string(c.Type))
}

type Row []Value

type ValueKind int

const (
	ValueKindNull ValueKind = iota
	ValueKindInt
	ValueKindDecimal
	ValueKindBool
	ValueKindText
)

type Value struct {
	Kind    ValueKind
	Int64   int64
	Float64 float64
	Bool    bool
	Text    string
}

func NullValue() Value {
	return Value{Kind: ValueKindNull}
}

func IntValue(v int64) Value {
	return Value{Kind: ValueKindInt, Int64: v}
}

func DecimalValue(v float64) Value {
	return Value{Kind: ValueKindDecimal, Float64: v}
}

func BoolValue(v bool) Value {
	return Value{Kind: ValueKindBool, Bool: v}
}

func TextValue(v string) Value {
	return Value{Kind: ValueKindText, Text: v}
}
