package port

import (
	"encoding/json"
	"net/http"

	"github.com/taiyoh/toyhose/errors"
)

// Output provides response builder for each request
type Output struct {
	err      errors.Raised
	resource interface{}
}

// Set provides response body and error
func (o *Output) Set(r interface{}, err errors.Raised) {
	o.resource = r
	o.err = err
}

// Fill provides setup response
func (o *Output) Fill(res http.ResponseWriter) {
	if err := o.err; err != nil {
		code, body := err.Output()
		res.WriteHeader(code)
		res.Write(body)
		return
	}
	if resource := o.resource; resource != nil {
		out, _ := json.Marshal(resource)
		res.Write(out)
	}
}
