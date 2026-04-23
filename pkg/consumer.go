package pkg

import (
	"github.com/segmentio/kafka-go"
)

// StartConsumer encapsulates the Segmentio reader logic
func StartConsumer(client *KafkaClient, topic, groupID string, handler func(key, value []byte) error) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: client.brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	// Ensure cleanup when the context is cancelled or function exits
	go func() {
		<-client.ctx.Done()
		reader.Close()
	}()

	client.log.Infof("Consumer started for topic: %s, group: %s", topic, groupID)

	for {
		// FetchMessage handles offset commits automatically by default
		m, err := reader.ReadMessage(client.ctx)
		if err != nil {
			// If context was cancelled, exit gracefully
			if client.ctx.Err() != nil {
				return
			}
			client.log.Errorf("Error reading message: %v", err)
			continue
		}

		// Trigger the callback
		if err := handler(m.Key, m.Value); err != nil {
			client.log.Errorf("Handler error: %v", err)
			// Decide here if you want to retry or skip
		}
	}
}
