package port

import (
	"encoding/json"
	"net/http"

	"github.com/taiyoh/toyhose/errors"
)

type Output struct {
	err      errors.Raised
	resource interface{}
}

func (o *Output) Set(r interface{}, err errors.Raised) {
	o.resource = r
	o.err = err
}

func (o *Output) Fill(res http.ResponseWriter) {
	if err := o.err; err != nil {
		res.WriteHeader(err.Code())
	}
	if resource := o.resource; resource != nil {
		out, _ := json.Marshal(resource)
		res.Write(out)
	}
}

func (o *Output) Error() error {
	return o.err
}

func (o *Output) Resource() interface{} {
	return o.resource
}
