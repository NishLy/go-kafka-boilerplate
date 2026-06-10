package kafkahelper

import (
	"encoding/json"

	protomsg "github.com/NishLy/go-kafka-boilerplate/proto"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ConsumerConfig struct {
	kafka.ReaderConfig
	MaxMessageFailures int32
}

type Consumer struct {
	Client    *Client
	Config    ConsumerConfig
	onFailure func(key []byte, e *protomsg.Envelope, err error)
}

type KafkaConsumer interface {
	Consume(handler func(key []byte, payload json.RawMessage) error)
	OnFailure(func(key []byte, e *protomsg.Envelope, err error))
}

func NewConsumer(client *Client, config ConsumerConfig) KafkaConsumer {
	config.Brokers = client.brokers

	if config.MaxMessageFailures == 0 {
		config.MaxMessageFailures = 5
	}

	return &Consumer{
		Client: client,
		Config: config,
	}
}

// OnFailure sets a callback function that will be called when a job fails to procceds successfully. The callback receives the message key, the job details, and the error that occurred.
func (c *Consumer) OnFailure(callback func(key []byte, e *protomsg.Envelope, err error)) {
	c.onFailure = callback
}

// Consume starts consuming messages from the configured Kafka topic and processes them using the provided handler function. (block until context is cancelled)
func (c *Consumer) Consume(handler func(key []byte, payload json.RawMessage) error) {
	if handler == nil {
		c.Client.log.Warn("No handler provided for Consume, exiting")
		return
	}

	reader := kafka.NewReader(c.Config.ReaderConfig)

	// Handle graceful shutdown
	defer func() {
		if err := reader.Close(); err != nil {
			c.Client.log.Error("Error closing reader", zap.Error(err))
		}
	}()

	// Ensure cleanup when the context is cancelled or function exits
	go func() {
		<-c.Client.ctx.Done()
		if err := reader.Close(); err != nil {
			c.Client.log.Error("Error closing reader on context cancellation", zap.Error(err))
		}
	}()

	c.Client.log.Info("Consumer started", zap.String("topic", c.Config.Topic), zap.String("group", c.Config.GroupID))

	for {
		m, err := reader.FetchMessage(c.Client.ctx)
		if err != nil {
			if c.Client.ctx.Err() != nil {
				return
			}
			c.Client.log.Error("Fetch error", zap.Error(err))
			continue
		}

		// Deserialize message into Envelope struct
		var envelope protomsg.Envelope
		if err := proto.Unmarshal(m.Value, &envelope); err != nil {
			c.Client.log.Error("Failed to unmarshal message", zap.Error(err))

			if c.onFailure != nil {
				c.onFailure(m.Key, nil, err)
			}

			continue
		}

		// Process first
		if err := handler(m.Key, envelope.Payload); err != nil {
			c.Client.log.Error("Handler error", zap.Error(err))

			if c.onFailure != nil {
				c.onFailure(m.Key, &envelope, err)
			}

			continue
		}

		if err := reader.CommitMessages(c.Client.ctx, m); err != nil {
			c.Client.log.Error("Commit error", zap.Error(err))
		}
	}
}
