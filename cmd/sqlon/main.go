package main

import (
	"fmt"
	"os"
	"path/filepath"

	"sqlon/internal/format/sql"
	"sqlon/internal/format/sqlon"
	"sqlon/internal/pipeline"
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
	case "roundtrip":
		if len(args) != 2 {
			usage()
			os.Exit(2)
		}
		if err := runRoundtrip(args[1]); err != nil {
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

func runRoundtrip(path string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	outDir := "out"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	logPath := filepath.Join(outDir, "pipeline.log.jsonl")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	r := &pipeline.Runner{
		OutDir: outDir,
		LogW:   logFile,
	}

	steps := []pipeline.Step{
		&pipeline.JSONToSQLONStep{},
		&pipeline.SQLONToSQLStep{},
		&pipeline.SQLToSQLONStep{},
		&pipeline.SQLONToJSONStep{},
	}

	_, err = r.Run(steps, input, "")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Artefacts written to:", outDir)
	fmt.Fprintln(os.Stdout, "Log written to:", logPath)
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "SQLON (Phase 1)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "    sqlon to-sql <file.sqlon>")
	fmt.Fprintln(os.Stderr, "    sqlon roundtrip <file.json>")
	fmt.Fprintln(os.Stderr, "")
}
