package kafkahelper

import (
	"context"

	"go.uber.org/zap"
)

type Client struct {
	log     zap.Logger
	ctx     context.Context
	brokers []string
}

func New(ctx context.Context, brokers []string, log *zap.Logger) *Client {
	return &Client{ctx: ctx, brokers: brokers, log: *log}
}
