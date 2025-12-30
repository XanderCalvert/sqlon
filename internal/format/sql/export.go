package sql

import (
	"fmt"
	"io"
	"strings"

	"sqlon/internal/model"
)

func ExportSQLite(w io.Writer, db *model.Database) error {
	for ti, t := range db.Tables {
		if err := emitCreateTable(w, t); err != nil {
			return err
		}

		if len(t.Rows) > 0 {
			if err := emitInserts(w, t); err != nil {
				return err
			}
		}

		if ti < len(db.Tables)-1 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func emitCreateTable(w io.Writer, t *model.Table) error {
	if _, err := fmt.Fprintf(w, "CREATE TABLE %s (\n", quoteIdent(t.Name)); err != nil {
		return err
	}

	for i, c := range t.Columns {
		colLine := "    " + quoteIdent(c.Name) + " " + sqliteType(c.Type)
		if t.PK != "" && c.Name == t.PK {
			colLine += " PRIMARY KEY"
		}

		if i < len(t.Columns)-1 {
			colLine += ","
		}
		colLine += "\n"

		if _, err := io.WriteString(w, colLine); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, ");\n"); err != nil {
		return err
	}

	return nil
}

func emitInserts(w io.Writer, t *model.Table) error {
	colNames := t.ColumnNames()
	quotedCols := make([]string, 0, len(colNames))
	for _, n := range colNames {
		quotedCols = append(quotedCols, quoteIdent(n))
	}

	prefix := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES ",
		quoteIdent(t.Name),
		strings.Join(quotedCols, ", "),
	)

	for _, row := range t.Rows {
		if _, err := io.WriteString(w, prefix); err != nil {
			return err
		}

		values := make([]string, 0, len(colNames))
		for i := 0; i < len(colNames); i++ {
			if i < len(row) {
				values = append(values, sqliteLiteral(row[i]))
			} else {
				values = append(values, "NULL")
			}
		}

		if _, err := fmt.Fprintf(w, "(%s);\n", strings.Join(values, ", ")); err != nil {
			return err
		}
	}

	return nil
}

func sqliteType(t model.ColumnType) string {
	switch t {
	case model.ColumnTypeInt:
		return "INTEGER"
	case model.ColumnTypeText:
		return "TEXT"
	case model.ColumnTypeBool:
		return "INTEGER"
	case model.ColumnTypeDecimal:
		return "REAL"
	case model.ColumnTypeDatetime:
		return "TEXT"
	case model.ColumnTypeNull:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func sqliteLiteral(v model.Value) string {
	switch v.Kind {
	case model.ValueKindNull:
		return "NULL"
	case model.ValueKindInt:
		return fmt.Sprintf("%d", v.Int64)
	case model.ValueKindDecimal:
		return trimFloat(v.Float64)
	case model.ValueKindBool:
		if v.Bool {
			return "1"
		}
		return "0"
	case model.ValueKindText:
		return "'" + escapeSQLString(v.Text) + "'"
	default:
		return "NULL"
	}
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, `'`, `''`)
}

func trimFloat(f float64) string {
	s := fmt.Sprintf("%.15g", f)
	return s
}
