package replay

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPrepareAppliesDefaultTopic(t *testing.T) {
	eventTime := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	events := []Event{
		{Key: "ord-1001", EventTime: eventTime, Payload: json.RawMessage(`{"id":"ord-1001"}`)},
		{Key: "ord-1002", Topic: "priority-orders", EventTime: eventTime.Add(time.Second), Payload: json.RawMessage(`{"id":"ord-1002"}`)},
	}

	messages, summary, err := Prepare(events, "orders")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if messages[0].Topic != "orders" {
		t.Fatalf("default topic = %s", messages[0].Topic)
	}
	if messages[1].Topic != "priority-orders" {
		t.Fatalf("event topic = %s", messages[1].Topic)
	}
	if summary.Topics["orders"] != 1 || summary.Topics["priority-orders"] != 1 {
		t.Fatalf("topics = %#v", summary.Topics)
	}
}

func TestPrepareRequiresTopic(t *testing.T) {
	_, _, err := Prepare([]Event{{
		Key:       "ord-1001",
		EventTime: time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC),
		Payload:   json.RawMessage(`{"id":"ord-1001"}`),
	}}, "")
	if err == nil {
		t.Fatal("expected topic error")
	}
}
