// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authentication

import "context"

type principalContextKey int

var PrincipalContextKey principalContextKey

type UserPrincipal struct {
	Subject   string `json:"sub"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	SessionID string `json:"sid"`
	Nonce     string `json:"nonce"`

	RawAccessToken  string `json:"-"`
	RawIdToken      string `json:"-"`
	RawRefreshToken string `json:"-"`
}

func (u *UserPrincipal) Session() string {
	return u.SessionID
}

func (u *UserPrincipal) AccessToken() string {
	return u.RawAccessToken
}

func (u *UserPrincipal) RefreshToken() string {
	return u.RawRefreshToken
}

func (u *UserPrincipal) IDToken() string {
	return u.RawIdToken
}

func (u *UserPrincipal) Identifier() string {
	return u.Email
}

type ServicePrincipal struct {
	Subject        string `json:"sub"`
	RawAccessToken string `json:"-"`
}

func (s *ServicePrincipal) Session() string {
	return ""
}

func (s *ServicePrincipal) AccessToken() string {
	return s.RawAccessToken
}

func (s *ServicePrincipal) RefreshToken() string {
	return ""
}

func (s *ServicePrincipal) IDToken() string {
	return ""
}

func (s *ServicePrincipal) Identifier() string {
	return s.Subject
}

func NewUserPrincipalFromClaims(c ReadableClaims) (*UserPrincipal, error) {
	a := new(UserPrincipal)
	if err := c.Claims(a); err != nil {
		return nil, err
	}
	return a, nil
}

func NewServicePrincipalFromClaims(c ReadableClaims) (*ServicePrincipal, error) {
	a := new(ServicePrincipal)
	if err := c.Claims(a); err != nil {
		return nil, err
	}

	return a, nil
}

func PrincipalContext(ctx context.Context, principal PrincipalInterface) context.Context {
	parent := ctx
	if ctx == nil {
		parent = context.Background()
	}

	if principal == nil {
		return parent
	}

	return context.WithValue(parent, PrincipalContextKey, principal)
}

func PrincipalFromContext(ctx context.Context) PrincipalInterface {
	if ctx == nil {
		return nil
	}

	if value, ok := ctx.Value(PrincipalContextKey).(PrincipalInterface); ok {
		return value
	}

	return nil
}
