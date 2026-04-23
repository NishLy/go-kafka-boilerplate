package reader

import (
	"context"
	"go-kafka-boilerlate/pkg"

	"github.com/segmentio/kafka-go"
)

// MessageHandler defines the callback signature
type MessageHandler func(ctx context.Context, key, value []byte) error

// StartConsumer encapsulates the Segmentio reader logic
func StartConsumer(ctx context.Context, brokers []string, topic, groupID string, handler MessageHandler) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	// Ensure cleanup when the context is cancelled or function exits
	go func() {
		<-ctx.Done()
		reader.Close()
	}()

	pkg.Logger.Infof("Consumer started for topic: %s, group: %s", topic, groupID)

	for {
		// FetchMessage handles offset commits automatically by default
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			// If context was cancelled, exit gracefully
			if ctx.Err() != nil {
				return
			}
			pkg.Logger.Errorf("Error reading message: %v", err)
			continue
		}

		// Trigger the callback
		if err := handler(ctx, m.Key, m.Value); err != nil {
			pkg.Logger.Errorf("Handler error: %v", err)
			// Decide here if you want to retry or skip
		}
	}
}
