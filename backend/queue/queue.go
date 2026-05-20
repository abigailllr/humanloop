package queue

import (
	"context"
	"encoding/json"
	"os"

	"github.com/redis/go-redis/v9"
)

const jobsKey = "humanloop:pipeline:jobs"

type Client struct {
	rdb *redis.Client
}

func New() *Client {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}
	return &Client{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (c *Client) Push(ctx context.Context, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.rdb.RPush(ctx, jobsKey, b).Err()
}

func (c *Client) Pop(ctx context.Context) ([]byte, error) {
	result, err := c.rdb.BLPop(ctx, 0, jobsKey).Result()
	if err != nil {
		return nil, err
	}
	return []byte(result[1]), nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Depth(ctx context.Context) int64 {
	n, _ := c.rdb.LLen(ctx, jobsKey).Result()
	return n
}
