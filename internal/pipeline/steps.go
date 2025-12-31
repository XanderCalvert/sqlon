package pipeline

import (
	"bytes"

	"sqlon/internal/format/json"
	"sqlon/internal/format/sql"
	"sqlon/internal/format/sqlon"
)

type JSONToSQLONStep struct{}

func (s *JSONToSQLONStep) Name() string {
	return "JSON → SQLON"
}

func (s *JSONToSQLONStep) Ext() string {
	return "sqlon"
}

func (s *JSONToSQLONStep) Run(in []byte) ([]byte, error) {
	db, err := json.Import(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := sqlon.Format(&buf, db); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type SQLONToSQLStep struct{}

func (s *SQLONToSQLStep) Name() string {
	return "SQLON → SQL (SQLite)"
}

func (s *SQLONToSQLStep) Ext() string {
	return "sqlite.sql"
}

func (s *SQLONToSQLStep) Run(in []byte) ([]byte, error) {
	db, err := sqlon.Parse(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := sql.ExportSQLite(&buf, db); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type SQLToSQLONStep struct{}

func (s *SQLToSQLONStep) Name() string {
	return "SQL → SQLON"
}

func (s *SQLToSQLONStep) Ext() string {
	return "roundtrip.sqlon"
}

func (s *SQLToSQLONStep) Run(in []byte) ([]byte, error) {
	db, err := sql.ParseSQLite(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := sqlon.Format(&buf, db); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type SQLONToJSONStep struct{}

func (s *SQLONToJSONStep) Name() string {
	return "SQLON → JSON"
}

func (s *SQLONToJSONStep) Ext() string {
	return "json.out.json"
}

func (s *SQLONToJSONStep) Run(in []byte) ([]byte, error) {
	db, err := sqlon.Parse(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.Export(&buf, db); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
