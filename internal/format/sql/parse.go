package sql

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"sqlon/internal/model"
)

func ParseSQLite(r io.Reader) (*model.Database, error) {
	// Read entire file
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	db := &model.Database{}
	sql := string(content)

	// Split by statements (semicolon followed by newline or end)
	statements := splitSQLStatements(sql)

	var currentTable *model.Table

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		upperStmt := strings.ToUpper(stmt)

		if strings.HasPrefix(upperStmt, "CREATE TABLE") {
			table, err := parseCreateTable(stmt)
			if err != nil {
				return nil, err
			}
			currentTable = table
			db.Tables = append(db.Tables, table)
		} else if strings.HasPrefix(upperStmt, "INSERT INTO") {
			if currentTable == nil {
				return nil, fmt.Errorf("INSERT statement before CREATE TABLE")
			}
			row, err := parseInsert(stmt, currentTable)
			if err != nil {
				return nil, err
			}
			if row != nil {
				currentTable.Rows = append(currentTable.Rows, row)
			}
		}
	}

	return db, nil
}

func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	depth := 0
	inString := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if inString {
			if ch == '\'' && i+1 < len(sql) && sql[i+1] == '\'' {
				// Escaped quote
				current.WriteByte(ch)
				current.WriteByte(sql[i+1])
				i++
				continue
			}
			if ch == '\'' {
				inString = false
			}
			current.WriteByte(ch)
			continue
		}

		switch ch {
		case '\'':
			inString = true
			current.WriteByte(ch)
		case '(':
			depth++
			current.WriteByte(ch)
		case ')':
			depth--
			current.WriteByte(ch)
		case ';':
			if depth == 0 {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	// Add final statement if any
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

var createTableRegex = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+"?([^"\s(]+)"?\s*\(`)

func parseCreateTable(sql string) (*model.Table, error) {
	matches := createTableRegex.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not parse table name from CREATE TABLE")
	}
	tableName := strings.Trim(matches[1], `"`)

	// Extract column definitions
	start := strings.Index(sql, "(")
	end := strings.LastIndex(sql, ")")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("invalid CREATE TABLE syntax")
	}

	colsDef := sql[start+1 : end]
	columns, pk, err := parseColumns(colsDef)
	if err != nil {
		return nil, err
	}

	table := &model.Table{
		Name:    tableName,
		Columns: columns,
		PK:      pk,
	}

	return table, nil
}

func parseColumns(colsDef string) ([]model.Column, string, error) {
	columns := []model.Column{}
	pk := ""

	// Split by comma, but be careful with parentheses
	parts := splitColumnDefinitions(colsDef)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for PRIMARY KEY
		if strings.Contains(strings.ToUpper(part), "PRIMARY KEY") {
			// Extract column name before PRIMARY KEY
			pkMatch := regexp.MustCompile(`"([^"]+)"`).FindStringSubmatch(part)
			if len(pkMatch) >= 2 {
				pk = pkMatch[1]
			}
			// Remove PRIMARY KEY part
			part = regexp.MustCompile(`\s+PRIMARY\s+KEY.*`).ReplaceAllString(part, "")
		}

		// Parse column name and type
		colParts := strings.Fields(part)
		if len(colParts) < 2 {
			continue
		}

		colName := strings.Trim(colParts[0], `"`)
		colTypeStr := strings.ToUpper(colParts[1])

		colType := sqliteTypeToColumnType(colTypeStr)

		columns = append(columns, model.Column{
			Name: colName,
			Type: colType,
		})
	}

	return columns, pk, nil
}

func splitColumnDefinitions(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		switch r {
		case '(':
			depth++
			current.WriteRune(r)
		case ')':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func sqliteTypeToColumnType(sqlType string) model.ColumnType {
	switch sqlType {
	case "INTEGER":
		return model.ColumnTypeInt
	case "TEXT":
		return model.ColumnTypeText
	case "REAL":
		return model.ColumnTypeDecimal
	default:
		return model.ColumnTypeText
	}
}

func parseInsert(sql string, table *model.Table) (model.Row, error) {
	// Find "VALUES (" (case insensitive)
	upperSQL := strings.ToUpper(sql)
	valuesIdx := strings.Index(upperSQL, "VALUES (")
	if valuesIdx == -1 {
		return nil, fmt.Errorf("could not find VALUES clause in INSERT statement")
	}

	// Skip "VALUES (" to get to the actual values
	valuesStart := valuesIdx + 8 // len("VALUES (")

	// Find the matching closing parenthesis, respecting strings
	valuesStr := ""
	depth := 1
	inString := false
	for i := valuesStart; i < len(sql); i++ {
		b := sql[i]

		switch b {
		case '\'':
			if inString {
				// Check for escaped quote ''
				if i+1 < len(sql) && sql[i+1] == '\'' {
					valuesStr += string(b)
					i++ // Skip next quote
					valuesStr += string(sql[i])
					continue
				}
				inString = false
			} else {
				inString = true
			}
			valuesStr += string(b)
		case '(':
			if !inString {
				depth++
			}
			valuesStr += string(b)
		case ')':
			if !inString {
				depth--
				if depth == 0 {
					// Found the matching closing paren
					break
				}
			}
			valuesStr += string(b)
		default:
			valuesStr += string(b)
		}

		if depth == 0 {
			break
		}
	}

	if depth != 0 {
		return nil, fmt.Errorf("unmatched parentheses in VALUES clause")
	}

	values := parseValueList(valuesStr)

	row := make(model.Row, len(table.Columns))
	for i := range table.Columns {
		if i < len(values) {
			row[i] = parseSQLValue(strings.TrimSpace(values[i]), table.Columns[i].Type)
		} else {
			row[i] = model.NullValue()
		}
	}

	return row, nil
}

func parseValueList(s string) []string {
	var values []string
	var current strings.Builder
	depth := 0
	inString := false

	bytes := []byte(s)
	i := 0
	for i < len(bytes) {
		b := bytes[i]

		switch b {
		case '\'':
			if inString {
				// Check if it's escaped (double quote '')
				if i+1 < len(bytes) && bytes[i+1] == '\'' {
					current.WriteByte('\'')
					i += 2 // Skip both quotes
					continue
				}
				inString = false
			} else {
				inString = true
			}
			current.WriteByte(b)
		case '(':
			if !inString {
				depth++
			}
			current.WriteByte(b)
		case ')':
			if !inString {
				depth--
			}
			current.WriteByte(b)
		case ',':
			if depth == 0 && !inString {
				values = append(values, current.String())
				current.Reset()
			} else {
				current.WriteByte(b)
			}
		default:
			current.WriteByte(b)
		}
		i++
	}

	if current.Len() > 0 {
		values = append(values, current.String())
	}

	return values
}

func parseSQLValue(s string, colType model.ColumnType) model.Value {
	s = strings.TrimSpace(s)

	if strings.EqualFold(s, "NULL") {
		return model.NullValue()
	}

	// Remove surrounding quotes for strings
	if (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) {
		// Unescape SQL string
		unquoted := s[1 : len(s)-1]
		unquoted = strings.ReplaceAll(unquoted, "''", "'")
		unquoted = strings.ReplaceAll(unquoted, `""`, `"`)
		return model.TextValue(unquoted)
	}

	// Try parsing as number
	if strings.Contains(s, ".") {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return model.DecimalValue(f)
		}
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Check if it's a boolean (0 or 1) for INTEGER type
		if colType == model.ColumnTypeBool {
			return model.BoolValue(i != 0)
		}
		return model.IntValue(i)
	}

	// Default to text
	return model.TextValue(s)
}
