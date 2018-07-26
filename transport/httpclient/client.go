package httpclient

import (
	"context"
	"io"
	"net/http"

	"github.com/quan-xie/tuba/util/retry"
)

type ClientWithContext interface {
	Get(ctx context.Context, url string, headers http.Header) (*http.Response, error)
	Post(ctx context.Context, url string, body io.Reader, headers http.Header) (*http.Response, error)
	Put(ctx context.Context, url string, body io.Reader, headers http.Header) (*http.Response, error)
	Patch(ctx context.Context, url string, body io.Reader, headers http.Header) (*http.Response, error)
	Delete(ctx context.Context, url string, headers http.Header) (*http.Response, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)

	SetRetryCount(count int)
	SetRetrier(retrier retry.Retriable)
}
