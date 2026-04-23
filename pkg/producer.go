package pkg

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer wraps the segmentio writer
type KafkaProducer struct {
	Writer *kafka.Writer
	Client *KafkaClient
}

// NewProducer initializes a new callable producer
func NewProducer(client *KafkaClient, topic string) *KafkaProducer {

	return &KafkaProducer{
		Client: client,
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(client.brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{}, // Ensures same Key goes to same Partition
			MaxAttempts:  3,
			RequiredAcks: kafka.RequireAll, // Strongest consistency (waits for all replicas)
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Publish is your callable function
func (p *KafkaProducer) Publish(key string, value string) error {

	var job = Job{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Payload:   json.RawMessage(value),
		Retries:   0,
		Topic:     p.Writer.Topic,
	}

	var jobBytes, err = json.Marshal(job)

	if err != nil {
		p.Client.log.Errorf("Failed to marshal job: %v", err)
		return err
	}

	err = p.Writer.WriteMessages(p.Client.ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(jobBytes),
	})

	if err != nil {
		p.Client.log.Errorf("Failed to publish message: %v", err)
		return err
	}

	return nil
}

// Close cleans up the connection
func (p *KafkaProducer) Close() error {
	p.Client.log.Infof("Closing producer for topic: %s", p.Writer.Topic)
	return p.Writer.Close()
}
