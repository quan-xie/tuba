package main

import (
	"net/http"

	"github.com/quan-xie/tuba"
)

func main() {
	r := tuba.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(":3000", r)
}
