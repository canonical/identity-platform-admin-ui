// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"net/http"
	"time"
)

const (
	authCookiePath  = "/api/v0/auth/callback"
	nonceCookieName = "nonce"
	stateCookieName = "state"
)

var (
	epoch = time.Unix(0, 0).UTC()
)

type AuthCookieManager struct{}

func (a *AuthCookieManager) SetNonceCookie(w http.ResponseWriter, nonce string, ttl time.Duration) {
	a.setCookie(w, nonceCookieName, nonce, authCookiePath, ttl)
}

func (a *AuthCookieManager) GetNonceCookie(r *http.Request) string {
	return a.getCookie(r, nonceCookieName)
}

func (a *AuthCookieManager) ClearNonceCookie(w http.ResponseWriter) {
	a.clearCookie(w, nonceCookieName)
}

func (a *AuthCookieManager) SetStateCookie(w http.ResponseWriter, state string, ttl time.Duration) {
	a.setCookie(w, stateCookieName, state, authCookiePath, ttl)
}

func (a *AuthCookieManager) GetStateCookie(r *http.Request) string {
	return a.getCookie(r, stateCookieName)
}

func (a *AuthCookieManager) ClearStateCookie(w http.ResponseWriter) {
	a.clearCookie(w, stateCookieName)
}

func (a *AuthCookieManager) setCookie(w http.ResponseWriter, name, value string, path string, ttl time.Duration) {
	expires := time.Now().Add(ttl)
	http.SetCookie(w, &http.Cookie{
		Name: name,
		// TODO @barco: encrypt this value when cookie encryption is added to the codebase
		Value:    value,
		Path:     path,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
	})
}

func (a *AuthCookieManager) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Expires: epoch})
}

func (a *AuthCookieManager) getCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func NewAuthCookieManager() *AuthCookieManager {
	return new(AuthCookieManager)
}
