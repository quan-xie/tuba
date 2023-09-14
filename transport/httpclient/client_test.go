package httpclient

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/quan-xie/tuba/util/xtime"
)

var (
	cfg    *Config
	client *HttpClient
)

func init() {
	cfg = &Config{
		Dial:            xtime.Duration(time.Second),
		Timeout:         xtime.Duration(time.Second),
		KeepAlive:       xtime.Duration(time.Second),
		BackoffInterval: xtime.Duration(time.Second),
		retryCount:      10,
	}
	client = NewHTTPClient(cfg)
}

// go test -v -test.run TestHttpClient_Get
func TestHttpClient_Get(t *testing.T) {
	var res interface{}
	client.SetRetryCount(5)
	err := client.Get(context.Background(), "https://http2.pro/api/v1", nil, &res)
	if err != nil {
		t.Log(err)
		return
	}
	resStr, _ := sonic.MarshalString(&res)
	t.Log(resStr)
}

// go test -v -test.run TestHttpClient_Post
func TestHttpClient_Post(t *testing.T) {
	var res interface{}
	param := make(map[string]interface{})
	param["xie"] = "quan"
	err := client.Post(context.Background(), "https://http2.pro/api/v1", MIMEJSON, nil, param, &res)
	if err != nil {
		t.Log(err)
		return
	}
	resStr, _ := sonic.MarshalString(&res)
	t.Log(resStr)
}
