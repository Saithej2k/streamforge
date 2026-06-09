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
	"github.com/saithej2k/streamforge/internal/replay"
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
	case "replay":
		return runReplay(args[1:], stdout)
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

func runReplay(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("replay", flag.ContinueOnError)
	flags.SetOutput(stdout)

	eventsPath := flags.String("events", "", "path to replay events JSONL")
	topic := flags.String("topic", "", "default Kafka topic when an event does not include one")
	brokers := flags.String("brokers", "localhost:9092", "comma-delimited Kafka broker addresses")
	clientID := flags.String("client-id", "streamforge-replay", "Kafka client id")
	timeout := flags.Duration("timeout", 30*time.Second, "Kafka write timeout")
	dryRun := flags.Bool("dry-run", false, "validate and summarize events without writing to Kafka")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if *eventsPath == "" {
		return errors.New("replay requires --events")
	}

	events, err := replay.LoadEvents(*eventsPath)
	if err != nil {
		return err
	}

	messages, summary, err := replay.Prepare(events, *topic)
	if err != nil {
		return err
	}

	if *dryRun {
		writeReplaySummary(stdout, summary, messages)
		return nil
	}

	opts := replay.KafkaOptions{
		Brokers:  splitCSV(*brokers),
		ClientID: *clientID,
		Timeout:  *timeout,
	}
	if len(opts.Brokers) == 0 {
		return errors.New("replay requires at least one Kafka broker")
	}
	if err := replay.PublishKafka(messages, opts); err != nil {
		return err
	}
	writeReplaySummary(stdout, summary, messages)
	return nil
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

func writeReplaySummary(w io.Writer, summary replay.Summary, messages []replay.Message) {
	fmt.Fprintln(w, "StreamForge Replay Summary")
	fmt.Fprintf(w, "events: %d\n", summary.Events)
	fmt.Fprintf(w, "first_event_time: %s\n", summary.FirstEventTime.Format(time.RFC3339))
	fmt.Fprintf(w, "last_event_time: %s\n", summary.LastEventTime.Format(time.RFC3339))
	fmt.Fprintln(w, "topics:")

	topics := summary.SortedTopics()
	for _, topic := range topics {
		fmt.Fprintf(w, "  %s: %d\n", topic, summary.Topics[topic])
	}

	previewLimit := len(messages)
	if previewLimit > 5 {
		previewLimit = 5
	}
	if previewLimit == 0 {
		return
	}
	fmt.Fprintln(w, "preview:")
	for index := 0; index < previewLimit; index++ {
		message := messages[index]
		fmt.Fprintf(w, "  %d topic=%s key=%s bytes=%d event_time=%s\n",
			index+1,
			message.Topic,
			message.Key,
			len(message.Value),
			message.EventTime.Format(time.RFC3339),
		)
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
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
  streamforge replay --events events.jsonl --topic orders [--brokers localhost:9092] [--dry-run]
  streamforge version

Inputs are newline-delimited JSON records. See examples/orders for a runnable fixture.`)
}
