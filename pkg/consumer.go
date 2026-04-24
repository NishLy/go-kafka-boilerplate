package pkg

import (
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Topic      string
	GroupID    string
	MaxRetries int
}

type Job struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	Retries   int             `json:"retries"`
	CreatedAt time.Time       `json:"created_at"`
	Payload   json.RawMessage `json:"payload"`
}

type kafkaConsumer struct {
	Client    *KafkaClient
	Config    kafka.ReaderConfig
	onFailure func(key []byte, job Job, err error)
}

type KafkaConsumer interface {
	Consume(handler func(key []byte, payload json.RawMessage) error)
	OnFailure(func(key []byte, job Job, err error))
}

func NewConsumer(client *KafkaClient, config kafka.ReaderConfig) KafkaConsumer {
	return &kafkaConsumer{
		Client: client,
		Config: config,
	}
}

func (c *kafkaConsumer) OnFailure(callback func(key []byte, job Job, err error)) {
	c.onFailure = callback
}

func (c *kafkaConsumer) Consume(handler func(key []byte, payload json.RawMessage) error) {
	if handler == nil {
		c.Client.log.Warnf("No handler provided for Consume, exiting")
		return
	}

	reader := kafka.NewReader(c.Config)

	// Handle graceful shutdown
	defer func() {
		if err := reader.Close(); err != nil {
			c.Client.log.Errorf("Error closing reader: %v", err)
		}
	}()

	// Ensure cleanup when the context is cancelled or function exits
	go func() {
		<-c.Client.ctx.Done()
		reader.Close()
	}()

	c.Client.log.Infof("Consumer started for topic: %s, group: %s", c.Config.Topic, c.Config.GroupID)

	for {
		m, err := reader.FetchMessage(c.Client.ctx)
		if err != nil {
			if c.Client.ctx.Err() != nil {
				return
			}
			c.Client.log.Errorf("Fetch error: %v", err)
			continue
		}

		// Deserialize message into Job struct
		var job Job
		if err := json.Unmarshal(m.Value, &job); err != nil {
			c.Client.log.Errorf("JSON unmarshal error: %v", err)
			// Optionally, commit the message to skip it
			if err := reader.CommitMessages(c.Client.ctx, m); err != nil {
				c.Client.log.Errorf("Commit error for malformed message: %v", err)
			}
			continue
		}

		// Process first
		if err := handler(m.Key, job.Payload); err != nil {
			c.Client.log.Errorf("Handler error: %v", err)

			// Increment retry count and optionally call onFailure callback
			job.Retries++

			if c.onFailure != nil {
				c.onFailure(m.Key, job, err)
			}
			continue
		}

		if err := reader.CommitMessages(c.Client.ctx, m); err != nil {
			c.Client.log.Errorf("Commit error: %v", err)
		}
	}
}
