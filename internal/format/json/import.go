package json

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

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

	// Handle root-level primitives if root object has primitives mixed with arrays/objects
	if rootObj, ok := data.(map[string]interface{}); ok {
		hasRootPrimitives := false
		rootPrimitives := make(map[string]interface{})
		for key, val := range rootObj {
			switch val.(type) {
			case []interface{}, map[string]interface{}:
				// Not a primitive
			default:
				hasRootPrimitives = true
				rootPrimitives[key] = val
			}
		}
		if hasRootPrimitives && len(rootPrimitives) > 0 {
			// Create a special table for root primitives (will be handled specially in export)
			if err := normalizer.createTableFromPrimitiveObject(rootPrimitives, "_root", db); err != nil {
				return nil, err
			}
		}
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
	// First pass: identify what types of values this object contains
	hasArrays := false
	hasNestedObjects := false
	hasPrimitives := false

	for _, val := range obj {
		switch val.(type) {
		case []interface{}:
			hasArrays = true
		case map[string]interface{}:
			hasNestedObjects = true
		default:
			hasPrimitives = true
		}
	}

	// If object only has primitives (no arrays, no nested objects), create a single-row table
	if !hasArrays && !hasNestedObjects && hasPrimitives {
		return n.createTableFromPrimitiveObject(obj, prefix, db)
	}

	// Otherwise, process arrays and nested objects normally
	for key, val := range obj {
		fullKey := prefix + key
		if fullKey == "" {
			fullKey = key
		}

		switch v := val.(type) {
		case []interface{}:
			// Array becomes a table
			if err := n.createTableFromArray(fullKey, v, ""); err != nil {
				return err
			}
		case map[string]interface{}:
			// Recursively process nested objects
			if err := n.normalizeObject(v, fullKey+"_", db); err != nil {
				return err
			}
		}
		// Primitives are ignored here - they're only handled when object has only primitives
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

	return n.createTableFromArray(tableName, arr, "")
}

func (n *normalizer) createTableFromArray(tableName string, arr []interface{}, parentTableName string) error {
	var columns []model.Column

	if len(arr) == 0 {
		// Empty array - create table with default placeholder schema
		// We can't infer the schema, so use a generic single-column schema
		columns = []model.Column{{Name: "value", Type: model.ColumnTypeText}}
		if parentTableName != "" {
			fkColName := parentTableName + "_id"
			columns = append([]model.Column{{Name: fkColName, Type: model.ColumnTypeInt}}, columns...)
		}

		// Create empty table
		table := &model.Table{
			Name:        tableName,
			Columns:     columns,
			Rows:        []model.Row{},
			ForeignKeys: []model.ForeignKey{},
		}

		if parentTableName != "" {
			fkColName := parentTableName + "_id"
			table.ForeignKeys = append(table.ForeignKeys, model.ForeignKey{
				Name:             fkColName,
				ReferencedTable:  parentTableName,
				ReferencedColumn: "id",
			})
		}

		n.tables[tableName] = table
		return nil
	}

	// Check first element to determine structure
	first := arr[0]

	var sampleRow map[string]interface{}

	switch v := first.(type) {
	case map[string]interface{}:
		// Array of objects - separate flat fields from nested structures
		sampleRow = v
		flatColumns, _ := n.separateFlatAndNested(sampleRow)
		columns = flatColumns

		// If there's a parent table, add FK column
		if parentTableName != "" {
			fkColName := parentTableName + "_id"
			columns = append([]model.Column{{Name: fkColName, Type: model.ColumnTypeInt}}, columns...)
		}
	case []interface{}:
		// Array of arrays - not supported in current model
		return nil
	default:
		// Array of primitives - create single column table
		colType := inferType(first)
		columns = []model.Column{{Name: "value", Type: colType}}
		if parentTableName != "" {
			fkColName := parentTableName + "_id"
			columns = append([]model.Column{{Name: fkColName, Type: model.ColumnTypeInt}}, columns...)
		}
	}

	// Get or create table
	table, exists := n.tables[tableName]
	if !exists {
		table = &model.Table{
			Name:        tableName,
			Columns:     columns,
			ForeignKeys: []model.ForeignKey{},
		}

		// Add FK reference if this is a child table
		if parentTableName != "" {
			fkColName := parentTableName + "_id"
			table.ForeignKeys = append(table.ForeignKeys, model.ForeignKey{
				Name:             fkColName,
				ReferencedTable:  parentTableName,
				ReferencedColumn: "id", // Assuming parent has id column (will need auto-generated PK)
			})
		}

		n.tables[tableName] = table
	}

	// Add rows
	for rowIndex, item := range arr {
		var row model.Row

		switch v := item.(type) {
		case map[string]interface{}:
			flatCols, nestedFields := n.separateFlatAndNested(v)

			// Create row with flat fields only
			row = make(model.Row, len(columns))
			colIdx := 0

			// Add FK value if this is a child table (for now, use row index as ID)
			if parentTableName != "" {
				row[colIdx] = model.IntValue(int64(rowIndex + 1))
				colIdx++
			}

			// Add flat field values
			for _, col := range flatCols {
				if colIdx < len(row) {
					val := v[col.Name]
					row[colIdx] = jsonValueToModelValue(val)
					colIdx++
				}
			}

			// Create child table rows for nested structures
			for fieldName, fieldValue := range nestedFields {
				childTableName := tableName + "_" + fieldName
				parentRowId := rowIndex + 1

				switch nestedVal := fieldValue.(type) {
				case []interface{}:
					// Ensure child table exists, then create rows with FK reference
					if _, exists := n.tables[childTableName]; !exists {
						if err := n.createChildTableStructure(childTableName, nestedVal, tableName); err != nil {
							return err
						}
					}
					if err := n.addChildRows(childTableName, nestedVal, parentRowId); err != nil {
						return err
					}
				case map[string]interface{}:
					// Ensure child table exists, then create single row with FK reference
					if _, exists := n.tables[childTableName]; !exists {
						if err := n.createChildTableStructure(childTableName, []interface{}{nestedVal}, tableName); err != nil {
							return err
						}
					}
					if err := n.addChildObjectRow(childTableName, nestedVal, parentRowId); err != nil {
						return err
					}
				}
			}
		default:
			// Primitive value
			row = make(model.Row, len(columns))
			colIdx := 0
			if parentTableName != "" {
				row[colIdx] = model.IntValue(int64(rowIndex + 1))
				colIdx++
			}
			if colIdx < len(row) {
				row[colIdx] = jsonValueToModelValue(v)
			}
		}

		table.Rows = append(table.Rows, row)
	}

	return nil
}

func (n *normalizer) separateFlatAndNested(obj map[string]interface{}) ([]model.Column, map[string]interface{}) {
	flatColumns := make([]model.Column, 0)
	nestedFields := make(map[string]interface{})

	// Collect all keys and sort for deterministic order
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := obj[key]

		// Check if this is a nested structure
		if _, isObj := val.(map[string]interface{}); isObj {
			nestedFields[key] = val
		} else if _, isArr := val.([]interface{}); isArr {
			nestedFields[key] = val
		} else {
			// Flat field
			colType := inferType(val)
			flatColumns = append(flatColumns, model.Column{
				Name: key,
				Type: colType,
			})
		}
	}

	return flatColumns, nestedFields
}

