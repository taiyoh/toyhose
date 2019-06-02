package main

import (
	"flag"
	"net/http"

	"github.com/taiyoh/toyhose"
	"github.com/taiyoh/toyhose/gateway"
)

func main() {
	fs := flag.NewFlagSet("toyhose", flag.ExitOnError)
	var region, accountID string
	fs.StringVar(&region, "region", "toyhose", "specified aws region (default: toyhose)")
	fs.StringVar(&accountID, "accountId", "0000000001", "accountId for ARN (default: 0000000001)")

	dsRepo := gateway.NewDeliveryStream()
	destRepo := gateway.NewDestination()

	th := toyhose.NewServe(region, accountID, dsRepo, destRepo)
	http.ListenAndServe(":8080", th.ServeMux())
}
