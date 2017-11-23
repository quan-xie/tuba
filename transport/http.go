package transport

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type ClientConfig struct {
	Dial      time.Duration
	Timeout   time.Duration
	KeepAlive time.Duration
}

type Client struct {
	conf      *ClientConfig
	client    *http.Client
	dialer    *net.Dialer
	transport *http.Transport
}

func NewClient(c *ClientConfig) *Client {
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
