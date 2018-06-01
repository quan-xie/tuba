package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// ClientConfig is http client config
type Config struct {
	Dial      time.Duration
	Timeout   time.Duration
	KeepAlive time.Duration
}

// Client is http client
type Client struct {
	conf      *Config
	client    *http.Client
	dialer    *net.Dialer
	transport *http.Transport
}

// NewClient returns a newly initialized Http Client object that implements the Client
func NewClient(c *Config) *Client {
	client := new(Client)
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
