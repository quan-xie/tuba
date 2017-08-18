package tuba

import "time"

// HTTPServer http server settings.
type HTTPServer struct {
	Addrs        []string
	MaxListen    int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
