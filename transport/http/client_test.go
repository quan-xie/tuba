package http

import (
	"context"
	"testing"
	"time"
)

var (
	httpConfig *HttpConfig
)

func init() {
	httpConfig = &HttpConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		Dial:                30 * time.Second,
		Timeout:             30 * time.Second,
		KeepAlive:           30 * time.Second,
		IdleConnTimeout:     30 * time.Second,
	}
}

// go test  -v -test.run TestGet
func TestGet(t *testing.T) {
	var (
		err error
	)
	client := New(httpConfig)
	var res interface{}
	header := make(map[string]string)
	header["User-Agent"] = "next test"
	if err = client.Get(context.TODO(), "https://api.github.com", header, &res); err != nil {
		t.Fatalf("client.Get error(%v)", err)
		return
	}
	if res != nil {
		t.Logf("response data: %v \n", res)
	}
}

// go test -v -test.run TestPost
func TestPost(t *testing.T) {
	var (
		err error
	)
	client := New(httpConfig)
	var res interface{}
	header := make(map[string]string)
	header["User-Agent"] = "next test"
	if err = client.Post(context.TODO(), "https://api.github.com", header, &res); err != nil {
		t.Fatalf("client.Get error(%v)", err)
		return
	}
	if res != nil {
		t.Logf("response data: %v \n", res)
	}
}
