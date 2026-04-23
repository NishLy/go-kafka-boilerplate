# Go Kafka Boilerplate

A minimal Kafka producer/consumer boilerplate in Go using `segmentio/kafka-go` with structured logging via `zap`.

## Project Structure

```text
go-kafka-boilerplate/
├── go.mod
├── go.sum
├── intenal/
│   ├── consumer/
│   │   └── consumer.go
│   └── producer/
│       └── producer.go
└── pkg/
    └── log.go
```

## What Each Package Does

- `intenal/producer/producer.go`
  - Defines `KafkaProducer`
  - Creates a configured Kafka writer
  - Exposes `Publish(ctx, key, value)` and `Close()`

- `intenal/consumer/consumer.go`
  - Defines `MessageHandler` callback type
  - Exposes `StartConsumer(ctx, brokers, topic, groupID, handler)`
  - Reads messages continuously and passes each message to your handler

- `pkg/log.go`
  - Initializes global `pkg.Logger` (`zap.SugaredLogger`)
  - Used by producer/consumer for logging

## Dependencies

Main runtime dependencies include:

- `github.com/segmentio/kafka-go`
- `go.uber.org/zap`

## Quick Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    reader "go-kafka-boilerlate/intenal/consumer"
    "go-kafka-boilerlate/intenal/producer"
)

func main() {
    brokers := []string{"localhost:9092"}
    topic := "events"
    groupID := "demo-group"

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Producer
    p := producer.NewProducer(brokers, topic)
    defer p.Close()

    _ = p.Publish(ctx, "user-1", "hello from producer")

    // Consumer
    go reader.StartConsumer(ctx, brokers, topic, groupID, func(ctx context.Context, key, value []byte) error {
        fmt.Printf("received key=%s value=%s\n", string(key), string(value))
        return nil
    })

    time.Sleep(5 * time.Second)
}
```

## Notes

- The current module name in `go.mod` is `go-kafka-boilerlate`, and imports should match it exactly.
