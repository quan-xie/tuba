package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v3"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/mercari/go-circuitbreaker"
	"github.com/quan-xie/tuba/log"
	"github.com/quan-xie/tuba/util/xtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ClientConfig struct {
	Addr           string
	LoadBalancing  string
	Timeout        xtime.Duration
	RequestTimeout xtime.Duration
	Circuitbreaker Circuitbreaker
}

type Circuitbreaker struct {
	CounterResetInterval xtime.Duration // 断路器间隔时间
	Threshold            int64          // 计数器阈值
	OpenTimeout          xtime.Duration // 超时时间
	HalfOpenMaxSuccesses int64          // HalfOpen 次数
}

func NewRPCClient(cfg *ClientConfig) *grpc.ClientConn {
	var err error
	if cfg.LoadBalancing == "" {
		cfg.LoadBalancing = fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = xtime.Duration(time.Second)
	}

	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = xtime.Duration(time.Second)
	}
	if cfg.Circuitbreaker.CounterResetInterval == 0 {
		cfg.Circuitbreaker.CounterResetInterval = xtime.Duration(time.Minute)
	}
	if cfg.Circuitbreaker.Threshold == 0 {
		cfg.Circuitbreaker.Threshold = 3
	}
	if cfg.Circuitbreaker.OpenTimeout == 0 {
		cfg.Circuitbreaker.OpenTimeout = xtime.Duration(20 * time.Second)
	}
	if cfg.Circuitbreaker.HalfOpenMaxSuccesses == 0 {
		cfg.Circuitbreaker.HalfOpenMaxSuccesses = 10
	}
	cb := circuitbreaker.New(
		circuitbreaker.WithFailOnContextCancel(true),
		circuitbreaker.WithFailOnContextDeadline(true),
		circuitbreaker.WithCounterResetInterval(time.Duration(cfg.Circuitbreaker.CounterResetInterval)),
		circuitbreaker.WithTripFunc(circuitbreaker.NewTripFuncThreshold(cfg.Circuitbreaker.Threshold)),
		circuitbreaker.WithOpenTimeout(time.Duration(cfg.Circuitbreaker.OpenTimeout)),
		circuitbreaker.WithOpenTimeoutBackOff(backoff.NewExponentialBackOff()),
		circuitbreaker.WithHalfOpenMaxSuccesses(cfg.Circuitbreaker.HalfOpenMaxSuccesses),
		circuitbreaker.WithOnStateChangeHookFn(func(from, to circuitbreaker.State) {
			log.Infof("state changed from %s to %s\n", from, to)
		}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout))
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Addr, grpc.WithTransportCredentials(
		insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(cfg.LoadBalancing),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				UnaryClientInterceptor(
					cb,
					func(ctx context.Context, method string, req interface{}) {
						log.Info("[Client] Circuit breaker is open.\n")
					},
				),
				timeout.TimeoutUnaryClientInterceptor(time.Duration(cfg.RequestTimeout)),
			),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             500 * time.Millisecond,
			PermitWithoutStream: true}),
	)

	if err != nil {
		log.Errorf("grpc client addr:%s, err:%v", cfg.Addr, err)
	}

	return conn
}

type OpenStateHandler func(ctx context.Context, method string, req interface{})

func UnaryClientInterceptor(cb *circuitbreaker.CircuitBreaker, handler OpenStateHandler) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := cb.Do(ctx, func() (interface{}, error) {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err != nil {
				return nil, err
			}

			return nil, nil
		})

		if err == circuitbreaker.ErrOpen {
			handler(ctx, method, req)
		}

		return err
	}
}
