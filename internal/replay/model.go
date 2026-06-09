package replay

import (
	"encoding/json"
	"sort"
	"time"
)

type Event struct {
	Key       string            `json:"key"`
	Topic     string            `json:"topic,omitempty"`
	EventTime time.Time         `json:"event_time"`
	Headers   map[string]string `json:"headers,omitempty"`
	Payload   json.RawMessage   `json:"payload"`
}

type Message struct {
	Topic     string            `json:"topic"`
	Key       string            `json:"key"`
	EventTime time.Time         `json:"event_time"`
	Headers   map[string]string `json:"headers,omitempty"`
	Value     json.RawMessage   `json:"value"`
}

type Summary struct {
	Events         int            `json:"events"`
	Topics         map[string]int `json:"topics"`
	FirstEventTime time.Time      `json:"first_event_time"`
	LastEventTime  time.Time      `json:"last_event_time"`
}

func (s Summary) SortedTopics() []string {
	topics := make([]string, 0, len(s.Topics))
	for topic := range s.Topics {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}
