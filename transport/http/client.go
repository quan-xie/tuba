package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// HttpConfig is http config ,include Dial Timeout and KeepAlive.
type HttpConfig struct {
	Dial                time.Duration
	Timeout             time.Duration
	KeepAlive           time.Duration
	IdleConnTimeout     time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
}

// Client is http Client .
type Client struct {
	conf      *HttpConfig
	dialer    *net.Dialer
	transport *http.Transport
	client    *http.Client
}

// DisableKeepAlives, if true, prevents re-use of TCP connections
// between different HTTP requests.
// MaxIdleConns controls the maximum number of idle (keep-alive)
// connections across all hosts. Zero means no limit.
// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
// (keep-alive) connections to keep per-host. If zero,
// DefaultMaxIdleConnsPerHost is used.
// New returns a new initialized Http Client.
func New(c *HttpConfig) *Client {
	client := &Client{
		conf: c,
		transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(c.Dial),
				KeepAlive: time.Duration(c.KeepAlive),
			}).DialContext,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        c.MaxIdleConns,
			MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
			IdleConnTimeout:     c.IdleConnTimeout,
		},
	}
	return client
}
