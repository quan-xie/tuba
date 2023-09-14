package redis

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/util/xtime"
	xredis "github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	redis         *xredis.Client
	clusterClient *xredis.ClusterClient
	config        *Config
}

type Config struct {
	Name         string         `json:"name"`
	Addrs        []string       `json:"addrs"`
	Password     string         `json:"password"`
	PoolSize     int            `json:"pool_size"`
	DB           int            `json:"db"`
	MinIdleConns int            `json:"minidle_conns"`
	Dial         xtime.Duration `json:"dial"`
	KeepAlive    xtime.Duration `json:"keep_alive"`
	Mode         string         `json:"mode"` // mode为"cluster"时按照集群模式初始化
}

func CreateRedisStorage(option *Config) (*RedisStorage, error) {
	if len(option.Addrs) == 0 || option == nil {
		return nil, errors.New("addrs cannot be empty")
	}
	if len(option.Addrs) > 1 || option.Mode == "cluster" {
		o := &xredis.ClusterOptions{
			Addrs:        option.Addrs,
			ReadOnly:     true,
			MinIdleConns: option.MinIdleConns,
		}
		o.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   time.Duration(option.Dial),
				KeepAlive: time.Duration(option.KeepAlive),
			}
			if o.TLSConfig == nil {
				return netDialer.DialContext(ctx, network, addr)
			}
			return tls.DialWithDialer(netDialer, network, addr, o.TLSConfig)
		}

		if option.PoolSize > 0 {
			o.PoolSize = option.PoolSize
		}

		client := xredis.NewClusterClient(o)
		return &RedisStorage{
			clusterClient: client,
			config:        option,
		}, nil
	} else {
		o := &xredis.Options{
			Addr:         option.Addrs[0],
			Password:     option.Password,
			DB:           option.DB,
			MinIdleConns: option.MinIdleConns,
		}

		o.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   time.Duration(option.Dial),
				KeepAlive: time.Duration(option.KeepAlive),
			}
			if o.TLSConfig == nil {
				return netDialer.DialContext(ctx, network, addr)
			}
			return tls.DialWithDialer(netDialer, network, addr, o.TLSConfig)
		}

		if option.PoolSize > 0 {
			o.PoolSize = option.PoolSize
		}
		client := xredis.NewClient(o)
		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			return nil, err
		}
		return &RedisStorage{
			redis:  client,
			config: option,
		}, nil
	}
}

func (rs *RedisStorage) DB() *xredis.Client {
	return rs.redis
}

func (rs *RedisStorage) ClusterDB() *xredis.ClusterClient {
	return rs.clusterClient
}

func (rs *RedisStorage) ZRevRange(ctx context.Context, key string, start, stop int64) *xredis.StringSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRevRange(ctx, key, start, stop)
	}
	return rs.redis.ZRevRange(ctx, key, start, stop)
}

func (rs *RedisStorage) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *xredis.ZSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRevRangeWithScores(ctx, key, start, stop)
	}
	return rs.redis.ZRevRangeWithScores(ctx, key, start, stop)
}

func (rs *RedisStorage) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *xredis.ZRangeBy) *xredis.ZSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRevRangeByScoreWithScores(ctx, key, opt)
	}
	return rs.redis.ZRevRangeByScoreWithScores(ctx, key, opt)
}

func (rs *RedisStorage) ZRangeByScoreWithScores(ctx context.Context, key string, opt *xredis.ZRangeBy) *xredis.ZSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRangeByScoreWithScores(ctx, key, opt)
	}
	return rs.redis.ZRangeByScoreWithScores(ctx, key, opt)
}

func (rs *RedisStorage) ZScore(ctx context.Context, key, member string) *xredis.FloatCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZScore(ctx, key, member)
	}
	return rs.redis.ZScore(ctx, key, member)
}

func (rs *RedisStorage) SMembers(ctx context.Context, key string) *xredis.StringSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SMembers(ctx, key)
	}
	return rs.redis.SMembers(ctx, key)
}

func (rs *RedisStorage) SAdd(ctx context.Context, key string, members ...interface{}) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SAdd(ctx, key, members...)
	}
	return rs.redis.SAdd(ctx, key, members...)
}

func (rs *RedisStorage) SRem(ctx context.Context, key string, members ...interface{}) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SRem(ctx, key, members...)
	}
	return rs.redis.SRem(ctx, key, members...)
}

func (rs *RedisStorage) SIsMember(ctx context.Context, key string, member interface{}) *xredis.BoolCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SIsMember(ctx, key, member)
	}
	return rs.redis.SIsMember(ctx, key, member)
}

func (rs *RedisStorage) Get(ctx context.Context, key string) *xredis.StringCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.Get(ctx, key)
	}
	return rs.redis.Get(ctx, key)
}

func (rs *RedisStorage) MGet(ctx context.Context, keys []string) *xredis.SliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.MGet(ctx, keys...)
	}
	return rs.redis.MGet(ctx, keys...)
}

func (rs *RedisStorage) IncrBy(ctx context.Context, key string, value int64) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.IncrBy(ctx, key, value)
	}
	return rs.redis.IncrBy(ctx, key, value)
}

func (rs *RedisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *xredis.StatusCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.Set(ctx, key, value, expiration)
	}
	return rs.redis.Set(ctx, key, value, expiration)
}

func (rs *RedisStorage) SCard(ctx context.Context, key string) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SCard(ctx, key)
	}
	return rs.redis.SCard(ctx, key)
}

