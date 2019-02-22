package main

import (
	"net/http"

	"github.com/taiyoh/toyhose"
	"github.com/taiyoh/toyhose/gateway"
)

func main() {
	repo := gateway.NewDeliveryStream()
	a := toyhose.NewAdapter(repo)
	http.ListenAndServe(":8080", a.ServeMux())
}
