package replay

import (
	"fmt"
)

func Prepare(events []Event, defaultTopic string) ([]Message, Summary, error) {
	messages := make([]Message, 0, len(events))
	summary := Summary{
		Events: len(events),
		Topics: make(map[string]int),
	}

	for index, event := range events {
		topic := event.Topic
		if topic == "" {
			topic = defaultTopic
		}
		if topic == "" {
			return nil, Summary{}, fmt.Errorf("event %d has no topic and --topic was not provided", index+1)
		}

		message := Message{
			Topic:     topic,
			Key:       event.Key,
			EventTime: event.EventTime,
			Headers:   event.Headers,
			Value:     event.Payload,
		}
		messages = append(messages, message)
		summary.Topics[topic]++

		if summary.FirstEventTime.IsZero() || event.EventTime.Before(summary.FirstEventTime) {
			summary.FirstEventTime = event.EventTime
		}
		if summary.LastEventTime.IsZero() || event.EventTime.After(summary.LastEventTime) {
			summary.LastEventTime = event.EventTime
		}
	}

	return messages, summary, nil
}
