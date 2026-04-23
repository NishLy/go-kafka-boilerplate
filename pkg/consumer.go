package pkg

import (
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Job struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	Retries   int             `json:"retries"`
	Status    string          `json:"status"` // optional
	CreatedAt time.Time       `json:"created_at"`
	Payload   json.RawMessage `json:"payload"`
}

// StartConsumer encapsulates the Segmentio reader logic
func StartConsumer(client *KafkaClient, topic, groupID string, maxRetries int, handler func(key []byte, payload json.RawMessage) error, onFailure func(key []byte, job Job, err error)) {

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
		m, err := reader.FetchMessage(client.ctx)
		if err != nil {
			if client.ctx.Err() != nil {
				return
			}
			client.log.Errorf("Fetch error: %v", err)
			continue
		}

		// Deserialize message into Job struct
		var job Job
		if err := json.Unmarshal(m.Value, &job); err != nil {
			client.log.Errorf("JSON unmarshal error: %v", err)
			// Optionally, commit the message to skip it
			if err := reader.CommitMessages(client.ctx, m); err != nil {
				client.log.Errorf("Commit error for malformed message: %v", err)
			}
			continue
		}

		// Process first
		if err := handler(m.Key, job.Payload); err != nil {
			client.log.Errorf("Handler error: %v", err)

			// Increment retry count and optionally call onFailure callback
			job.Retries++

			if onFailure != nil {
				onFailure(m.Key, job, err)
			}
			continue
		}

		if err := reader.CommitMessages(client.ctx, m); err != nil {
			client.log.Errorf("Commit error: %v", err)
		}
	}
}
