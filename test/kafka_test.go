package test

import (
	"context"
	"encoding/json"
	"testing"

	kafkahelper "github.com/NishLy/go-kafka-boilerplate/pkg"
	protomsg "github.com/NishLy/go-kafka-boilerplate/proto"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var brokers = []string{"kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"} // example broker address, adjust as needed

func TestAddQueue(t *testing.T) {
	// Create a new email queue item
	Intr := protomsg.EmailRequest{
		ProjectId:  "project_123",
		TemplateId: "welcome_email",
		Recipient:  "test@example.com",
		Variables: map[string]string{
			"first_name": "John",
			"last_name":  "Doe",
		},
	}

	emailQueueItemBytes, err := proto.Marshal(&Intr)
	if err != nil {
		t.Fatalf("Failed to marshal email queue item: %v", err)
	}

	logger := zap.NewExample()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaClient := kafkahelper.New(ctx, brokers, logger) // example broker address, adjust as needed

	producer := kafkahelper.NewProducer(kafkaClient, "email_queue", &kafkahelper.ProducerConfig{})

	err = producer.Publish(ctx, "TEST_KEY", emailQueueItemBytes)

	if err != nil {
		t.Fatalf("Failed to publish email queue item: %v", err)
	}

}

func TestListQueues(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaClient := kafkahelper.New(ctx, brokers, logger) // example broker address, adjust as needed

	consumer := kafkahelper.NewConsumer(kafkaClient, kafkahelper.ConsumerConfig{
		ReaderConfig: kafka.ReaderConfig{
			// Brokers: brokers,
			GroupID: "email_queue_group",
			Topic:   "email_queue",
		},
		MaxMessageFailures: 5,
	})

	t.Log("Starting Consumer")

	consumer.Consume(func(key []byte, payload json.RawMessage) error {
		var envelope protomsg.EmailRequest
		err := proto.Unmarshal(payload, &envelope)
		if err != nil {
			t.Errorf("Failed to unmarshal message payload: %v", err)
			return err
		}
		t.Logf("Received message: %+v", &envelope)
		return nil
	})
}
