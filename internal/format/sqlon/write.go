package sqlon

import (
	"fmt"
	"io"

	"sqlon/internal/model"
)

func Format(w io.Writer, db *model.Database) error {
	for ti, table := range db.Tables {
		if ti > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		// Write @table directive
		if _, err := fmt.Fprintf(w, "@table %s\n", table.Name); err != nil {
			return err
		}

		// Write @cols directive
		cols := make([]string, 0, len(table.Columns))
		for _, col := range table.Columns {
			cols = append(cols, fmt.Sprintf("%s:%s", col.Name, string(col.Type)))
		}
		if _, err := fmt.Fprintf(w, "@cols %s\n", joinColumns(cols)); err != nil {
			return err
		}

		// Write @pk directive if present
		if table.PK != "" {
			if _, err := fmt.Fprintf(w, "@pk %s\n", table.PK); err != nil {
				return err
			}
		}

		// Write rows
		for _, row := range table.Rows {
			if err := formatRow(w, row); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func joinColumns(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	result := cols[0]
	for i := 1; i < len(cols); i++ {
		result += "," + cols[i]
	}
	return result
}

func formatRow(w io.Writer, row model.Row) error {
	if _, err := io.WriteString(w, "["); err != nil {
		return err
	}

	for i, val := range row {
		if i > 0 {
			if _, err := io.WriteString(w, ","); err != nil {
				return err
			}
		}

		if err := formatValue(w, val); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, "]"); err != nil {
		return err
	}

	return nil
}

func formatValue(w io.Writer, v model.Value) error {
	switch v.Kind {
	case model.ValueKindNull:
		if _, err := io.WriteString(w, "null"); err != nil {
			return err
		}
	case model.ValueKindInt:
		if _, err := fmt.Fprintf(w, "%d", v.Int64); err != nil {
			return err
		}
	case model.ValueKindDecimal:
		if _, err := fmt.Fprintf(w, "%g", v.Float64); err != nil {
			return err
		}
	case model.ValueKindBool:
		if v.Bool {
			if _, err := io.WriteString(w, "true"); err != nil {
				return err
			}
		} else {
			if _, err := io.WriteString(w, "false"); err != nil {
				return err
			}
		}
	case model.ValueKindText:
		if _, err := fmt.Fprintf(w, "%q", v.Text); err != nil {
			return err
		}
	default:
		if _, err := io.WriteString(w, "null"); err != nil {
			return err
		}
	}
	return nil
}
