package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	xhttp "net/http"
	"strings"
	"sync"
	"time"
)

const (
	_minRead = 16 * 1024 // 16kb
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
	transport *xhttp.Transport
	client    *xhttp.Client
	mutex     sync.RWMutex
}

// New returns a new initialized Http Client.
func New(c *HttpConfig) *Client {
	client := new(Client)
	client.conf = c
	client.dialer = &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	client.transport = &xhttp.Transport{
		DialContext:         client.dialer.DialContext,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        c.MaxIdleConns,
		MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
		IdleConnTimeout:     c.IdleConnTimeout,
	}
	client.client = &xhttp.Client{
		Transport: client.transport,
	}
	return client
}

//Get send a get request, returns a usable response .
func (client *Client) Get(ctx context.Context, url string, header map[string]string, res interface{}) (err error) {
	req, err := newRequest(xhttp.MethodGet, url, header)
	if err != nil {
		log.Fatalf("newRequest error(%v)", err)
		return
	}
	return client.Do(ctx, req, res)
}

//Post send a post request, returns a usable response .
func (client *Client) Post(ctx context.Context, url string, header map[string]string, res interface{}) (err error) {
	req, err := newRequest(xhttp.MethodPost, url, header)
	if err != nil {
		log.Fatalf("newRequest error(%v)", err)
		return
	}
	return client.Do(ctx, req, res)
}

// Do send an http request and retun an http response .
func (client *Client) Do(ctx context.Context, req *xhttp.Request, res interface{}) (err error) {
	var (
		resp   *xhttp.Response
		bs     []byte
		cancel func()
	)
	ctx, cancel = context.WithTimeout(ctx, client.conf.Timeout)
	defer cancel()
	req = req.WithContext(ctx)
	if resp, err = client.client.Do(req); err != nil {
		log.Fatalf("http client.Do error (%v)", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= xhttp.StatusBadRequest {
		log.Fatalf("response status code error (%v)", err)
		return
	}
	if bs, err = readAll(resp.Body, _minRead); err != nil {
		log.Fatalf("readAll error (%v)", err)
		return
	}
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
			log.Fatalf("json.Unmarshal error (%v)", err)
		}
	}
	return
}

// newRequest new http request with method, url, and header.
func newRequest(method, url string, header map[string]string) (req *xhttp.Request, err error) {
	if method == xhttp.MethodGet {
		req, err = xhttp.NewRequest(xhttp.MethodGet, url, nil)
	} else {
		req, err = xhttp.NewRequest(xhttp.MethodPost, url, strings.NewReader(url))
	}
	if err != nil {
		log.Fatalf("xhttp.NewRequest error (%v)")
		return
	}
	for k, v := range header {
		if k != "" && v != "" {
			req.Header.Set(k, v)
		}
	}
	return
}

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