func (n *normalizer) createChildTableStructure(childTableName string, nestedData interface{}, parentTableName string) error {
	// Create the structure of a child table (columns only, no rows yet)
	var columns []model.Column

	switch v := nestedData.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		first := v[0]
		switch firstVal := first.(type) {
		case map[string]interface{}:
			flatCols, _ := n.separateFlatAndNested(firstVal)
			columns = flatCols
		default:
			colType := inferType(firstVal)
			columns = []model.Column{{Name: "value", Type: colType}}
		}
	case map[string]interface{}:
		flatCols, _ := n.separateFlatAndNested(v)
		columns = flatCols
	default:
		return fmt.Errorf("unexpected nested data type")
	}

	// Add FK column as first column
	fkColName := parentTableName + "_id"
	columns = append([]model.Column{{Name: fkColName, Type: model.ColumnTypeInt}}, columns...)

	// Create table structure
	table := &model.Table{
		Name:    childTableName,
		Columns: columns,
		ForeignKeys: []model.ForeignKey{
			{
				Name:             fkColName,
				ReferencedTable:  parentTableName,
				ReferencedColumn: "id",
			},
		},
	}
	n.tables[childTableName] = table

	return nil
}

func (n *normalizer) addChildRows(childTableName string, arr []interface{}, parentRowId int) error {
	childTable, exists := n.tables[childTableName]
	if !exists {
		return fmt.Errorf("child table %s does not exist", childTableName)
	}

	fkColIdx := 0 // FK is first column

	for _, item := range arr {
		var childRow model.Row

		switch v := item.(type) {
		case map[string]interface{}:
			// This shouldn't happen in our current structure, but handle it
			flatCols, _ := n.separateFlatAndNested(v)
			childRow = make(model.Row, len(childTable.Columns))
			childRow[fkColIdx] = model.IntValue(int64(parentRowId))

			for i, col := range flatCols {
				if i+1 < len(childRow) {
					val := v[col.Name]
					childRow[i+1] = jsonValueToModelValue(val)
				}
			}
		default:
			// Primitive value
			childRow = make(model.Row, len(childTable.Columns))
			childRow[fkColIdx] = model.IntValue(int64(parentRowId))
			if len(childRow) > 1 {
				childRow[1] = jsonValueToModelValue(v)
			}
		}

		childTable.Rows = append(childTable.Rows, childRow)
	}

	return nil
}

