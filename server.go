package next

import (
	"golang.org/x/net/netutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
)

// Serve listen and serve http handlers with limitListener.
// LimitListener returns a Listener that accepts at most n simultaneous
func Serve(m *http.ServeMux, h *HTTPServer) (err error) {
	for _, addr := range h.Addrs {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("net.Listen(\"tcp\", \"%s\") error(%v)", addr, err)
		}
		if h.MaxListen > 0 {
			l = netutil.LimitListener(l, h.MaxListen)
		}
		log.Printf("start http listen addr: %s", addr)
		for i := 0; i < runtime.NumCPU(); i++ {
			go func() {
				s := &http.Server{Handler: m, ReadTimeout: time.Duration(h.ReadTimeout), WriteTimeout: time.Duration(h.WriteTimeout)}
				if err := s.Serve(l); err != nil {
					log.Fatalf("Server Serve error(%v)", err)
				}
			}()
		}
	}
	return
}
