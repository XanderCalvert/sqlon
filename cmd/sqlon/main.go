package main

import (
	"fmt"
	"os"

	"sqlon/internal/format/sql"
	"sqlon/internal/format/sqlon"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}

	switch args[0] {
	case "to-sql":
		if len(args) != 2 {
			usage()
			os.Exit(2)
		}
		if err := runToSQL(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		usage()
	default:
		usage()
		os.Exit(2)
	}
}

func runToSQL(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	db, err := sqlon.Parse(f)
	if err != nil {
		return err
	}

	return sql.ExportSQLite(os.Stdout, db)
}

func usage() {
	fmt.Fprintln(os.Stderr, "SQLON (Phase 1)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "    sqlon to-sql <file.sqlon>")
	fmt.Fprintln(os.Stderr, "")
}
