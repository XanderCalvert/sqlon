package sqlon

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sqlon/internal/model"
)

func Parse(r io.Reader) (*model.Database, error) {
	scanner := bufio.NewScanner(r)
	db := &model.Database{}

	var current *model.Table
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "--") {
			continue
		}

		if strings.HasPrefix(line, "@") {
			if strings.HasPrefix(line, "@table") {
				name := strings.TrimSpace(strings.TrimPrefix(line, "@table"))
				if name == "" {
					return nil, fmt.Errorf("line %d: @table requires a name", lineNo)
				}

				t := &model.Table{Name: name}
				db.Tables = append(db.Tables, t)
				current = t
				continue
			}

			if current == nil {
				return nil, fmt.Errorf("line %d: directive %q appears before @table", lineNo, line)
			}

			if strings.HasPrefix(line, "@cols") {
				spec := strings.TrimSpace(strings.TrimPrefix(line, "@cols"))
				cols, err := parseCols(spec)
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", lineNo, err)
				}
				current.Columns = cols
				continue
			}

			if strings.HasPrefix(line, "@pk") {
				pk := strings.TrimSpace(strings.TrimPrefix(line, "@pk"))
				if pk == "" {
					return nil, fmt.Errorf("line %d: @pk requires a column name", lineNo)
				}
				current.PK = pk
				continue
			}

			return nil, fmt.Errorf("line %d: unknown directive %q", lineNo, line)
		}

		if current == nil {
			return nil, fmt.Errorf("line %d: row appears before @table", lineNo)
		}
		if len(current.Columns) == 0 {
			return nil, fmt.Errorf("line %d: row appears before @cols for table %q", lineNo, current.Name)
		}

		row, err := parseRow(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		current.Rows = append(current.Rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Infer foreign keys from column names (e.g., "parentTable_id" -> FK to "parentTable")
	inferForeignKeys(db)

	return db, nil
}

func inferForeignKeys(db *model.Database) {
	// Build table name map
	tableMap := make(map[string]*model.Table)
	for _, table := range db.Tables {
		tableMap[table.Name] = table
	}

	// For each table, check if any columns look like foreign keys
	for _, table := range db.Tables {
		for _, col := range table.Columns {
			// Check if column name ends with "_id" and matches a table name
			if strings.HasSuffix(col.Name, "_id") {
				potentialTableName := strings.TrimSuffix(col.Name, "_id")
				if _, exists := tableMap[potentialTableName]; exists {
					// This looks like a foreign key
					table.ForeignKeys = append(table.ForeignKeys, model.ForeignKey{
						Name:             col.Name,
						ReferencedTable:  potentialTableName,
						ReferencedColumn: "id", // Default assumption
					})
				}
			}
		}
	}
}

func parseCols(spec string) ([]model.Column, error) {
	if spec == "" {
		return nil, errors.New("@cols requires a list like name:type,name:type")
	}

	parts := splitByCommaRespectingWhitespace(spec)
	cols := make([]model.Column, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		pair := strings.SplitN(p, ":", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid column definition %q (expected name:type)", p)
		}

		name := strings.TrimSpace(pair[0])
		typ := model.ColumnType(strings.TrimSpace(pair[1]))

		if name == "" {
			return nil, fmt.Errorf("invalid column definition %q (missing name)", p)
		}
		if !typ.Valid() {
			return nil, fmt.Errorf("invalid column type %q for column %q", string(typ), name)
		}

		cols = append(cols, model.Column{Name: name, Type: typ})
	}

	if len(cols) == 0 {
		return nil, errors.New("@cols produced no columns")
	}

	return cols, nil
}

func parseRow(line string) (model.Row, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return nil, fmt.Errorf("row must be a positional array like [1,\"Matt\",true], got %q", line)
	}

	inner := strings.TrimSpace(line[1 : len(line)-1])
	if inner == "" {
		return model.Row{}, nil
	}

	tokens, err := splitRowTokens(inner)
	if err != nil {
		return nil, err
	}

	row := make(model.Row, 0, len(tokens))
	for _, tok := range tokens {
		v, err := parseValue(strings.TrimSpace(tok))
		if err != nil {
			return nil, err
		}
		row = append(row, v)
	}

	return row, nil
}

func parseValue(tok string) (model.Value, error) {
	if tok == "" {
		return model.Value{}, errors.New("empty value token")
	}

	lower := strings.ToLower(tok)
	if lower == "null" {
		return model.NullValue(), nil
	}
	if lower == "true" {
		return model.BoolValue(true), nil
	}
	if lower == "false" {
		return model.BoolValue(false), nil
	}

	if strings.HasPrefix(tok, "\"") {
		s, err := parseDoubleQuotedString(tok)
		if err != nil {
			return model.Value{}, err
		}
		return model.TextValue(s), nil
	}

	if strings.Contains(tok, ".") {
		f, err := strconv.ParseFloat(tok, 64)
		if err == nil {
			return model.DecimalValue(f), nil
		}
	}

	i, err := strconv.ParseInt(tok, 10, 64)
	if err == nil {
		return model.IntValue(i), nil
	}

	return model.Value{}, fmt.Errorf("unable to parse value token %q", tok)
}

func parseDoubleQuotedString(tok string) (string, error) {
	if len(tok) < 2 || tok[0] != '"' || tok[len(tok)-1] != '"' {
		return "", fmt.Errorf("invalid quoted string %q", tok)
	}

	s := tok[1 : len(tok)-1]

	var b strings.Builder
	b.Grow(len(s))

	escaping := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaping {
			switch ch {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(ch)
			}
			escaping = false
			continue
		}

		if ch == '\\' {
			escaping = true
			continue
		}

		b.WriteByte(ch)
	}

	if escaping {
		return "", fmt.Errorf("unterminated escape in string %q", tok)
	}

	return b.String(), nil
}

func splitRowTokens(inner string) ([]string, error) {
	tokens := make([]string, 0, 8)

	start := 0
	inString := false
	escaping := false

	for i := 0; i < len(inner); i++ {
		ch := inner[i]

		if inString {
			if escaping {
				escaping = false
				continue
			}
			if ch == '\\' {
				escaping = true
				continue
			}
			if ch == '"' {
				inString = false
				continue
			}
			continue
		}

		if ch == '"' {
			inString = true
			continue
		}

		if ch == ',' {
			tokens = append(tokens, inner[start:i])
			start = i + 1
		}
	}

	if inString {
		return nil, errors.New("unterminated string in row")
	}

	tokens = append(tokens, inner[start:])
	return tokens, nil
}

func splitByCommaRespectingWhitespace(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}
