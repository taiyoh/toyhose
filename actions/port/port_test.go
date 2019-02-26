package port_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taiyoh/toyhose/actions/port"
	"github.com/taiyoh/toyhose/errors"
)

func TestInputAndOutput(t *testing.T) {
	t.Run("compare arg", func(t *testing.T) {
		hoge := []byte("hoge")
		argBuffer := bytes.NewBuffer(hoge)
		input, _ := port.New(ioutil.NopCloser(argBuffer))
		if bytes.Compare(input.Arg(), hoge) != 0 {
			t.Error("wrong arg captured: expected: hoge", "actual:", string(input.Arg()))
		}
	})
	t.Run("correct response", func(t *testing.T) {
		hoge := []byte("hoge")
		argBuffer := bytes.NewBuffer(hoge)
		_, output := port.New(ioutil.NopCloser(argBuffer))
		recorder := httptest.NewRecorder()
		output.Set(map[string]interface{}{
			"foo": "bar",
		}, nil)
		output.Fill(recorder)
		if recorder.Code != http.StatusOK {
			t.Errorf(`msg="captured wrong response code" expected="200" actual="%d"`, recorder.Code)
		}
		if b := recorder.Body.String(); b != `{"foo":"bar"}` {
			t.Errorf(`msg="captured wrong body" expected='{"foo":"bar"}' actual="%s"`, b)
		}
	})
	t.Run("incorrect response", func(t *testing.T) {
		hoge := []byte("hoge")
		argBuffer := bytes.NewBuffer(hoge)
		_, output := port.New(ioutil.NopCloser(argBuffer))
		recorder := httptest.NewRecorder()
		output.Set(map[string]interface{}{
			"foo": "bar",
		}, errors.NewMissingParameter("aaa"))
		output.Fill(recorder)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf(`msg="captured wrong response code" expected="400" actual="%d"`, recorder.Code)
		}
		if b := recorder.Body.String(); b != `{"foo":"bar"}` {
			t.Errorf(`msg="captured wrong body" expected='{"foo":"bar"}' actual="%s"`, b)
		}
	})
}
