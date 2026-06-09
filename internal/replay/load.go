package replay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func LoadEvents(path string) ([]Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("%s:%d: decode JSONL: %w", path, lineNumber, err)
		}
		if err := validateEvent(event); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNumber, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return events, nil
}

func validateEvent(event Event) error {
	if event.Key == "" {
		return fmt.Errorf("key is required")
	}
	if event.EventTime.IsZero() {
		return fmt.Errorf("event_time is required")
	}
	if len(event.Payload) == 0 || string(event.Payload) == "null" {
		return fmt.Errorf("payload is required")
	}
	return nil
}
