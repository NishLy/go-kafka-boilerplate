package pkg

import "context"

type Logger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
}

type KafkaClient struct {
	log     Logger
	ctx     context.Context
	brokers []string
}

func New(l Logger, ctx context.Context, brokers []string) *KafkaClient {
	return &KafkaClient{log: l, ctx: ctx, brokers: brokers}
}
