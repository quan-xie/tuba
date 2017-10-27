package transport

import (
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
)

func Serve(mux *http.ServeMux) (err error) {
	//for _, addr := range c.Addrs {
	l, err := net.Listen("tcp", "8080")
	if err != nil {
		log.Fatalf("net.Listen(\"tcp\", \"%s\") error(%v)", "8080", err)
		return err
	}
	//if c.MaxListen > 0 {
	//	l = netutil.LimitListener(l, c.MaxListen)
	//}
	//log.Info("start http listen addr: %s", addr)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			server := &http.Server{Handler: mux, ReadTimeout: time.Duration(1000), WriteTimeout: time.Duration(1000)}
			if err := server.Serve(l); err != nil {
				//log.Info("server.Serve error(%v)", err)
			}
		}()
	}
	return nil
}
