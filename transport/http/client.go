package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// DisableKeepAlives, if true, prevents re-use of TCP connections
// between different HTTP requests.
// MaxIdleConns controls the maximum number of idle (keep-alive)
// connections across all hosts. Zero means no limit.
// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
// (keep-alive) connections to keep per-host. If zero,
// DefaultMaxIdleConnsPerHost is used.
// HttpConfig is http config ,include Dial Timeout and KeepAlive.
type HttpConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	Dial                time.Duration
	Timeout             time.Duration
	KeepAlive           time.Duration
	IdleConnTimeout     time.Duration
}

// Client is http Client .
type Client struct {
	conf      *HttpConfig
	dialer    *net.Dialer
	transport *http.Transport
	client    *http.Client
}

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

