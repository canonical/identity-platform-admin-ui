// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"

	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/interfaces"
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

//go:generate mockgen -package interfaces -destination ./interfaces/mock_authentication.go -source=./interfaces/authentication.go
//go:generate mockgen -package v1 -destination ./mock_error_response.go -source=./error.go

// noopAuthenticator is a no-op implementation of Authenticator interface, to be
// used in tests.
type noopAuthenticator struct{}

func (a noopAuthenticator) Authenticate(r *http.Request) (any, error) {
	return struct{}{}, nil
}

func TestContextualAuthenticatedIdentity(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		name     string
		identity any
	}{{
		name:     "identity as a struct",
		identity: struct{ foo string }{foo: "bar"},
	}, {
		name:     "identity as a struct pointer",
		identity: &struct{ foo string }{foo: "bar"},
	}, {
		name:     "identity as a string",
		identity: "some-identity",
	}, {
		name:     "identity as a string pointer",
		identity: stringPtr("some-identity"),
	}, {
		name:     "identity as a slice pointer",
		identity: &[]string{"foo"},
	}, {
		name:     "identity as a map pointer",
		identity: &map[string]any{"foo": "bar"},
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
			baseRequest, err := http.NewRequest(http.MethodGet, "/blah", nil)
			c.Assert(err, qt.IsNil)

			req := newRequestWithIdentityInContext(baseRequest, tt.identity)
			c.Assert(req, qt.IsNotNil)

			fetchedIdentity, err := GetIdentityFromContext(req.Context())
			c.Assert(err, qt.IsNil)
			c.Assert(fetchedIdentity, qt.Equals, tt.identity)
		})
	}
}

func TestContextualAuthenticatedIdentity_UnsetIdentity(t *testing.T) {
	c := qt.New(t)

	req, err := http.NewRequest(http.MethodGet, "/blah", nil)
	c.Assert(err, qt.IsNil)

	fetchedIdentity, err := GetIdentityFromContext(req.Context())
	c.Assert(err, qt.ErrorMatches, "Unauthorized: authentication failed: missing caller identity")
	c.Assert(fetchedIdentity, qt.IsNil)
}

func TestContextualAuthenticatedIdentity_MiddlewareAndContext(t *testing.T) {
	c := qt.New(t)

	defaultRequest, _ := http.NewRequest(http.MethodGet, "/blah", nil)

	writeResponse := func(w http.ResponseWriter, status int, body any) {
		raw, _ := json.Marshal(body)
		w.WriteHeader(status)
		_, _ = w.Write(raw)
	}

	tests := []struct {
		name               string
		setupRequest       func() *http.Request
		authenticatorFunc  func(r *http.Request) (any, error)
		mapErrorFunc       func(error) *resources.Response
		nextHandler        func(c *qt.C, w http.ResponseWriter, r *http.Request)
		expectedStatusCode int
		expectedMessage    string
	}{{
		name: "authentication successful",
		authenticatorFunc: func(r *http.Request) (any, error) {
			return "some-identity", nil
		},
		nextHandler: func(c *qt.C, w http.ResponseWriter, r *http.Request) {
			identity, err := GetIdentityFromContext(r.Context())
			c.Assert(err, qt.IsNil)
			c.Assert(identity, qt.Equals, "some-identity")

			writeResponse(w, http.StatusOK, resources.Response{
				Status:  http.StatusOK,
				Message: "done",
			})
		},
		expectedStatusCode: http.StatusOK,
		expectedMessage:    "done",
	}, {
		name: "authenticator returns error",
		authenticatorFunc: func(r *http.Request) (any, error) {
			return nil, NewAuthenticationError("some error")
		},
		expectedStatusCode: http.StatusUnauthorized,
		expectedMessage:    "Unauthorized: authentication failed: some error",
	}, {
		name: "authenticator returns error (non-nil error mapper)",
		authenticatorFunc: func(r *http.Request) (any, error) {
			return nil, errors.New("some error")
		},
		mapErrorFunc: func(err error) *resources.Response {
			return &resources.Response{
				Status:  999, // Some bizarre code
				Message: "mapped error message",
			}
		},
		expectedStatusCode: 999, // The same bizarre code
		expectedMessage:    "mapped error message",
	}, {
		name: "authenticator returns nil identity",
		authenticatorFunc: func(r *http.Request) (any, error) {
			return nil, nil
		},
		expectedStatusCode: http.StatusUnauthorized,
		expectedMessage:    "Unauthorized: authentication failed: nil identity",
	},
	}

	for _, t := range tests {
		tt := t
		c.Run(tt.name, func(c *qt.C) {
			ctrl := gomock.NewController(c)
			defer ctrl.Finish()

			req := defaultRequest
			if tt.setupRequest != nil {
				req = tt.setupRequest()
			}

			var mockAuthenticator interfaces.Authenticator = &noopAuthenticator{}
			if tt.authenticatorFunc != nil {
				mockIdentity, mockAuthError := tt.authenticatorFunc(req)
				authenticator := interfaces.NewMockAuthenticator(ctrl)
				authenticator.EXPECT().Authenticate(req).Return(mockIdentity, mockAuthError)
				mockAuthenticator = authenticator
			}

			var mockErrorMapper ErrorResponseMapper
			if tt.mapErrorFunc != nil {
				mapper := NewMockErrorResponseMapper(ctrl)
				mapper.EXPECT().MapError(gomock.Any()).DoAndReturn(tt.mapErrorFunc)
				mockErrorMapper = mapper
			}

			sut, err := NewReBACAdminBackend(ReBACAdminBackendParams{
				Authenticator:            mockAuthenticator,
				AuthenticatorErrorMapper: mockErrorMapper,
			})
			c.Assert(err, qt.IsNil)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.nextHandler(c, w, r)
			})

			recorder := httptest.NewRecorder()
			sut.authenticationMiddleware()(next).ServeHTTP(recorder, req)

			c.Assert(recorder.Code, qt.Equals, tt.expectedStatusCode)

			raw, err := io.ReadAll(recorder.Body)
			c.Assert(err, qt.IsNil)

			parsedResponse := &resources.Response{}
			err = json.Unmarshal(raw, parsedResponse)
			c.Assert(err, qt.IsNil)
			c.Assert(parsedResponse.Status, qt.Equals, tt.expectedStatusCode)
			c.Assert(parsedResponse.Message, qt.Matches, tt.expectedMessage)
		})
	}
}
