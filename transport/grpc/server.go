package grpc

import (
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/log"
	"github.com/quan-xie/tuba/util/xtime"
	xgprc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
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

	server       *xgprc.Server
	healthServer *health.Server
	interceptor  []xgprc.UnaryServerInterceptor
}

func NewServer(c *ServerConfig, opts ...xgprc.ServerOption) (s *Server, err error) {
	s = &Server{}
	if err := s.configuration(c); err != nil {
		panic("grpc config error")
	}
	kp := xgprc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
	})
	opts = append(opts, kp)
	kaep := xgprc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	})
	opts = append(opts, kaep)
	s.server = xgprc.NewServer(opts...)
	s.healthServer = health.NewServer()
	healthpb.RegisterHealthServer(s.server, s.healthServer)
	return
}

func (s *Server) Start() {
	var err error
	l, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf("failed to net Listen: %v", err)
		return
	}
	reflection.Register(s.server)
	go func() {
		log.Infof("grpc server succeed listening at %v", l.Addr())
		if err := s.server.Serve(l); err != nil {
			err = errors.WithStack(err)
			log.Errorf("failed to serve: %v", err)
		}
	}()
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) Use(interceptors ...xgprc.UnaryServerInterceptor) *Server {
	s.interceptor = append(s.interceptor, interceptors...)
	return s
}

func (s *Server) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func (s *Server) Server() *xgprc.Server {
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
