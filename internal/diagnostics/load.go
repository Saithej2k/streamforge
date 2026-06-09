package diagnostics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func LoadExpected(path string) ([]ExpectedRecord, error) {
	return loadJSONL[ExpectedRecord](path, validateExpected)
}

func LoadActual(path string) ([]ActualWrite, error) {
	return loadJSONL[ActualWrite](path, validateActual)
}

func loadJSONL[T any](path string, validate func(T) error) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var records []T
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record T
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("%s:%d: decode JSONL: %w", path, lineNumber, err)
		}
		if err := validate(record); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNumber, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return records, nil
}

func validateExpected(record ExpectedRecord) error {
	if record.RecordID == "" {
		return fmt.Errorf("record_id is required")
	}
	if record.EventTime.IsZero() {
		return fmt.Errorf("event_time is required")
	}
	if record.PayloadHash == "" {
		return fmt.Errorf("payload_hash is required")
	}
	return nil
}

func validateActual(record ActualWrite) error {
	if record.RecordID == "" {
		return fmt.Errorf("record_id is required")
	}
	if record.EventTime.IsZero() {
		return fmt.Errorf("event_time is required")
	}
	if record.WrittenAt.IsZero() {
		return fmt.Errorf("written_at is required")
	}
	if record.PayloadHash == "" {
		return fmt.Errorf("payload_hash is required")
	}
	return nil
}
