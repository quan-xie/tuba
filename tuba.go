package tuba

import (
	"net/http"
)

type Tuba struct {
	router *Router
	mux    *http.ServeMux
	hs     *HTTPServer
}

func New(m *http.ServeMux, h *HTTPServer) *Tuba {
	return &Tuba{
		router: newRouter(m),
		mux:    m,
		hs:     h,
	}
}

func (t *Tuba) Run() {
	Serve(t.mux, t.hs)
}
