package kafkahelper

import (
	"context"
	"time"

	protomsg "github.com/NishLy/go-kafka-boilerplate/proto"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Producer wraps the segmentio writer
type Producer struct {
	Config *ProducerConfig
	Client *Client
}

type ProducerConfig struct {
	kafka.Writer
}

// NewProducer initializes a new callable producer with robust defaults.
func NewProducer(client *Client, topic string, config *ProducerConfig) *Producer {
	// 1. Connection & Routing Configurations
	if config.Addr == nil {
		config.Addr = kafka.TCP(client.brokers...)
	}
	if config.Topic == "" {
		config.Topic = topic
	}
	if config.Balancer == nil {
		config.Balancer = &kafka.Hash{}
	}

	// 2. Reliability & Performance Configurations
	if config.MaxAttempts == 0 {
		config.MaxAttempts = 3
	}
	if config.RequiredAcks == 0 {
		config.RequiredAcks = kafka.RequireAll
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}

	// 3. Compression Configuration
	if config.Compression == 0 {
		config.Compression = kafka.Zstd
	}

	// 4. Observability
	if config.Logger == nil {
		config.Logger = kafka.LoggerFunc(func(msg string, args ...interface{}) {
			client.log.Info(msg, zap.Any("args", args))
		})
	}

	return &Producer{
		Client: client,
		Config: config,
	}
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {

	envelope := protomsg.Envelope{
		Id:        uuid.New().String(),
		CreatedAt: timestamppb.Now(),
		Payload:   value,
	}

	envelopeBytes, err := proto.Marshal(&envelope)
	if err != nil {
		p.Client.log.Error("Failed to marshal protobuf envelope", zap.Error(err))
		return err
	}

	err = p.Config.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: envelopeBytes,
	})

	if err != nil {
		p.Client.log.Error("Failed to publish message with key "+key, zap.Error(err))
		return err
	}

	return nil
}

// Close cleans up the connection
func (p *Producer) Close() error {
	p.Client.log.Info("Closing producer", zap.String("topic", p.Config.Topic))
	return p.Config.Close()
}
