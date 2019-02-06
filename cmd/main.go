package main

import (
	"net/http"

	"github.com/taiyoh/toyhose"
)

func main() {
	mux := http.NewServeMux()
	hdl := toyhose.NewHandler()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch req.URL.Path {
		case "/":
			hdl.ServeHTTP(res, req)
		default:
			http.NotFound(res, req)
		}
	})

	http.ListenAndServe(":8080", mux)
}
