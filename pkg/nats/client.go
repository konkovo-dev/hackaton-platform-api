package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Client struct {
	conn   *nats.Conn
	logger *slog.Logger
}

func NewClient(cfg *Config, logger *slog.Logger) (*Client, error) {
	conn, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("connected to NATS", "url", cfg.URL)

	return &Client{
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, subject string, data []byte) error {
	if err := c.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}
	return nil
}

func (c *Client) Subscribe(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to NATS: %w", err)
	}
	return sub, nil
}

func (c *Client) QueueSubscribe(subject, queue string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	sub, err := c.conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe to NATS: %w", err)
	}
	return sub, nil
}
