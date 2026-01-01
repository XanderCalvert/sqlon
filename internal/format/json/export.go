package json

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"sqlon/internal/model"
)

func Export(w io.Writer, db *model.Database) error {
	// Build table map and identify parent-child relationships
	tableMap := make(map[string]*model.Table)
	childTables := make(map[string][]*model.Table) // parent -> children

	for _, table := range db.Tables {
		tableMap[table.Name] = table

		// Find child tables (tables that reference this one)
		for _, otherTable := range db.Tables {
			for _, fk := range otherTable.ForeignKeys {
				if fk.ReferencedTable == table.Name {
					childTables[table.Name] = append(childTables[table.Name], otherTable)
				}
			}
		}
	}

	result := make(map[string]interface{})
	var rootPrimitives map[string]interface{}

	// Process tables, skipping child tables (they'll be merged into parents)
	for _, table := range db.Tables {
		// Skip if this is a child table (has FK to another table)
		if len(table.ForeignKeys) > 0 {
			continue
		}

		// Handle special "_root" table - merge its fields into top level
		if table.Name == "_root" {
			rows := exportTableRows(table, childTables[table.Name], tableMap)
			if len(rows) == 1 {
				rootPrimitives = rows[0]
			}
			continue
		}

		rows := exportTableRows(table, childTables[table.Name], tableMap)

		// If table has exactly one row and no child tables, export as object instead of array
		if len(rows) == 1 && len(childTables[table.Name]) == 0 {
			result[table.Name] = rows[0]
		} else {
			result[table.Name] = rows
		}
	}

	// Merge root primitives into top level
	if rootPrimitives != nil {
		for key, val := range rootPrimitives {
			result[key] = val
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func exportTableRows(table *model.Table, childTables []*model.Table, tableMap map[string]*model.Table) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(table.Rows))

	colNames := table.ColumnNames()

	for rowIndex, row := range table.Rows {
		rowObj := make(map[string]interface{})

		// Get row ID (using row index + 1, or from id column if exists)
		rowId := rowIndex + 1
		if len(table.Columns) > 0 {
			// Check if first column is an id column
			for i, colName := range colNames {
				if colName == "id" && i < len(row) && row[i].Kind == model.ValueKindInt {
					rowId = int(row[i].Int64)
					break
				}
			}
		}

		// Add flat field values (parent tables don't have FK columns)
		for i, colName := range colNames {
			var val interface{}
			if i < len(row) {
				val = modelValueToJSONValue(row[i])
			} else {
				val = nil
			}
			rowObj[colName] = val
		}

		// Merge child table data back into parent row
		for _, childTable := range childTables {
			if len(childTable.ForeignKeys) == 0 {
				continue
			}

			fkColIdx := 0 // FK is first column in child table

			// Extract field name from child table name (e.g., "settings_color_duotone_colors" -> "colors")
			fieldName := strings.TrimPrefix(childTable.Name, table.Name+"_")

			// Collect child rows that reference this parent row
			childRows := make([]interface{}, 0)
			for _, childRow := range childTable.Rows {
				if fkColIdx < len(childRow) && childRow[fkColIdx].Kind == model.ValueKindInt {
					if int(childRow[fkColIdx].Int64) == rowId {
						// This child row belongs to this parent row
						childRowObj := buildChildRowObject(childTable, childRow, fkColIdx)

						// If child table has only one non-FK column named "value", extract just the value
						nonFKCols := getNonFKColumns(childTable)
						if len(nonFKCols) == 1 && nonFKCols[0].Name == "value" {
							// This is an array of primitives
							if val, ok := childRowObj["value"]; ok {
								childRows = append(childRows, val)
							}
						} else {
							// This is an array of objects
							childRows = append(childRows, childRowObj)
						}
					}
				}
			}

			// Add nested array to parent row
			rowObj[fieldName] = childRows
		}

		rows = append(rows, rowObj)
	}

	return rows
}

func buildChildRowObject(childTable *model.Table, childRow model.Row, fkColIdx int) map[string]interface{} {
	rowObj := make(map[string]interface{})
	colNames := childTable.ColumnNames()

	for i, colName := range colNames {
		// Skip FK column
		if i == fkColIdx {
			continue
		}

		var val interface{}
		if i < len(childRow) {
			val = modelValueToJSONValue(childRow[i])
		} else {
			val = nil
		}
		rowObj[colName] = val
	}

	return rowObj
}

func getNonFKColumns(table *model.Table) []model.Column {
	nonFKCols := make([]model.Column, 0)
	fkColNames := make(map[string]bool)

	for _, fk := range table.ForeignKeys {
		fkColNames[fk.Name] = true
	}

	for _, col := range table.Columns {
		if !fkColNames[col.Name] {
			nonFKCols = append(nonFKCols, col)
		}
	}

	return nonFKCols
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
