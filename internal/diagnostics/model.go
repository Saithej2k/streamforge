package diagnostics

import (
	"time"
)

type Trace struct {
	Topic     string `json:"topic,omitempty"`
	Partition int    `json:"partition,omitempty"`
	Offset    int64  `json:"offset,omitempty"`
	Key       string `json:"key,omitempty"`
	Producer  string `json:"producer,omitempty"`
	RunID     string `json:"run_id,omitempty"`
}

type ExpectedRecord struct {
	RecordID    string    `json:"record_id"`
	EventTime   time.Time `json:"event_time"`
	PayloadHash string    `json:"payload_hash"`
	Trace       Trace     `json:"trace,omitempty"`
}

type ActualWrite struct {
	RecordID    string    `json:"record_id"`
	EventTime   time.Time `json:"event_time"`
	WrittenAt   time.Time `json:"written_at"`
	PayloadHash string    `json:"payload_hash"`
	SnapshotID  string    `json:"snapshot_id,omitempty"`
	FilePath    string    `json:"file_path,omitempty"`
	Trace       Trace     `json:"trace,omitempty"`
}

type Options struct {
	LateAfter time.Duration `json:"late_after"`
}

type Summary struct {
	ExpectedRecords   int `json:"expected_records"`
	ActualWrites      int `json:"actual_writes"`
	MatchedRecords    int `json:"matched_records"`
	MissingRecords    int `json:"missing_records"`
	DuplicateRecords  int `json:"duplicate_records"`
	LateRecords       int `json:"late_records"`
	HashMismatches    int `json:"hash_mismatches"`
	UnexpectedRecords int `json:"unexpected_records"`
}

type FindingKind string

const (
	FindingMissing      FindingKind = "missing"
	FindingDuplicate    FindingKind = "duplicate"
	FindingLate         FindingKind = "late"
	FindingHashMismatch FindingKind = "hash_mismatch"
	FindingUnexpected   FindingKind = "unexpected"
)

type TraceContext struct {
	Role        string    `json:"role"`
	Topic       string    `json:"topic,omitempty"`
	Partition   int       `json:"partition,omitempty"`
	Offset      int64     `json:"offset,omitempty"`
	Key         string    `json:"key,omitempty"`
	Producer    string    `json:"producer,omitempty"`
	RunID       string    `json:"run_id,omitempty"`
	EventTime   time.Time `json:"event_time,omitempty"`
	WrittenAt   time.Time `json:"written_at,omitempty"`
	PayloadHash string    `json:"payload_hash,omitempty"`
	SnapshotID  string    `json:"snapshot_id,omitempty"`
	FilePath    string    `json:"file_path,omitempty"`
}

type Finding struct {
	Kind     FindingKind    `json:"kind"`
	Severity string         `json:"severity"`
	RecordID string         `json:"record_id"`
	Message  string         `json:"message"`
	Contexts []TraceContext `json:"contexts"`
}

type Report struct {
	GeneratedAt time.Time `json:"generated_at"`
	Options     Options   `json:"options"`
	Summary     Summary   `json:"summary"`
	Findings    []Finding `json:"findings"`
}
