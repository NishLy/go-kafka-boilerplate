package test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/NishLy/go-kafka-boilerplate/pkg"
	"go.uber.org/zap"
)

func TestAddQueue(t *testing.T) {
	// Create a new email queue item
	Intr := map[string]interface{}{
		"ProjectID":  "project_123",
		"TemplateID": "welcome_email",
		"Recipient":  "test@example.com",
		"Variables": map[string]interface{}{
			"name": "John Doe",
		},
		"AttachmentPaths": []string{"path/to/attachment1", "path/to/attachment2"},
	}

	emailQueueItemBytes, err := json.Marshal(Intr)

	if err != nil {
		t.Fatalf("Failed to marshal email queue item: %v", err)
	}

	logger := zap.NewExample()
	sugar := logger.Sugar()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaClient := pkg.New(sugar, ctx, []string{"localhost:9093"}) // example broker address, adjust as needed

	producer := pkg.NewProducer(kafkaClient, "email_queue")

	err = producer.Publish("TEST_KEY", string(emailQueueItemBytes))

	if err != nil {
		t.Fatalf("Failed to publish email queue item: %v", err)
	}

}
