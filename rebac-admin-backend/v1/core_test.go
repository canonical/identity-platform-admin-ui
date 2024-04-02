// Copyright 2024 Canonical Ltd.

package v1

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/go-chi/chi/v5"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_authentication.go -source=./interfaces/authentication.go

// TestHandlerWorksWithStandardMux this test ensures that the returned Handler
// can be used with the Golang standard library multiplexers, and it's not tied
// to the underlying router library.
func TestHandlerWorksWithStandardMux(t *testing.T) {
	c := qt.New(t)

	sut, _ := NewReBACAdminBackend(ReBACAdminBackendParams{
		Authenticator: &noopAuthenticator{},
	})
	handler := sut.Handler("/some/base/path/")

	mux := http.NewServeMux()
	mux.Handle("/some/base/path/", handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	println(server.URL)

	res, err := http.Get(server.URL + "/some/base/path/v1/swagger.json")
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	defer res.Body.Close()

	out, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(len(out) > 0, qt.IsTrue)
}

// TestHandlerWorksWithChiMux this test ensures that the returned Handler
// can be used with the Chi multiplexers.
func TestHandlerWorksWithChiMux(t *testing.T) {
	c := qt.New(t)

	sut, _ := NewReBACAdminBackend(ReBACAdminBackendParams{
		Authenticator: &noopAuthenticator{},
	})
	handler := sut.Handler("")

	mux := chi.NewMux()
	mux.Mount("/some/base/path", handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	println(server.URL)

	res, err := http.Get(server.URL + "/some/base/path/v1/swagger.json")
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	defer res.Body.Close()

	out, err := io.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(len(out) > 0, qt.IsTrue)
}
