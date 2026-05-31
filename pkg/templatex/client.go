package templatex

import (
	"context"
	"sync"
)

type Client struct {
	cfg     Config
	metrics Metrics
	mu      sync.Mutex
	closed  bool
}

func New(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Client{cfg: cfg, metrics: options.metrics}, nil
}

func (c *Client) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	return nil
}
