// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import "net/http"

// Authenticator defines an abstract backend to perform authentication on HTTP requests.
type Authenticator interface {
	// Authenticate receives an HTTP request and returns the identity of the caller.
	// The same identity will be available to the service backend through the request
	// context.
	//
	// Note that the implementations of this method should not alter the state of the
	// received request instance.
	Authenticate(r http.Request) (any, error)
}
