package toyhose

import (
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// ctx := context.Background()
	switch FindCommand(req.Header.Get("X-Amz-Target")) {
	case CommandCreateDeliveryStream:
	case CommandDeleteDeliveryStream:
	case CommandDescribeDeliveryStream:
	case CommandListDeliveryStreams:
	case CommandPutRecord:
	case CommandPutRecordBatch:
	default:
		http.NotFound(res, req)
	}
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("hello!"))
}
