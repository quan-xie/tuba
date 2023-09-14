package redis

import (
	"context"
	"testing"
	"time"

	"github.com/quan-xie/tuba/util/xtime"
)

func TestRedisInit(t *testing.T) {
	configs := &Config{
		Name:         "fortest",
		Addrs:        []string{"127.0.0.1:6379"},
		Password:     "",
		DB:           1,
		Dial:         xtime.Duration(5 * time.Second),
		KeepAlive:    xtime.Duration(5 * time.Minute),
		MinIdleConns: 1,
	}

	redisStorage, err := NewRedis(configs)
	if err != nil {
		t.Fatal(err)
	}

	r, e := redisStorage.redis.Set(context.Background(), "test", "1", time.Second*100).Result()
	if e != nil {
		t.Fatal(e)
	}
	t.Log(r)

	r2, e2 := redisStorage.HSet(context.Background(), "test_table", "field1", "value1").Result()
	if e2 != nil {
		t.Fatal(e2)
	}
	t.Log(r2)

	r3, e3 := redisStorage.HGet(context.Background(), "test_table", "field1").Result()
	if e3 != nil {
		t.Fatal(e3)
	}
	t.Log(r3)
}
