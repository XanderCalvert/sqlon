package sqlon

import (
	"strings"
	"testing"
)

func TestParseBasicTable(t *testing.T) {
	input := `
@table people
@cols id:int, name:text
@pk id

[1,"Matt"]
`

	db, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(db.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(db.Tables))
	}
}
