package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/saithej2k/streamforge/internal/diagnostics"
)

const version = "0.1.0"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeUsage(stdout)
		return nil
	}

	switch args[0] {
	case "diagnose":
		return runDiagnose(args[1:], stdout)
	case "version":
		fmt.Fprintf(stdout, "streamforge %s\n", version)
		return nil
	case "help", "-h", "--help":
		writeUsage(stdout)
		return nil
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		writeUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runDiagnose(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("diagnose", flag.ContinueOnError)
	flags.SetOutput(stdout)

	expectedPath := flags.String("expected", "", "path to expected records JSONL")
	actualPath := flags.String("actual", "", "path to actual Iceberg writes JSONL")
	outPath := flags.String("out", "", "write report JSON to this path")
	format := flags.String("format", "text", "report format: text or json")
	lateAfter := flags.Duration("late-after", 5*time.Minute, "mark writes later than this duration as late")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if *expectedPath == "" || *actualPath == "" {
		return errors.New("diagnose requires --expected and --actual")
	}

	expected, err := diagnostics.LoadExpected(*expectedPath)
	if err != nil {
		return err
	}
	actual, err := diagnostics.LoadActual(*actualPath)
	if err != nil {
		return err
	}

	report := diagnostics.Compare(expected, actual, diagnostics.Options{LateAfter: *lateAfter})

	if *outPath != "" {
		if err := writeJSON(*outPath, report); err != nil {
			return err
		}
	}

	switch strings.ToLower(*format) {
	case "json":
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case "text":
		writeTextReport(stdout, report)
		return nil
	default:
		return fmt.Errorf("unsupported format %q", *format)
	}
}

func writeJSON(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeTextReport(w io.Writer, report diagnostics.Report) {
	fmt.Fprintln(w, "StreamForge Diagnostics Report")
	fmt.Fprintf(w, "generated_at: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "late_after: %s\n\n", report.Options.LateAfter)

	fmt.Fprintln(w, "Summary")
	fmt.Fprintf(w, "  expected_records: %d\n", report.Summary.ExpectedRecords)
	fmt.Fprintf(w, "  actual_writes: %d\n", report.Summary.ActualWrites)
	fmt.Fprintf(w, "  matched_records: %d\n", report.Summary.MatchedRecords)
	fmt.Fprintf(w, "  missing_records: %d\n", report.Summary.MissingRecords)
	fmt.Fprintf(w, "  duplicate_records: %d\n", report.Summary.DuplicateRecords)
	fmt.Fprintf(w, "  late_records: %d\n", report.Summary.LateRecords)
	fmt.Fprintf(w, "  hash_mismatches: %d\n", report.Summary.HashMismatches)
	fmt.Fprintf(w, "  unexpected_records: %d\n", report.Summary.UnexpectedRecords)

	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "\nNo findings.")
		return
	}

	fmt.Fprintln(w, "\nFindings")
	for _, finding := range report.Findings {
		fmt.Fprintf(w, "- [%s] %s %s: %s\n", finding.Severity, finding.Kind, finding.RecordID, finding.Message)
		for _, context := range finding.Contexts {
			fmt.Fprintf(w, "    %s topic=%s partition=%d offset=%d key=%s snapshot=%s file=%s\n",
				context.Role,
				context.Topic,
				context.Partition,
				context.Offset,
				context.Key,
				context.SnapshotID,
				context.FilePath,
			)
		}
	}
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, `StreamForge - lakehouse replay and diagnostics toolkit

Usage:
  streamforge diagnose --expected expected.jsonl --actual actual.jsonl [--late-after 5m] [--format text|json] [--out report.json]
  streamforge version

Inputs are newline-delimited JSON records. See examples/orders for a runnable fixture.`)
}
