package json

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"sqlon/internal/model"
)

func Import(r io.Reader) (*model.Database, error) {
	var data interface{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	db := &model.Database{}
	normalizer := &normalizer{
		tables:  make(map[string]*model.Table),
		counter: 1,
	}

	if err := normalizer.normalize(data, "", db); err != nil {
		return nil, err
	}

	// Convert map to slice and sort by table name for deterministic output
	db.Tables = make([]*model.Table, 0, len(normalizer.tables))
	tableNames := make([]string, 0, len(normalizer.tables))
	for name := range normalizer.tables {
		tableNames = append(tableNames, name)
	}

	// Sort table names for deterministic order
	sort.Strings(tableNames)

	for _, name := range tableNames {
		table := normalizer.tables[name]
		// Sort columns by name for deterministic order
		sort.Slice(table.Columns, func(i, j int) bool {
			return table.Columns[i].Name < table.Columns[j].Name
		})
		db.Tables = append(db.Tables, table)
	}

	return db, nil
}

type normalizer struct {
	tables  map[string]*model.Table
	counter int
}

func (n *normalizer) normalize(data interface{}, prefix string, db *model.Database) error {
	switch v := data.(type) {
	case map[string]interface{}:
		return n.normalizeObject(v, prefix, db)
	case []interface{}:
		return n.normalizeArray(v, prefix, db)
	default:
		// Primitive value - should not happen at root
		return nil
	}
}

func (n *normalizer) normalizeObject(obj map[string]interface{}, prefix string, db *model.Database) error {
	// Check if this object contains arrays - those become tables
	for key, val := range obj {
		fullKey := prefix + key
		if fullKey == "" {
			fullKey = key
		}

		switch v := val.(type) {
		case []interface{}:
			// Array becomes a table
			if err := n.createTableFromArray(fullKey, v); err != nil {
				return err
			}
		case map[string]interface{}:
			// Recursively process nested objects
			if err := n.normalizeObject(v, fullKey+"_", db); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *normalizer) normalizeArray(arr []interface{}, prefix string, db *model.Database) error {
	if len(arr) == 0 {
		return nil
	}

	tableName := prefix
	if tableName == "" {
		tableName = "array_" + strconv.Itoa(n.counter)
		n.counter++
	}

	return n.createTableFromArray(tableName, arr)
}

func (n *normalizer) createTableFromArray(tableName string, arr []interface{}) error {
	if len(arr) == 0 {
		return nil
	}

	// Check first element to determine structure
	first := arr[0]

	var columns []model.Column
	var sampleRow map[string]interface{}

	switch v := first.(type) {
	case map[string]interface{}:
		// Array of objects - infer columns from first object
		sampleRow = v
		columns = n.inferColumns(sampleRow)
	case []interface{}:
		// Array of arrays - not supported in current model
		return nil
	default:
		// Array of primitives - create single column table
		colType := inferType(first)
		columns = []model.Column{{Name: "value", Type: colType}}
	}

	// Get or create table
	table, exists := n.tables[tableName]
	if !exists {
		table = &model.Table{
			Name:    tableName,
			Columns: columns,
		}
		n.tables[tableName] = table
	}

	// Add rows
	for _, item := range arr {
		var row model.Row

		switch v := item.(type) {
		case map[string]interface{}:
			row = n.createRowFromObject(v, columns)
		default:
			// Primitive value
			row = model.Row{jsonValueToModelValue(v)}
		}

		table.Rows = append(table.Rows, row)
	}

	return nil
}

func (n *normalizer) inferColumns(obj map[string]interface{}) []model.Column {
	columns := make([]model.Column, 0, len(obj))

	// Collect all keys and sort for deterministic order
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := obj[key]
		colType := inferType(val)

		// For nested objects/arrays, use text for now (could be normalized further)
		if _, isObj := val.(map[string]interface{}); isObj {
			colType = model.ColumnTypeText
		} else if _, isArr := val.([]interface{}); isArr {
			colType = model.ColumnTypeText
		}

		columns = append(columns, model.Column{
			Name: key,
			Type: colType,
		})
	}

	return columns
}

func (n *normalizer) createRowFromObject(obj map[string]interface{}, columns []model.Column) model.Row {
	row := make(model.Row, len(columns))

	for i, col := range columns {
		val := obj[col.Name]

		// For nested objects/arrays, serialize as JSON string
		if obj, ok := val.(map[string]interface{}); ok {
			jsonBytes, _ := json.Marshal(obj)
			row[i] = model.TextValue(string(jsonBytes))
		} else if arr, ok := val.([]interface{}); ok {
			jsonBytes, _ := json.Marshal(arr)
			row[i] = model.TextValue(string(jsonBytes))
		} else {
			row[i] = jsonValueToModelValue(val)
		}
	}

	return row
}

func inferType(val interface{}) model.ColumnType {
	if val == nil {
		return model.ColumnTypeNull
	}

	switch v := val.(type) {
	case float64:
		// JSON numbers are always float64, check if it's an integer
		if v == float64(int64(v)) {
			return model.ColumnTypeInt
		}
		return model.ColumnTypeDecimal
	case bool:
		return model.ColumnTypeBool
	case string:
		return model.ColumnTypeText
	default:
		return model.ColumnTypeText
	}
}

func jsonValueToModelValue(val interface{}) model.Value {
	if val == nil {
		return model.NullValue()
	}

	switch v := val.(type) {
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return model.IntValue(int64(v))
		}
		return model.DecimalValue(v)
	case bool:
		return model.BoolValue(v)
	case string:
		return model.TextValue(v)
	default:
		return model.NullValue()
	}
}
