package redis

import (
	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/log"
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
