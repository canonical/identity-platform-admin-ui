// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import (
	"net/http"
	"time"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
)

const (
	defaultCookiePath = "/"
	authCookiePath    = "/api/v0/auth/callback"
	nonceCookieName   = "nonce"
	stateCookieName   = "state"
	itCookieName      = "id-token"
	atCookieName      = "access-token"
	rtCookieName      = "refresh-token"
)

var (
	epoch = time.Unix(0, 0).UTC()
)

type AuthCookieManager struct {
	authCookiesTTL       time.Duration
	userSessionCookieTTL time.Duration
	encrypt              EncryptInterface

	logger logging.LoggerInterface
}

func (a *AuthCookieManager) SetIDTokenCookie(w http.ResponseWriter, rawIDToken string) {
	a.setCookie(w, itCookieName, rawIDToken, defaultCookiePath, a.userSessionCookieTTL)
}

func (a *AuthCookieManager) GetIDTokenCookie(r *http.Request) string {
	return a.getCookie(r, itCookieName)
}

func (a *AuthCookieManager) ClearIDTokenCookie(w http.ResponseWriter) {
	a.clearCookie(w, itCookieName, defaultCookiePath)
}

func (a *AuthCookieManager) SetAccessTokenCookie(w http.ResponseWriter, rawAccessToken string) {
	a.setCookie(w, atCookieName, rawAccessToken, defaultCookiePath, a.userSessionCookieTTL)
}

func (a *AuthCookieManager) GetAccessTokenCookie(r *http.Request) string {
	return a.getCookie(r, atCookieName)
}

func (a *AuthCookieManager) ClearAccessTokenCookie(w http.ResponseWriter) {
	a.clearCookie(w, atCookieName, defaultCookiePath)
}

func (a *AuthCookieManager) SetRefreshTokenCookie(w http.ResponseWriter, rawRefreshToken string) {
	a.setCookie(w, rtCookieName, rawRefreshToken, defaultCookiePath, a.userSessionCookieTTL)
}

func (a *AuthCookieManager) GetRefreshTokenCookie(r *http.Request) string {
	return a.getCookie(r, rtCookieName)
}

func (a *AuthCookieManager) ClearRefreshTokenCookie(w http.ResponseWriter) {
	a.clearCookie(w, rtCookieName, defaultCookiePath)
}

func (a *AuthCookieManager) SetNonceCookie(w http.ResponseWriter, nonce string) {
	a.setCookie(w, nonceCookieName, nonce, authCookiePath, a.authCookiesTTL)
}

func (a *AuthCookieManager) GetNonceCookie(r *http.Request) string {
	return a.getCookie(r, nonceCookieName)
}

func (a *AuthCookieManager) ClearNonceCookie(w http.ResponseWriter) {
	a.clearCookie(w, nonceCookieName, authCookiePath)
}

func (a *AuthCookieManager) SetStateCookie(w http.ResponseWriter, state string) {
	a.setCookie(w, stateCookieName, state, authCookiePath, a.authCookiesTTL)
}

func (a *AuthCookieManager) GetStateCookie(r *http.Request) string {
	return a.getCookie(r, stateCookieName)
}

func (a *AuthCookieManager) ClearStateCookie(w http.ResponseWriter) {
	a.clearCookie(w, stateCookieName, authCookiePath)
}

func (a *AuthCookieManager) setCookie(w http.ResponseWriter, name, value string, path string, ttl time.Duration) {
	if value == "" {
		return
	}

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
		Domain:   "",
		Expires:  expires,
		MaxAge:   int(ttl.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *AuthCookieManager) clearCookie(w http.ResponseWriter, name string, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   "",
		Expires:  epoch,
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
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

func NewAuthCookieManager(
	authCookiesTTLSeconds,
	userSessionCookieTTLSeconds int,
	encrypt EncryptInterface,
	logger logging.LoggerInterface,
) *AuthCookieManager {
	a := new(AuthCookieManager)
	a.authCookiesTTL = time.Duration(authCookiesTTLSeconds) * time.Second
	a.userSessionCookieTTL = time.Duration(userSessionCookieTTLSeconds) * time.Second
	a.encrypt = encrypt

	a.logger = logger
	return a

}
