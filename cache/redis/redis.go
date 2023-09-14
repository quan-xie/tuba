package redis

import (
	"github.com/quan-xie/tuba/log"

	"github.com/pkg/errors"
)

func NewRedis(config *Config) (*RedisStorage, error) {
	if config == nil {
		return nil, errors.New("init fail: config is nil")
	}

	return CreateRedisStorage(config)
}

func NewRedisClient(config *Config) (client *RedisStorage) {
	var err error
	client, err = CreateRedisStorage(config)
	if err != nil {
		log.Fatalf("NewRedisClient error %v", err)
	}
	return
}
