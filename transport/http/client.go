package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// HttpConfig is http config ,include Dial Timeout and KeepAlive.
type HttpConfig struct {
	Dial      time.Duration
	Timeout   time.Duration
	KeepAlive time.Duration
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
	client := &Client{}
	client.conf = c
	client.dialer = &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	client.transport = &http.Transport{
		DialContext:     client.dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.client = &http.Client{
		Transport: client.transport,
	}
	return client
}
