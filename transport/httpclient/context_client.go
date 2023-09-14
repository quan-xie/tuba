package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	xhttp "net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"

	"go.opencensus.io/trace"
	"golang.org/x/net/http2"

	"github.com/pkg/errors"

	"github.com/quan-xie/tuba/backoff"
	"github.com/quan-xie/tuba/retry"
	"github.com/quan-xie/tuba/util/xtime"
)

const (
	minRead               = 16 * 1024 // 16kb
	defaultRetryCount int = 0
)

type Config struct {
	Dial            xtime.Duration
	Timeout         xtime.Duration
	KeepAlive       xtime.Duration
	MaxConns        int
	MaxIdle         int
	BackoffInterval xtime.Duration // Interval is second
	retryCount      int
}

type HttpClient struct {
	conf       *Config
	client     *xhttp.Client
	retryCount int
	retrier    retry.Retriable
}

// NewHTTPClient returns a new instance of httpClient
func NewHTTPClient(c *Config) *HttpClient {
	dialer := &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	transport := &xhttp.Transport{
		DialContext:         dialer.DialContext,
		MaxConnsPerHost:     c.MaxConns,
		MaxIdleConnsPerHost: c.MaxIdle,
		IdleConnTimeout:     time.Duration(c.KeepAlive),
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	_ = http2.ConfigureTransport(transport)
	bo := backoff.NewConstantBackoff(c.BackoffInterval)
	return &HttpClient{
		conf: c,
		client: &xhttp.Client{
			Transport: transport,
		},
		retryCount: defaultRetryCount,
		retrier:    retry.NewRetrier(bo),
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
func (c *HttpClient) Get(ctx context.Context, url string, headers xhttp.Header, res interface{}) (err error) {
	ctx, span := trace.StartSpan(ctx, "httpclient Get")
	defer span.End()
	request, err := xhttp.NewRequest(xhttp.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "GET - request creation failed")
	}

	request.Header = headers
	ats := []trace.Attribute{
		trace.StringAttribute("Get URL", url),
	}
	span.Annotate(ats, "GET Request")
	span.AddAttributes(ats...)
	return c.Do(ctx, request, res)
}

// Post makes a HTTP POST request to provided URL with context passed in
func (c *HttpClient) Post(ctx context.Context, url, contentType string, headers xhttp.Header, param, res interface{}) (err error) {
	ctx, span := trace.StartSpan(ctx, "httpclient Post")
	defer span.End()
	request, err := xhttp.NewRequest(xhttp.MethodPost, url, reqBody(contentType, param))
	if err != nil {
		return errors.Wrap(err, "POST - request creation failed")
	}
	if headers == nil {
		headers = make(xhttp.Header)
	}
	headers.Set("Content-Type", contentType)
	request.Header = headers
	paramByte, _ := sonic.Marshal(param)
	ats := []trace.Attribute{
		trace.StringAttribute("POST URL", url),
		trace.StringAttribute("POST PARAM ", string(paramByte)),
	}
	span.Annotate(ats, "POST Request")
	span.AddAttributes(ats...)
	return c.Do(ctx, request, res)
}

// Put makes a HTTP PUT request to provided URL with context passed in
func (c *HttpClient) Put(ctx context.Context, url, contentType string, headers xhttp.Header, param, res interface{}) (err error) {
	ctx, span := trace.StartSpan(ctx, "httpclient Put")
	defer span.End()
	request, err := xhttp.NewRequest(xhttp.MethodPut, url, reqBody(contentType, param))
	if err != nil {
		return errors.Wrap(err, "PUT - request creation failed")
	}

	if headers == nil {
		headers = make(xhttp.Header)
	}
	headers.Set("Content-Type", contentType)
	request.Header = headers
	paramByte, _ := sonic.Marshal(param)
	ats := []trace.Attribute{
		trace.StringAttribute("PUT URL", url),
		trace.StringAttribute("PUT PARAM ", string(paramByte)),
	}
	span.Annotate(ats, "PUT Request")
	span.AddAttributes(ats...)
	return c.Do(ctx, request, res)
}

// Patch makes a HTTP PATCH request to provided URL with context passed in
func (c *HttpClient) PATCH(ctx context.Context, url, contentType string, headers xhttp.Header, param, res interface{}) (err error) {
	ctx, span := trace.StartSpan(ctx, "httpclient Patch")
	defer span.End()
	request, err := xhttp.NewRequest(xhttp.MethodPatch, url, reqBody(contentType, param))
	if err != nil {
		return errors.Wrap(err, "PATCH - request creation failed")
	}

	if headers == nil {
		headers = make(xhttp.Header)
	}
	headers.Set("Content-Type", contentType)
	request.Header = headers
	paramByte, _ := sonic.Marshal(param)
	ats := []trace.Attribute{
		trace.StringAttribute("PATCH URL", url),
		trace.StringAttribute("PATCH PARAM ", string(paramByte)),
	}
	span.Annotate(ats, "PATCH Request")
	span.AddAttributes(ats...)
	return c.Do(ctx, request, res)
}

// Delete makes a HTTP DELETE request to provided URL with context passed in
func (c *HttpClient) Delete(ctx context.Context, url, contentType string, headers xhttp.Header, param, res interface{}) (err error) {
	ctx, span := trace.StartSpan(ctx, "httpclient Delete")
	defer span.End()
	request, err := xhttp.NewRequest(xhttp.MethodDelete, url, nil)
	if err != nil {
		return errors.Wrap(err, "DELETE - request creation failed")
	}

	if headers == nil {
		headers = make(xhttp.Header)
	}
	headers.Set("Content-Type", contentType)
	request.Header = headers
	paramByte, _ := sonic.Marshal(param)
	ats := []trace.Attribute{
		trace.StringAttribute("DELETE URL", url),
		trace.StringAttribute("DELETE PARAM ", string(paramByte)),
	}
	span.Annotate(ats, "DELETE Request")
	span.AddAttributes(ats...)
	return c.Do(ctx, request, res)
}

// Do makes an HTTP request with the native `http.Do` interface and context passed in
func (c *HttpClient) Do(ctx context.Context, req *xhttp.Request, res interface{}) (err error) {
	for i := 0; i <= c.retryCount; i++ {
		if err = c.request(ctx, req, res); err != nil {
			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			continue
		}
		break
	}
	return
}

func (c *HttpClient) request(ctx context.Context, req *xhttp.Request, res interface{}) (err error) {
	var (
		response *xhttp.Response
		bs       []byte
		cancel   func()
	)
	ctx, cancel = context.WithTimeout(ctx, time.Duration(c.conf.Timeout))
	defer cancel()
	response, err = c.client.Do(req.WithContext(ctx))
	if err != nil {
		<-ctx.Done()
		err = ctx.Err()
		return
	}
	defer response.Body.Close()
	if response.StatusCode >= xhttp.StatusInternalServerError {
		err = errors.Wrap(err, fmt.Sprintf("response.StatusCode %d", response.StatusCode))
		return
	}
	if bs, err = readAll(response.Body, minRead); err != nil {
		return
	}
	err = sonic.Unmarshal(bs, &res)
	return
}

func reqBody(contentType string, param interface{}) (body io.Reader) {
	if contentType == MIMEPOSTForm {
		enc, ok := param.(string)
		if ok {
			body = strings.NewReader(enc)
		}
	}
	if contentType == MIMEJSON {
		buff := new(bytes.Buffer)
		sonic.ConfigDefault.NewEncoder(buff).Encode(param)
		body = buff
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
