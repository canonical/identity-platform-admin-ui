// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import "net/http"

// Authenticator defines an abstract backend to perform authentication on HTTP requests.
type Authenticator interface {
	// Authenticate receives an HTTP request and returns the identity of the caller.
	// The same identity will be available to the service backend through the request
	// context. To avoid issues with value types, it's best to return a pointer type.
	//
	// Note that the implementations of this method should not alter the state of the
	// received request instance.
	//
	// If the returned identity is nil it will be regarded as authentication failure.
	//
	// To return an error, the implementations should use the provided error functions
	// (e.g., `NewAuthenticationError`) and avoid creating ad-hoc errors.
	Authenticate(r *http.Request) (any, error)
}
