# Go Kafka Boilerplate

A minimal Kafka producer/consumer boilerplate in Go using `segmentio/kafka-go` with a centralized `KafkaClient`.

## Project Structure

```text
go-kafka-boilerplate/
├── go.mod
├── go.sum
└── pkg/
    ├── client.go
    ├── consumer.go
    └── producer.go
```

## Core Components

- `pkg/client.go`
  - Defines `KafkaClient` that stores logger, context, and brokers
  - Exposes `New(logger, ctx, brokers)` constructor
  - Must be created first and passed to producer/consumer

- `pkg/producer.go`
  - Defines `KafkaProducer`
  - Exposes `NewProducer(client, topic)`
  - `Publish(key, value)` wraps payload into a `Job` envelope (with UUID and metadata) before writing to Kafka
  - `Close()` closes producer writer

- `pkg/consumer.go`
  - Defines `Job` model
  - Exposes:
    - `StartConsumer(client, topic, groupID, maxRetries, handler, onFailure)`
  - Reads Kafka message, unmarshals into `Job`, passes `job.Payload` to `handler`, and commits on success

## Dependencies

Main runtime dependencies:

- `github.com/segmentio/kafka-go`
- `github.com/google/uuid`

## Quick Usage

```go
package main

import (
    "context"
    "encoding/json"
    "time"

    "github.com/NishLy/go-kafka-boilerplate/pkg"
    "go.uber.org/zap"
)

func main() {
    brokers := []string{"localhost:9092"}
    topic := "events"
    groupID := "demo-group"
    maxRetries := 3

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    logger, _ := zap.NewProduction()
    sugar := logger.Sugar()
    defer logger.Sync()

    // 1) Declare Kafka client first
    client := pkg.New(sugar, ctx, brokers)

    // 2) Use client to create producer
    p := pkg.NewProducer(client, topic)
    defer p.Close()

    // value must be valid JSON string because producer stores it as json.RawMessage
    _ = p.Publish("user-1", `{"event":"hello from producer"}`)

    // 3) Use same client to start consumer
    go pkg.StartConsumer(
        client,
        topic,
        groupID,
        maxRetries,
        func(key []byte, payload json.RawMessage) error {
            sugar.Infof("received key=%s payload=%s", string(key), string(payload))
            return nil
        },
        func(key []byte, job pkg.Job, err error) {
            sugar.Errorf("failed key=%s retries=%d err=%v", string(key), job.Retries, err)
        },
    )

    time.Sleep(5 * time.Second)
}
```

## Key Design Pattern

Always declare `KafkaClient` first:

```go
client := pkg.New(logger, ctx, brokers)
```

Then pass that client into producer and consumer so they share the same context, logger, and broker configuration.
