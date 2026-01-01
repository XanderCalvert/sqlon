package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	case "json-to-sqlon":
		if len(args) < 2 || len(args) > 3 {
			usage()
			os.Exit(2)
		}
		output := ""
		if len(args) == 3 {
			output = args[2]
		} else {
			output = strings.TrimSuffix(args[1], ".json") + ".sqlon"
		}
		if err := runJSONToSQLON(args[1], output); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "sqlon-to-json":
		if len(args) < 2 || len(args) > 3 {
			usage()
			os.Exit(2)
		}
		output := ""
		if len(args) == 3 {
			output = args[2]
		} else {
			output = strings.TrimSuffix(args[1], ".sqlon") + ".json"
		}
		if err := runSQLONToJSON(args[1], output); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "convert-json":
		if len(args) != 2 {
			usage()
			os.Exit(2)
		}
		if err := runConvertJSON(args[1]); err != nil {
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

func runJSONToSQLON(inputPath, outputPath string) error {
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	step := &pipeline.JSONToSQLONStep{}
	output, err := step.Run(input)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, output, 0o644)
}

func runSQLONToJSON(inputPath, outputPath string) error {
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	step := &pipeline.SQLONToJSONStep{}
	output, err := step.Run(input)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, output, 0o644)
}

func runConvertJSON(jsonPath string) error {
	// Read original JSON
	input, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Determine output paths
	absJsonPath, err := filepath.Abs(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to resolve JSON path: %w", err)
	}
	dir := filepath.Dir(absJsonPath)
	baseName := strings.TrimSuffix(filepath.Base(absJsonPath), ".json")

	// SQLON output: examples/sqlon/<name>.sqlon
	sqlonDir := filepath.Join(dir, "..", "sqlon")
	sqlonDir = filepath.Clean(sqlonDir)
	if err := os.MkdirAll(sqlonDir, 0o755); err != nil {
		return fmt.Errorf("failed to create sqlon directory: %w", err)
	}
	sqlonPath := filepath.Join(sqlonDir, baseName+".sqlon")

	// Roundtrip JSON output: examples/json/<name>.roundtrip.json
	roundtripPath := filepath.Join(dir, baseName+".roundtrip.json")

	// Step 1: JSON → SQLON
	fmt.Fprintf(os.Stdout, "Converting JSON → SQLON: %s\n", sqlonPath)
	step1 := &pipeline.JSONToSQLONStep{}
	sqlonOutput, err := step1.Run(input)
	if err != nil {
		return fmt.Errorf("JSON → SQLON conversion failed: %w", err)
	}
	if err := os.WriteFile(sqlonPath, sqlonOutput, 0o644); err != nil {
		return fmt.Errorf("failed to write SQLON file: %w", err)
	}

	// Step 2: SQLON → JSON
	fmt.Fprintf(os.Stdout, "Converting SQLON → JSON: %s\n", roundtripPath)
	step2 := &pipeline.SQLONToJSONStep{}
	jsonOutput, err := step2.Run(sqlonOutput)
	if err != nil {
		return fmt.Errorf("SQLON → JSON conversion failed: %w", err)
	}
	if err := os.WriteFile(roundtripPath, jsonOutput, 0o644); err != nil {
		return fmt.Errorf("failed to write roundtrip JSON file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✓ Original JSON: %s\n", jsonPath)
	fmt.Fprintf(os.Stdout, "✓ SQLON: %s\n", sqlonPath)
	fmt.Fprintf(os.Stdout, "✓ Roundtrip JSON: %s\n", roundtripPath)
	return nil
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
	fmt.Fprintln(os.Stderr, "    sqlon json-to-sqlon <input.json> [output.sqlon]")
	fmt.Fprintln(os.Stderr, "    sqlon sqlon-to-json <input.sqlon> [output.json]")
	fmt.Fprintln(os.Stderr, "    sqlon convert-json <input.json>")
	fmt.Fprintln(os.Stderr, "    sqlon roundtrip <file.json>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "convert-json: Converts JSON → SQLON → JSON, preserving original")
	fmt.Fprintln(os.Stderr, "             Outputs: examples/sqlon/<name>.sqlon")
	fmt.Fprintln(os.Stderr, "                      examples/json/<name>.roundtrip.json")
	fmt.Fprintln(os.Stderr, "")
}