func (rs *RedisStorage) Expire(ctx context.Context, key string, expiration time.Duration) *xredis.BoolCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.Expire(ctx, key, expiration)
	}
	return rs.redis.Expire(ctx, key, expiration)
}

func (rs *RedisStorage) ZAdd(ctx context.Context, key string, members ...xredis.Z) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZAdd(ctx, key, members...)
	}
	return rs.redis.ZAdd(ctx, key, members...)
}

func (rs *RedisStorage) ZRem(ctx context.Context, key string, members ...interface{}) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRem(ctx, key, members...)
	}
	return rs.redis.ZRem(ctx, key, members...)
}

func (rs *RedisStorage) ZRangeByScore(ctx context.Context, key string, opt *xredis.ZRangeBy) *xredis.StringSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRangeByScore(ctx, key, opt)
	}
	return rs.redis.ZRangeByScore(ctx, key, opt)
}

func (rs *RedisStorage) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *xredis.ZSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZRangeWithScores(ctx, key, start, stop)
	}
	return rs.redis.ZRangeWithScores(ctx, key, start, stop)
}

func (rs *RedisStorage) ZCard(ctx context.Context, key string) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZCard(ctx, key)
	}
	return rs.redis.ZCard(ctx, key)
}

func (rs *RedisStorage) ZPopMax(ctx context.Context, key string, count ...int64) *xredis.ZSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZPopMax(ctx, key, count...)
	}
	return rs.redis.ZPopMax(ctx, key, count...)
}

func (rs *RedisStorage) ZPopMin(ctx context.Context, key string, count ...int64) *xredis.ZSliceCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZPopMin(ctx, key, count...)
	}
	return rs.redis.ZPopMin(ctx, key, count...)
}

func (rs *RedisStorage) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) *xredis.ZWithKeyCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.BZPopMax(ctx, timeout, keys...)
	}
	return rs.redis.BZPopMax(ctx, timeout, keys...)
}

func (rs *RedisStorage) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) *xredis.ZWithKeyCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.BZPopMin(ctx, timeout, keys...)
	}
	return rs.redis.BZPopMin(ctx, timeout, keys...)
}

func (rs *RedisStorage) Del(ctx context.Context, key string) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.Del(ctx, key)
	}
	return rs.redis.Del(ctx, key)
}

func (rs *RedisStorage) SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) *xredis.BoolCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SetNX(ctx, key, value, expiration)
	}
	return rs.redis.SetNX(ctx, key, value, expiration)
}

func (rs *RedisStorage) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *xredis.StatusCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.SetEx(ctx, key, value, expiration)
	}
	return rs.redis.SetEx(ctx, key, value, expiration)
}

func (rs *RedisStorage) HSet(ctx context.Context, key string, field interface{}, value interface{}) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.HSet(ctx, key, field, value)
	}
	return rs.redis.HSet(ctx, key, field, value)
}

func (rs *RedisStorage) HGet(ctx context.Context, key string, field string) *xredis.StringCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.HGet(ctx, key, field)
	}
	return rs.redis.HGet(ctx, key, field)
}

func (rs *RedisStorage) HGetAll(ctx context.Context, key string) *xredis.MapStringStringCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.HGetAll(ctx, key)
	}
	return rs.redis.HGetAll(ctx, key)
}

func (rs *RedisStorage) HDel(ctx context.Context, key string, field string) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.HDel(ctx, key, field)
	}
	return rs.redis.HDel(ctx, key, field)
}

func (rs *RedisStorage) HIncrBy(ctx context.Context, key string, field string, incr int64) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.HIncrBy(ctx, key, field, incr)
	}
	return rs.redis.HIncrBy(ctx, key, field, incr)
}

func (rs *RedisStorage) HIncrByFloat(ctx context.Context, key string, field string, incr float64) *xredis.FloatCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.HIncrByFloat(ctx, key, field, incr)
	}
	return rs.redis.HIncrByFloat(ctx, key, field, incr)
}

func (rs *RedisStorage) LPush(ctx context.Context, key string, values ...interface{}) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.LPush(ctx, key, values)
	}
	return rs.redis.LPush(ctx, key, values)
}

func (rs *RedisStorage) RPop(ctx context.Context, key string) *xredis.StringCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.RPop(ctx, key)
	}
	return rs.redis.RPop(ctx, key)
}

func (rs *RedisStorage) RPush(ctx context.Context, key string, values ...interface{}) *xredis.IntCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.RPush(ctx, key, values)
	}
	return rs.redis.RPush(ctx, key, values)
}

func (rs *RedisStorage) LPop(ctx context.Context, key string) *xredis.StringCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.LPop(ctx, key)
	}
	return rs.redis.LPop(ctx, key)
}

func (rs *RedisStorage) Publish(ctx context.Context, channel string, msg interface{}) (err error) {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.Publish(ctx, channel, msg).Err()
	}
	return rs.redis.Publish(ctx, channel, msg).Err()
}

func (rs *RedisStorage) Subscribe(ctx context.Context, channels []string) *xredis.PubSub {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.Subscribe(ctx, channels...)
	}
	return rs.redis.Subscribe(ctx, channels...)
}

func (rs *RedisStorage) ZIncrBy(ctx context.Context, key string, increment float64, member string) *xredis.FloatCmd {
	if rs.config.Mode == "cluster" {
		return rs.clusterClient.ZIncrBy(ctx, key, increment, member)
	}
	return rs.redis.ZIncrBy(ctx, key, increment, member)
}