package diagnostics

import (
	"fmt"
	"sort"
	"time"
)

func Compare(expected []ExpectedRecord, actual []ActualWrite, opts Options) Report {
	if opts.LateAfter == 0 {
		opts.LateAfter = 5 * time.Minute
	}

	expectedByID := make(map[string]ExpectedRecord, len(expected))
	expectedIDs := make([]string, 0, len(expected))
	for _, record := range expected {
		if _, exists := expectedByID[record.RecordID]; !exists {
			expectedIDs = append(expectedIDs, record.RecordID)
		}
		expectedByID[record.RecordID] = record
	}
	sort.Strings(expectedIDs)

	actualByID := make(map[string][]ActualWrite, len(actual))
	actualIDs := make([]string, 0, len(actual))
	seenActualID := make(map[string]bool)
	for _, write := range actual {
		if !seenActualID[write.RecordID] {
			actualIDs = append(actualIDs, write.RecordID)
			seenActualID[write.RecordID] = true
		}
		actualByID[write.RecordID] = append(actualByID[write.RecordID], write)
	}
	sort.Strings(actualIDs)
	for recordID := range actualByID {
		sort.Slice(actualByID[recordID], func(left, right int) bool {
			return actualByID[recordID][left].WrittenAt.Before(actualByID[recordID][right].WrittenAt)
		})
	}

	report := Report{
		GeneratedAt: time.Now().UTC(),
		Options:     opts,
		Summary: Summary{
			ExpectedRecords: len(expected),
			ActualWrites:    len(actual),
		},
	}

	for _, recordID := range expectedIDs {
		expectedRecord := expectedByID[recordID]
		writes := actualByID[recordID]
		if len(writes) == 0 {
			report.Summary.MissingRecords++
			report.Findings = append(report.Findings, missingFinding(expectedRecord))
			continue
		}

		report.Summary.MatchedRecords++
		if len(writes) > 1 {
			report.Summary.DuplicateRecords++
			report.Findings = append(report.Findings, duplicateFinding(expectedRecord, writes))
		}

		hashMismatchRecorded := false
		lateRecorded := false
		for _, write := range writes {
			if write.PayloadHash != expectedRecord.PayloadHash && !hashMismatchRecorded {
				report.Summary.HashMismatches++
				report.Findings = append(report.Findings, hashMismatchFinding(expectedRecord, write))
				hashMismatchRecorded = true
			}
			if write.WrittenAt.Sub(write.EventTime) > opts.LateAfter && !lateRecorded {
				report.Summary.LateRecords++
				report.Findings = append(report.Findings, lateFinding(expectedRecord, write, opts.LateAfter))
				lateRecorded = true
			}
		}
	}

	for _, recordID := range actualIDs {
		if _, exists := expectedByID[recordID]; exists {
			continue
		}
		report.Summary.UnexpectedRecords++
		report.Findings = append(report.Findings, unexpectedFinding(actualByID[recordID]))
	}

	sort.SliceStable(report.Findings, func(left, right int) bool {
		if report.Findings[left].RecordID == report.Findings[right].RecordID {
			return report.Findings[left].Kind < report.Findings[right].Kind
		}
		return report.Findings[left].RecordID < report.Findings[right].RecordID
	})
	return report
}

func missingFinding(record ExpectedRecord) Finding {
	return Finding{
		Kind:     FindingMissing,
		Severity: "critical",
		RecordID: record.RecordID,
		Message:  "expected record was not written to the lakehouse",
		Contexts: []TraceContext{expectedContext(record)},
	}
}

func duplicateFinding(record ExpectedRecord, writes []ActualWrite) Finding {
	contexts := []TraceContext{expectedContext(record)}
	for _, write := range writes {
		contexts = append(contexts, actualContext(write))
	}
	return Finding{
		Kind:     FindingDuplicate,
		Severity: "high",
		RecordID: record.RecordID,
		Message:  fmt.Sprintf("record was written %d times", len(writes)),
		Contexts: contexts,
	}
}

func lateFinding(record ExpectedRecord, write ActualWrite, threshold time.Duration) Finding {
	return Finding{
		Kind:     FindingLate,
		Severity: "medium",
		RecordID: record.RecordID,
		Message:  fmt.Sprintf("write landed %s after event time, above %s threshold", write.WrittenAt.Sub(write.EventTime), threshold),
		Contexts: []TraceContext{expectedContext(record), actualContext(write)},
	}
}

func hashMismatchFinding(record ExpectedRecord, write ActualWrite) Finding {
	return Finding{
		Kind:     FindingHashMismatch,
		Severity: "high",
		RecordID: record.RecordID,
		Message:  "actual payload hash does not match expected payload hash",
		Contexts: []TraceContext{expectedContext(record), actualContext(write)},
	}
}

func unexpectedFinding(writes []ActualWrite) Finding {
	contexts := make([]TraceContext, 0, len(writes))
	for _, write := range writes {
		contexts = append(contexts, actualContext(write))
	}
	return Finding{
		Kind:     FindingUnexpected,
		Severity: "medium",
		RecordID: writes[0].RecordID,
		Message:  "lakehouse write has no matching expected replay record",
		Contexts: contexts,
	}
}

func expectedContext(record ExpectedRecord) TraceContext {
	return TraceContext{
		Role:        "expected",
		Topic:       record.Trace.Topic,
		Partition:   record.Trace.Partition,
		Offset:      record.Trace.Offset,
		Key:         record.Trace.Key,
		Producer:    record.Trace.Producer,
		RunID:       record.Trace.RunID,
		EventTime:   record.EventTime,
		PayloadHash: record.PayloadHash,
	}
}

func actualContext(write ActualWrite) TraceContext {
	return TraceContext{
		Role:        "actual",
		Topic:       write.Trace.Topic,
		Partition:   write.Trace.Partition,
		Offset:      write.Trace.Offset,
		Key:         write.Trace.Key,
		Producer:    write.Trace.Producer,
		RunID:       write.Trace.RunID,
		EventTime:   write.EventTime,
		WrittenAt:   write.WrittenAt,
		PayloadHash: write.PayloadHash,
		SnapshotID:  write.SnapshotID,
		FilePath:    write.FilePath,
	}
}
