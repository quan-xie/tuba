package transport

import (
	"crypto/tls"
	"net"
	"net/http"
)

type Client struct {
	client    *http.Client
	dialer    *net.Dialer
	transport *http.Transport
}

func NewClient() *Client {
	client := new(Client)
	client.transport = &http.Transport{
		DialContext:     client.dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.client = &http.Client{
		Transport: client.transport,
	}
	return client
}
