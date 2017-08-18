package tuba

import "net/http"

type Handle func(w http.ResponseWriter, r *http.Request)

type Router interface {
}

type router struct {
	mux *http.ServeMux
}

func newRouter(m *http.ServeMux) Router {
	return &router{
		mux: m,
	}
}
