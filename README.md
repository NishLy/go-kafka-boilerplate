# Go Kafka Boilerplate

A minimal Kafka producer/consumer boilerplate in Go using `segmentio/kafka-go` with a centralized `KafkaClient` that encapsulates context, logger, and broker configuration.

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
  - Defines `KafkaClient` struct: holds logger, context, and brokers
  - Exposes `New(logger, ctx, brokers)` constructor
  - Central point for all Kafka operations

- `pkg/producer.go`
  - Defines `KafkaProducer` struct
  - Exposes `NewProducer(client, topic)` constructor
  - Provides `Publish(key, value)` and `Close()` methods

- `pkg/consumer.go`
  - Defines `MessageHandler` callback type: `func(client, key, value) error`
  - Exposes `StartConsumer(client, topic, groupID, handler)`
  - Reads messages continuously and passes each message to your handler

## Dependencies

Main runtime dependencies include:

- `github.com/segmentio/kafka-go`

## Quick Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/NishLy/go-kafka-boilerplate/pkg"
    "go.uber.org/zap"
)

func main() {
    brokers := []string{"localhost:9092"}
    topic := "events"
    groupID := "demo-group"

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Initialize logger
    logger, _ := zap.NewProduction()
    sugar := logger.Sugar()
    defer logger.Sync()

    // Create KafkaClient first
    client := pkg.New(sugar, ctx, brokers)

    // Producer
    p := pkg.NewProducer(client, topic)
    defer p.Close()

    _ = p.Publish("user-1", "hello from producer")

    // Consumer
    go pkg.StartConsumer(client, topic, groupID, func(c *pkg.KafkaClient, key, value []byte) error {
        fmt.Printf("received key=%s value=%s\n", string(key), string(value))
        return nil
    })

    time.Sleep(5 * time.Second)
}
```

## Key Design Pattern

Always initialize `KafkaClient` first:

```go
client := pkg.New(logger, ctx, brokers)
```

Then pass the client to producer and consumer. This ensures:

- Unified context management
- Consistent logging across operations
- Centralized broker configuration
