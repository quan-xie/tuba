package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/quan-xie/tuba/util/retry"
)

const defaultRetryCount int = 0

type Config struct {
	Dial       time.Duration
	Timeout    time.Duration
	KeepAlive  time.Duration
	retryCount int
}

type httpClientWithContext struct {
	client     *http.Client
	dialer     *net.Dialer
	transport  *http.Transport
	retryCount int
	retrier    retry.Retriable
}

// NewHTTPClientWithContext returns a new instance of httpClientWithContext
func NewHTTPClientWithContext(c *Config) ClientWithContext {
	dialer := &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	transport := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &httpClientWithContext{
		client: &http.Client{
			Transport: transport,
		},
		retryCount: defaultRetryCount,
		retrier:    retry.NewNoRetrier(),
	}
}

// SetRetryCount sets the retry count for the httpClient
func (c *httpClientWithContext) SetRetryCount(count int) {
	c.retryCount = count
}

// SetRetryCount sets the retry count for the httpClient
func (c *httpClientWithContext) SetRetrier(retrier retry.Retriable) {
	c.retrier = retrier
}

// Get makes a HTTP GET request to provided URL with context passed in
func (c *httpClientWithContext) Get(ctx context.Context, url string, headers http.Header) (*http.Response, error) {
	var response *http.Response
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return response, errors.Wrap(err, "GET - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request)
}

// Post makes a HTTP POST request to provided URL with context passed in
func (c *httpClientWithContext) Post(ctx context.Context, url string, body io.Reader, headers http.Header) (*http.Response, error) {
	var response *http.Response
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return response, errors.Wrap(err, "POST - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request)
}

// Put makes a HTTP PUT request to provided URL with context passed in
func (c *httpClientWithContext) Put(ctx context.Context, url string, body io.Reader, headers http.Header) (*http.Response, error) {
	var response *http.Response
	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return response, errors.Wrap(err, "PUT - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request)
}

// Patch makes a HTTP PATCH request to provided URL with context passed in
func (c *httpClientWithContext) Patch(ctx context.Context, url string, body io.Reader, headers http.Header) (*http.Response, error) {
	var response *http.Response
	request, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return response, errors.Wrap(err, "PATCH - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request)
}

// Delete makes a HTTP DELETE request to provided URL with context passed in
func (c *httpClientWithContext) Delete(ctx context.Context, url string, headers http.Header) (*http.Response, error) {
	var response *http.Response
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return response, errors.Wrap(err, "DELETE - request creation failed")
	}

	request.Header = headers

	return c.Do(ctx, request)
}

// Do makes an HTTP request with the native `http.Do` interface and context passed in
func (c *httpClientWithContext) Do(ctx context.Context, req *http.Request) (response *http.Response, err error) {

	for i := 0; i <= c.retryCount; i++ {
		contextCancelled := false
		var err error
		response, err = c.client.Do(req.WithContext(ctx))
		if err != nil {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				contextCancelled = true
			}

			if contextCancelled {
				break
			}
			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			continue
		}

		if response.StatusCode >= http.StatusInternalServerError {

			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			fmt.Println("R: ", response.StatusCode)
			continue
		}
		// Clear errors if any iteration succeeds
		break
	}

	return response, err
}
