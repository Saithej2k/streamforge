package diagnostics

import (
	"testing"
	"time"
)

func TestCompareFindsReplayFailures(t *testing.T) {
	base := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	expected := []ExpectedRecord{
		{RecordID: "ord-1001", EventTime: base, PayloadHash: "h1", Trace: Trace{Topic: "orders", Partition: 0, Offset: 1, Key: "ord-1001"}},
		{RecordID: "ord-1002", EventTime: base.Add(time.Second), PayloadHash: "h2", Trace: Trace{Topic: "orders", Partition: 0, Offset: 2, Key: "ord-1002"}},
		{RecordID: "ord-1003", EventTime: base.Add(2 * time.Second), PayloadHash: "h3", Trace: Trace{Topic: "orders", Partition: 0, Offset: 3, Key: "ord-1003"}},
		{RecordID: "ord-1004", EventTime: base.Add(3 * time.Second), PayloadHash: "h4", Trace: Trace{Topic: "orders", Partition: 0, Offset: 4, Key: "ord-1004"}},
	}
	actual := []ActualWrite{
		{RecordID: "ord-1001", EventTime: base, WrittenAt: base.Add(time.Minute), PayloadHash: "h1", SnapshotID: "101", FilePath: "warehouse/orders/data-1.parquet"},
		{RecordID: "ord-1002", EventTime: base.Add(time.Second), WrittenAt: base.Add(2 * time.Minute), PayloadHash: "h2", SnapshotID: "101", FilePath: "warehouse/orders/data-1.parquet"},
		{RecordID: "ord-1002", EventTime: base.Add(time.Second), WrittenAt: base.Add(3 * time.Minute), PayloadHash: "h2", SnapshotID: "102", FilePath: "warehouse/orders/data-2.parquet"},
		{RecordID: "ord-1003", EventTime: base.Add(2 * time.Second), WrittenAt: base.Add(20 * time.Minute), PayloadHash: "h3", SnapshotID: "103", FilePath: "warehouse/orders/data-3.parquet"},
		{RecordID: "ord-1004", EventTime: base.Add(3 * time.Second), WrittenAt: base.Add(time.Minute), PayloadHash: "changed", SnapshotID: "104", FilePath: "warehouse/orders/data-4.parquet"},
		{RecordID: "ord-9999", EventTime: base, WrittenAt: base.Add(time.Minute), PayloadHash: "hx", SnapshotID: "105", FilePath: "warehouse/orders/data-5.parquet"},
	}

	report := Compare(expected, actual, Options{LateAfter: 5 * time.Minute})

	if report.Summary.ExpectedRecords != 4 {
		t.Fatalf("expected record count = %d", report.Summary.ExpectedRecords)
	}
	if report.Summary.DuplicateRecords != 1 {
		t.Fatalf("duplicate records = %d", report.Summary.DuplicateRecords)
	}
	if report.Summary.LateRecords != 1 {
		t.Fatalf("late records = %d", report.Summary.LateRecords)
	}
	if report.Summary.HashMismatches != 1 {
		t.Fatalf("hash mismatches = %d", report.Summary.HashMismatches)
	}
	if report.Summary.UnexpectedRecords != 1 {
		t.Fatalf("unexpected records = %d", report.Summary.UnexpectedRecords)
	}
	if len(report.Findings) != 4 {
		t.Fatalf("findings = %d", len(report.Findings))
	}
}

func TestCompareFindsMissingRecords(t *testing.T) {
	eventTime := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	report := Compare(
		[]ExpectedRecord{{RecordID: "ord-1001", EventTime: eventTime, PayloadHash: "h1"}},
		nil,
		Options{LateAfter: time.Minute},
	)

	if report.Summary.MissingRecords != 1 {
		t.Fatalf("missing records = %d", report.Summary.MissingRecords)
	}
	if report.Findings[0].Kind != FindingMissing {
		t.Fatalf("finding kind = %s", report.Findings[0].Kind)
	}
}
