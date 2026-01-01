package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"sqlon/internal/model"
)

func Import(r io.Reader) (*model.Database, error) {
	// Read the entire JSON to parse it twice: once to get key order, once to decode
	jsonBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON: %w", err)
	}

	// First pass: extract root-level key order by parsing tokens
	rootPrimitiveKeys := extractRootKeyOrder(jsonBytes)

	// Second pass: decode normally
	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	db := &model.Database{}
	normalizer := &normalizer{
		tables:  make(map[string]*model.Table),
		counter: 1,
	}

	// Extract root-level primitives using the preserved key order
	var rootPrimitives map[string]interface{}
	if rootObj, ok := data.(map[string]interface{}); ok {
		hasRootPrimitives := false
		rootPrimitives = make(map[string]interface{})
		// Use the preserved key order, filtering to only primitives
		for _, key := range rootPrimitiveKeys {
			val, exists := rootObj[key]
			if !exists {
				continue
			}
			switch val.(type) {
			case []interface{}, map[string]interface{}:
				// Not a primitive
			default:
				hasRootPrimitives = true
				rootPrimitives[key] = val
			}
		}
		if hasRootPrimitives && len(rootPrimitives) > 0 {
			// Filter rootPrimitiveKeys to only include keys that are actually primitives
			filteredKeys := make([]string, 0, len(rootPrimitiveKeys))
			for _, key := range rootPrimitiveKeys {
				if _, ok := rootPrimitives[key]; ok {
					filteredKeys = append(filteredKeys, key)
				}
			}
			// Create a special table for root primitives with preserved order
			if err := normalizer.createTableFromPrimitiveObjectWithOrder(rootPrimitives, filteredKeys, "_root", db); err != nil {
				return nil, err
			}
		}
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
		// Sort columns by name for deterministic order, except:
		// - _root table (preserve original order)
		// - Tables with foreign keys (keep FK columns first)
		if name != "_root" {
			// Check if table has foreign keys
			if len(table.ForeignKeys) > 0 {
				// Separate FK columns from non-FK columns
				fkColNames := make(map[string]bool)
				for _, fk := range table.ForeignKeys {
					fkColNames[fk.Name] = true
				}
				fkCols := make([]model.Column, 0)
				nonFKCols := make([]model.Column, 0)
				for _, col := range table.Columns {
					if fkColNames[col.Name] {
						fkCols = append(fkCols, col)
					} else {
						nonFKCols = append(nonFKCols, col)
					}
				}
				// Sort non-FK columns, keep FK columns in their current order (should be first)
				sort.Slice(nonFKCols, func(i, j int) bool {
					return nonFKCols[i].Name < nonFKCols[j].Name
				})
				// Reconstruct columns: FK first, then sorted non-FK
				table.Columns = append(fkCols, nonFKCols...)
			} else {
				// No foreign keys, sort all columns
				sort.Slice(table.Columns, func(i, j int) bool {
					return table.Columns[i].Name < table.Columns[j].Name
				})
			}
		}
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

	// If object has primitives AND nested structures, create a table for primitives first
	// then process nested structures as child tables
	if hasPrimitives && (hasArrays || hasNestedObjects) {
		// Extract primitives
		primitives := make(map[string]interface{})
		for key, val := range obj {
			switch val.(type) {
			case []interface{}, map[string]interface{}:
				// Not a primitive
			default:
				primitives[key] = val
			}
		}
		if len(primitives) > 0 {
			// Create table for primitives - nested structures will become child tables
			if err := n.createTableFromPrimitiveObject(primitives, prefix, db); err != nil {
				return err
			}
		}
	}

	// Process arrays and nested objects (they become child tables if parent has primitives)
	parentTableName := prefix
	if strings.HasSuffix(parentTableName, "_") {
		parentTableName = strings.TrimSuffix(parentTableName, "_")
	}
	if parentTableName == "" {
		parentTableName = prefix
	}

	for key, val := range obj {
		fullKey := prefix + key
		if fullKey == "" {
			fullKey = key
		}

		switch v := val.(type) {
		case []interface{}:
			// Array becomes a table (child table if parent has primitives)
			parentName := ""
			if hasPrimitives && (hasArrays || hasNestedObjects) {
				parentName = parentTableName
			}
			if err := n.createTableFromArray(fullKey, v, parentName); err != nil {
				return err
			}
		case map[string]interface{}:
			// For nested objects, if parent has primitives, create as child table
			// Otherwise, recursively process normally
			if hasPrimitives && (hasArrays || hasNestedObjects) {
				// Create as child table - check if it only has primitives
				hasOnlyPrimitives := true
				for _, nestedVal := range v {
					switch nestedVal.(type) {
					case []interface{}, map[string]interface{}:
						hasOnlyPrimitives = false
						break
					}
				}
				if hasOnlyPrimitives {
					// Create single-row child table with FK as first column
					keys := make([]string, 0, len(v))
					for key := range v {
						keys = append(keys, key)
					}
					sort.Strings(keys)

					// Create table with FK column first
					childTableName := strings.TrimSuffix(fullKey+"_", "_")
					fkColName := parentTableName + "_id"

					// Create columns with FK first
					columns := make([]model.Column, 0, len(keys)+1)
					columns = append(columns, model.Column{Name: fkColName, Type: model.ColumnTypeInt})
					for _, key := range keys {
						if val, ok := v[key]; ok {
							colType := inferType(val)
							columns = append(columns, model.Column{
								Name: key,
								Type: colType,
							})
						}
					}

					// Create table
					table := &model.Table{
						Name:    childTableName,
						Columns: columns,
						Rows:    []model.Row{},
						ForeignKeys: []model.ForeignKey{
							{
								Name:             fkColName,
								ReferencedTable:  parentTableName,
								ReferencedColumn: "id",
							},
						},
					}

					// Create single row with FK value first, then primitive values
					row := make(model.Row, len(columns))
					row[0] = model.IntValue(1) // FK value
					for i, key := range keys {
						if val, ok := v[key]; ok {
							row[i+1] = jsonValueToModelValue(val)
						}
					}
					table.Rows = append(table.Rows, row)

					n.tables[childTableName] = table
				} else {
					// Has nested structures - recursively process
					if err := n.normalizeObject(v, fullKey+"_", db); err != nil {
						return err
					}
				}
			} else {
				// No parent primitives, process normally
				if err := n.normalizeObject(v, fullKey+"_", db); err != nil {
					return err
				}
			}
		}
		// Primitives are already handled above
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

	// Check if table already exists (e.g., _root was already created for root-level primitives)
	if _, exists := n.tables[tableName]; exists {
		// Table already exists, skip
		return nil
	}
	// Also check if _root exists and we're trying to create root (avoid duplicate)
	if tableName == "root" {
		if _, exists := n.tables["_root"]; exists {
			// _root already exists, don't create root
			return nil
		}
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

func (n *normalizer) createTableFromPrimitiveObjectWithOrder(obj map[string]interface{}, keys []string, prefix string, db *model.Database) error {
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

	// Create columns from keys in the specified order
	columns := make([]model.Column, 0, len(keys))
	for _, key := range keys {
		if val, ok := obj[key]; ok {
			colType := inferType(val)
			columns = append(columns, model.Column{
				Name: key,
				Type: colType,
			})
		}
	}

	// Create table
	table := &model.Table{
		Name:        tableName,
		Columns:     columns,
		Rows:        []model.Row{},
		ForeignKeys: []model.ForeignKey{},
	}

	// Create single row with all primitive values in the specified order
	row := make(model.Row, len(columns))
	for i, key := range keys {
		if val, ok := obj[key]; ok {
			row[i] = jsonValueToModelValue(val)
		}
	}
	table.Rows = append(table.Rows, row)

	n.tables[tableName] = table
	return nil
}

// extractRootKeyOrder parses JSON tokens to extract root-level keys in their original order
func extractRootKeyOrder(jsonBytes []byte) []string {
	decoder := json.NewDecoder(bytes.NewReader(jsonBytes))
	keys := make([]string, 0)

	// Skip to the first token (should be '{')
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return keys
	}

	// Read key-value pairs
	for decoder.More() {
		// Read key
		keyToken, err := decoder.Token()
		if err != nil {
			break
		}
		if key, ok := keyToken.(string); ok {
			keys = append(keys, key)
		}

		// Skip the value by reading its token
		valueToken, err := decoder.Token()
		if err != nil {
			break
		}

		// If value is a nested object or array, skip its contents
		if delim, ok := valueToken.(json.Delim); ok {
			if delim == '{' {
				skipObject(decoder)
			} else if delim == '[' {
				skipArray(decoder)
			}
		}
		// Primitive values are already consumed by Token()
	}

	return keys
}

func skipObject(decoder *json.Decoder) {
	for decoder.More() {
		decoder.Token() // skip key
		valueToken, err := decoder.Token()
		if err != nil {
			break
		}
		if delim, ok := valueToken.(json.Delim); ok {
			if delim == '{' {
				skipObject(decoder)
			} else if delim == '[' {
				skipArray(decoder)
			}
		}
	}
	decoder.Token() // consume closing '}'
}

func skipArray(decoder *json.Decoder) {
	for decoder.More() {
		valueToken, err := decoder.Token()
		if err != nil {
			break
		}
		if delim, ok := valueToken.(json.Delim); ok {
			if delim == '{' {
				skipObject(decoder)
			} else if delim == '[' {
				skipArray(decoder)
			}
		}
	}
	decoder.Token() // consume closing ']'
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
