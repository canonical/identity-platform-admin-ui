// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"net/http"
	"time"
)

var (
	EPOCH             = time.Unix(0, 0)
	NONCE_COOKIE_NAME = "nonce"
)

func SetNonceCookie(w http.ResponseWriter, nonce string, ttl time.Duration) {
	expires := time.Now().Add(ttl)
	http.SetCookie(w, &http.Cookie{
		Name: NONCE_COOKIE_NAME,
		// TODO @barco: encrypt this value when cookie encryption is added to the codebase
		Value:    nonce,
		Path:     "/api/v0/login",
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
	})
}

func GetNonceCookie(r *http.Request) string {
	cookie, err := r.Cookie(NONCE_COOKIE_NAME)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func ClearNonceCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: NONCE_COOKIE_NAME, Expires: EPOCH})
}
