package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/quan-xie/tuba/log"
	"github.com/quan-xie/tuba/util/xtime"
)

type Config struct {
	Addr         string
	Handler      http.Handler
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

var httpServer *http.Server

func Run(c *Config) {
	httpServer = &http.Server{
		Addr:         c.Addr,
		Handler:      c.Handler,
		ReadTimeout:  time.Duration(c.ReadTimeout),
		WriteTimeout: time.Duration(c.WriteTimeout),
	}

	go func() {
		log.Infof("http server succeed listening at %v", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

func Stop(ctx context.Context) {
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Errorf("http server stop error %v", err)
	}
}
