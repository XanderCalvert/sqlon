package json

import (
	"encoding/json"
	"fmt"
	"io"

	"sqlon/internal/model"
)

func Export(w io.Writer, db *model.Database) error {
	result := make(map[string]interface{})

	for _, table := range db.Tables {
		rows := make([]map[string]interface{}, 0, len(table.Rows))

		colNames := table.ColumnNames()

		for _, row := range table.Rows {
			rowObj := make(map[string]interface{})

			for i, colName := range colNames {
				var val interface{}
				if i < len(row) {
					val = modelValueToJSONValue(row[i])
				} else {
					val = nil
				}
				rowObj[colName] = val
			}

			rows = append(rows, rowObj)
		}

		result[table.Name] = rows
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func modelValueToJSONValue(v model.Value) interface{} {
	switch v.Kind {
	case model.ValueKindNull:
		return nil
	case model.ValueKindInt:
		return v.Int64
	case model.ValueKindDecimal:
		return v.Float64
	case model.ValueKindBool:
		return v.Bool
	case model.ValueKindText:
		return v.Text
	default:
		return nil
	}
}
