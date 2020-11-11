package grpc

import (
	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/util/xtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
	"time"
)

type ServerConfig struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
	LogFlag           int8
}

type Server struct {
	conf  *ServerConfig
	mutex sync.RWMutex

	server      *grpc.Server
	interceptor []grpc.UnaryServerInterceptor
}

func NewServer(c *ServerConfig, opts ...grpc.ServerOption) (s *Server, err error) {
	s = &Server{}
	if err := s.configuration(c); err != nil {
		panic("grpc config error")
	}
	kp := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     0,
		MaxConnectionAge:      0,
		MaxConnectionAgeGrace: 0,
		Time:                  0,
		Timeout:               0,
	})
	opts = append(opts, kp)
	s.server = grpc.NewServer(opts...)
	return
}

func (s *Server) Start() (err error) {
	listener, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	reflection.Register(s.server)
	return s.Serve(listener)
}

func (s *Server) Use(interceptors ...grpc.UnaryServerInterceptor) *Server {
	s.interceptor = append(s.interceptor, interceptors...)
	return s
}

func (s *Server) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func (s *Server) Server() *grpc.Server {
	return s.server
}

func (s *Server) configuration(c *ServerConfig) (err error) {
	if c.Addr == "" {
		c.Addr = "0.0.0.0:9000"
	}
	if c.Network == "" {
		c.Network = "tcp"
	}
	if c.Timeout <= 0 {
		c.Timeout = xtime.Duration(time.Second)
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = xtime.Duration(time.Second * 60)
	}
	if c.MaxLifeTime <= 0 {
		c.MaxLifeTime = xtime.Duration(time.Hour * 2)
	}
	if c.ForceCloseWait <= 0 {
		c.ForceCloseWait = xtime.Duration(time.Second * 20)
	}
	if c.KeepAliveInterval <= 0 {
		c.KeepAliveInterval = xtime.Duration(time.Second * 60)
	}
	if c.KeepAliveTimeout <= 0 {
		c.KeepAliveTimeout = xtime.Duration(time.Second * 20)
	}
	s.mutex.Lock()
	s.conf = c
	s.mutex.Unlock()
	return
}