func (n *normalizer) addChildObjectRow(childTableName string, obj map[string]interface{}, parentRowId int) error {
	childTable, exists := n.tables[childTableName]
	if !exists {
		return fmt.Errorf("child table %s does not exist", childTableName)
	}

	flatCols, _ := n.separateFlatAndNested(obj)
	childRow := make(model.Row, len(childTable.Columns))

	// Set FK
	childRow[0] = model.IntValue(int64(parentRowId))

	// Set flat field values
	for i, col := range flatCols {
		if i+1 < len(childRow) {
			val := obj[col.Name]
			childRow[i+1] = jsonValueToModelValue(val)
		}
	}

	childTable.Rows = append(childTable.Rows, childRow)
	return nil
}

func (n *normalizer) createTableFromPrimitiveObject(obj map[string]interface{}, prefix string, db *model.Database) error {
	tableName := prefix
	// Strip trailing underscore if present (from nested object prefixing)
	if strings.HasSuffix(tableName, "_") {
		tableName = strings.TrimSuffix(tableName, "_")
	}
	if tableName == "" {
		tableName = "root"
	}

	// Check if table already exists
	if _, exists := n.tables[tableName]; exists {
		// Table already exists, skip (shouldn't happen for primitive-only objects)
		return nil
	}

	// Create columns from object keys
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Sort for deterministic order

	columns := make([]model.Column, 0, len(keys))
	for _, key := range keys {
		val := obj[key]
		colType := inferType(val)
		columns = append(columns, model.Column{
			Name: key,
			Type: colType,
		})
	}

	// Create table
	table := &model.Table{
		Name:        tableName,
		Columns:     columns,
		Rows:        []model.Row{},
		ForeignKeys: []model.ForeignKey{},
	}

	// Create single row with all primitive values
	row := make(model.Row, len(columns))
	for i, key := range keys {
		val := obj[key]
		row[i] = jsonValueToModelValue(val)
	}
	table.Rows = append(table.Rows, row)

	n.tables[tableName] = table
	return nil
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
