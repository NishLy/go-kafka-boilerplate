package producer

import (
	"context"

	"time"

	"github.com/NishLy/go-kafka-boilerplate/pkg"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer wraps the segmentio writer
type KafkaProducer struct {
	Writer *kafka.Writer
}

// NewProducer initializes a new callable producer
func NewProducer(brokers []string, topic string) *KafkaProducer {
	// check logger is initialized
	if pkg.Logger == nil {
		pkg.InitKafkaLogger()
	}

	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{}, // Ensures same Key goes to same Partition
			MaxAttempts:  3,
			RequiredAcks: kafka.RequireAll, // Strongest consistency (waits for all replicas)
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Publish is your callable function
func (p *KafkaProducer) Publish(ctx context.Context, key string, value string) error {
	err := p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	})

	if err != nil {
		pkg.Logger.Errorf("Failed to publish message: %v", err)
		return err
	}

	return nil
}

// Close cleans up the connection
func (p *KafkaProducer) Close() error {
	return p.Writer.Close()
}
