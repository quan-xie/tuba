package redis

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/log"
	"github.com/redis/go-redis/v9"
	"go.opencensus.io/trace"
)

func NewRedis(config *Config) (client *RedisStorage, err error) {
	if config == nil {
		return nil, errors.New("init fail: config is nil")
	}
	client, err = CreateRedisStorage(config)
	return
}

func NewRedisClient(config *Config) (client *RedisStorage) {
	var err error
	client, err = CreateRedisStorage(config)
	if err != nil {
		log.Fatalf("NewRedisClient error %v", err)
	}
	return
}

// PerCommandTracer provides the instrumented WrapProcess function that you can attach to any
// client. It specifically takes in a context.Context as the first argument because
// you could be using the same client but wrapping it in a different context.
func PerCommandTracer(ctx context.Context) func(oldProcess func(cmd redis.Cmder) error) func(redis.Cmder) error {
	return func(fn func(cmd redis.Cmder) error) func(redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			_, span := trace.StartSpan(ctx, fmt.Sprintf("redis-go/%s", cmd.Name()))
			defer span.End()
			err := fn(cmd)
			if err != nil {
				span.SetStatus(trace.Status{Code: int32(trace.StatusCodeUnknown), Message: err.Error()})
			}
			return err
		}
	}
}
