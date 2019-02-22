package toyhose

import (
	"net/http"

	"github.com/taiyoh/toyhose/actions"
)

type Adapter struct {
	mux  *http.ServeMux
	repo actions.DeliveryStreamRepository
}

func NewAdapter(repo actions.DeliveryStreamRepository) *Adapter {
	mux := http.NewServeMux()
	a := &Adapter{
		mux:  mux,
		repo: repo,
	}
	mux.HandleFunc("/", a.handleFn)
	return a
}

func (a *Adapter) handleFn(res http.ResponseWriter, req *http.Request) {
	if !a.validateRequest(res, req) {
		return
	}
	d := NewDispatcher(a.repo, req.Body)
	fn := d.Dispatch(req.Header.Get("X-Amz-Target"))
	if fn == nil {
		http.NotFound(res, req)
		return
	}
	fn(d.Arg())
}

func (a *Adapter) validateRequest(res http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return false
	}
	return true
}

func (a *Adapter) ServeMux() *http.ServeMux {
	return a.mux
}
