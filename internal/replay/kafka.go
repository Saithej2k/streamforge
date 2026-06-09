package replay

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaOptions struct {
	Brokers  []string
	ClientID string
	Timeout  time.Duration
}

func PublishKafka(messages []Message, opts KafkaOptions) error {
	if len(messages) == 0 {
		return nil
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.ClientID == "" {
		opts.ClientID = "streamforge-replay"
	}

	writer := kafka.Writer{
		Addr:                   kafka.TCP(opts.Brokers...),
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		Async:                  false,
		WriteTimeout:           opts.Timeout,
		Transport: &kafka.Transport{
			ClientID: opts.ClientID,
		},
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	kafkaMessages := make([]kafka.Message, 0, len(messages))
	for _, message := range messages {
		kafkaMessages = append(kafkaMessages, kafka.Message{
			Topic:   message.Topic,
			Key:     []byte(message.Key),
			Value:   []byte(message.Value),
			Time:    message.EventTime,
			Headers: kafkaHeaders(message.Headers),
		})
	}

	if err := writer.WriteMessages(ctx, kafkaMessages...); err != nil {
		return fmt.Errorf("publish Kafka replay messages: %w", err)
	}
	return nil
}

func kafkaHeaders(headers map[string]string) []kafka.Header {
	if len(headers) == 0 {
		return nil
	}
	out := make([]kafka.Header, 0, len(headers))
	for key, value := range headers {
		out = append(out, kafka.Header{Key: key, Value: []byte(value)})
	}
	return out
}
