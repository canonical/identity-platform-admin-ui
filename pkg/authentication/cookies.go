// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"net/http"
	"time"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
)

const (
	authCookiePath  = "/api/v0/auth/callback"
	nonceCookieName = "nonce"
	stateCookieName = "state"
)

var (
	epoch = time.Unix(0, 0).UTC()
)

type AuthCookieManager struct {
	encrypt EncryptInterface
	logger  logging.LoggerInterface
}

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

	encrypted, err := a.encrypt.Encrypt(value)
	if err != nil {
		a.logger.Errorf("can't encrypt cookie value, %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    encrypted,
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
		a.logger.Errorf("can't get cookie %s, %v", name, err)
		return ""
	}

	value, err := a.encrypt.Decrypt(cookie.Value)
	if err != nil {
		a.logger.Errorf("can't decrypt cookie value, %v", err)
		return ""
	}
	return value
}

func NewAuthCookieManager(encrypt EncryptInterface, logger logging.LoggerInterface) *AuthCookieManager {
	a := new(AuthCookieManager)
	a.encrypt = encrypt
	a.logger = logger
	return a

}
