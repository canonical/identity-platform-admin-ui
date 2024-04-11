// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package groups

import (
	"net/http"
	"strings"
)

func shouldValidate(r *http.Request) bool {
	return r.Method == http.MethodPost || r.Method == http.MethodPatch
}

func isCreateGroup(r *http.Request, endpoint string) bool {
	return r.Method == http.MethodPost && endpoint == ""
}

func isUpdateGroup(r *http.Request, endpoint string) bool {
	// make sure at least one character is present for the Group ID URL Param
	return r.Method == http.MethodPatch && strings.HasPrefix(endpoint, "/") && len(endpoint) > 1
}

func isAssignRoles(r *http.Request, endpoint string) bool {
	return r.Method == http.MethodPost && strings.HasSuffix(endpoint, "/roles")
}

func isAssignPermissions(r *http.Request, endpoint string) bool {
	return r.Method == http.MethodPatch && strings.HasSuffix(endpoint, "/entitlements")
}

func isAssignIdentities(r *http.Request, endpoint string) bool {
	return r.Method == http.MethodPatch && strings.HasSuffix(endpoint, "/identities")
}
