package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/util/retry"
	"github.com/quan-xie/tuba/util/xtime"
)

const (
	minRead               = 16 * 1024 // 16kb
	defaultRetryCount int = 0
)

type Config struct {
	Dial       xtime.Duration
	Timeout    xtime.Duration
	KeepAlive  xtime.Duration
	retryCount int
}

type HttpClient struct {
	conf       *Config
	client     *http.Client
	dialer     *net.Dialer
	transport  *http.Transport
	retryCount int
	retrier    retry.Retriable
}

// NewHTTPClient returns a new instance of httpClient
func NewHTTPClient(c *Config) *HttpClient {
	dialer := &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	transport := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &HttpClient{
		conf: c,
		client: &http.Client{
			Transport: transport,
		},
		retryCount: defaultRetryCount,
		retrier:    retry.NewNoRetrier(),
	}
}

// SetRetryCount sets the retry count for the httpClient
func (c *HttpClient) SetRetryCount(count int) {
	c.retryCount = count
}

// SetRetryCount sets the retry count for the httpClient
func (c *HttpClient) SetRetrier(retrier retry.Retriable) {
	c.retrier = retrier
}

// Get makes a HTTP GET request to provided URL with context passed in
func (c *HttpClient) Get(ctx context.Context, url string, headers http.Header, res interface{}) (err error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "GET - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request, res)
}

// Post makes a HTTP POST request to provided URL with context passed in
func (c *HttpClient) Post(ctx context.Context, url string, body io.Reader, headers http.Header, res interface{}) (err error) {
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return errors.Wrap(err, "POST - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request, res)
}

// Put makes a HTTP PUT request to provided URL with context passed in
func (c *HttpClient) Put(ctx context.Context, url string, body io.Reader, headers http.Header, res interface{}) (err error) {
	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return errors.Wrap(err, "PUT - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request, res)
}

// Patch makes a HTTP PATCH request to provided URL with context passed in
func (c *HttpClient) Patch(ctx context.Context, url string, body io.Reader, headers http.Header, res interface{}) (err error) {
	request, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return errors.Wrap(err, "PATCH - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request, res)
}

// Delete makes a HTTP DELETE request to provided URL with context passed in
func (c *HttpClient) Delete(ctx context.Context, url string, headers http.Header, res interface{}) (err error) {
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return errors.Wrap(err, "DELETE - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request, res)
}

// Do makes an HTTP request with the native `http.Do` interface and context passed in
func (c *HttpClient) Do(ctx context.Context, req *http.Request, res interface{}) (err error) {
	for i := 0; i <= c.retryCount; i++ {
		if err = c.request(ctx, req, res); err != nil {
			err = errors.Wrap(err, "request - request failed")
			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			continue
		}
		break
	}
	return
}

func (c *HttpClient) request(ctx context.Context, req *http.Request, res interface{}) (err error) {
	var (
		response *http.Response
		bs       []byte
		cancel   func()
	)
	ctx, cancel = context.WithTimeout(ctx, time.Duration(c.conf.Timeout))
	defer cancel()
	response, err = c.client.Do(req.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		}
		return
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusInternalServerError {
		err = errors.Wrap(err, "StatusInternalServerError - Status Internal ServerError")
		return
	}
	if bs, err = readAll(response.Body, minRead); err != nil {
		err = errors.Wrap(err, "readAll - readAll failed")
		return
	}
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
			err = errors.Wrap(err, "Unmarshal failed")
		}
	}
	return
}

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
