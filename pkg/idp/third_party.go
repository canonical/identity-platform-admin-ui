// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import "encoding/json"

const (
	SUPPORTED_PROVIDERS = "generic google github githubapp gitlab microsoft discord slack facebook auth0 vk yandex apple spotify netid dingtalk linkedin patreon"
)

// TODO @shipperizer once import of library is fixed find a way to use this schema with extra yaml annotations
// coming from https://pkg.go.dev/github.com/ory/kratos@v1.0.0/selfservice/strategy/oidc#Configuration
// importing the library github.com/ory/kratos fails due to compilations on their side
type Configuration struct {
	// ID is the provider's ID
	ID string `json:"id" yaml:"id" validate:"required"`

	// Provider is either "generic" for a generic OAuth 2.0 / OpenID Connect Provider or one of:
	// - generic
	// - google
	// - github
	// - github-app
	// - gitlab
	// - microsoft
	// - discord
	// - slack
	// - facebook
	// - auth0
	// - vk
	// - yandex
	// - apple
	// - spotify
	// - netid
	// - dingtalk
	// - linkedin
	// - patreon
	// validate that Provider is not nil and not empty string, and its value is one of the supported ones
	Provider string `json:"provider" yaml:"provider" validate:"required,supported_provider"`

	// Label represents an optional label which can be used in the UI generation.
	Label string `json:"label"`

	// ClientID is the application's Client ID.
	ClientID string `json:"client_id" yaml:"client_id"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"client_secret" yaml:"client_secret"`

	// IssuerURL is the OpenID Connect Server URL. You can leave this empty if `provider` is not set to `generic`.
	// If set, neither `auth_url` nor `token_url` are required.
	// validate that this field is required only when Provider field == "generic"
	IssuerURL string `json:"issuer_url" yaml:"issuer_url"`

	// AuthURL is the authorize url, typically something like: https://example.org/oauth2/auth
	// Should only be used when the OAuth2 / OpenID Connect server is not supporting OpenID Connect Discovery and when
	// `provider` is set to `generic`.
	// validate that this field is required only when Provider field == "generic" and IssuerURL is empty
	AuthURL string `json:"auth_url" yaml:"auth_url"`

	// TokenURL is the token url, typically something like: https://example.org/oauth2/token
	// Should only be used when the OAuth2 / OpenID Connect server is not supporting OpenID Connect Discovery and when
	// `provider` is set to `generic`.
	// validate that this field is required only when Provider field == "generic" and IssuerURL is empty
	TokenURL string `json:"token_url" yaml:"token_url"`

	// Tenant is the Azure AD Tenant to use for authentication, and must be set when `provider` is set to `microsoft`.
	// Can be either `common`, `organizations`, `consumers` for a multitenant application or a specific tenant like
	// `8eaef023-2b34-4da1-9baa-8bc8c9d6a490` or `contoso.onmicrosoft.com`.
	// validate that this field is required only when Provider field == "microsoft"
	Tenant string `json:"microsoft_tenant" yaml:"microsoft_tenant" validate:"required_if=Provider microsoft"`

	// SubjectSource is a flag which controls from which endpoint the subject identifier is taken by microsoft provider.
	// Can be either `userinfo` or `me`.
	// If the value is `userinfo` then the subject identifier is taken from sub field of userinfo standard endpoint response.
	// If the value is `me` then the `id` field of https://graph.microsoft.com/v1.0/me response is taken as subject.
	// The default is `userinfo`.
	// validate that SubjectSource is either "userinfo" or "me"
	SubjectSource string `json:"subject_source" yaml:"subject_source" validate:"oneof=userinfo me"`

	// TeamId is the Apple Developer Team ID that's needed for the `apple` `provider` to work.
	// It can be found Apple Developer website and combined with `apple_private_key` and `apple_private_key_id`
	// is used to generate `client_secret`
	// validate that this field is required only when Provider field == "apple"
	TeamId string `json:"apple_team_id" yaml:"apple_team_id" validate:"required_if=Provider apple"`

	// PrivateKeyId is the private Apple key identifier. Keys can be generated via developer.apple.com.
	// This key should be generated with the `Sign In with Apple` option checked.
	// This is needed when `provider` is set to `apple`
	// validate that this field is required only when Provider field == "apple"
	PrivateKeyId string `json:"apple_private_key_id" yaml:"apple_private_key_id" validate:"required_if=Provider apple"`

	// PrivateKeyId is the Apple private key identifier that can be downloaded during key generation.
	// This is needed when `provider` is set to `apple`
	// validate that this field is required only when Provider field == "apple"
	PrivateKey string `json:"apple_private_key" yaml:"apple_private_key" validate:"required_if=Provider apple"`

	// Scope specifies optional requested permissions.
	// validate that Scope is either nil or a slice with non-empty strings
	Scope []string `json:"scope" yaml:"scope" validate:"omitempty,dive,required"`

	// Mapper specifies the JSONNet code snippet which uses the OpenID Connect Provider's data (e.g. GitHub or Google
	// profile information) to hydrate the identity's data.
	//
	// It can be either a URL (file://, http(s)://, base64://) or an inline JSONNet code snippet.
	Mapper string `json:"mapper_url" yaml:"mapper_url"`

	// RequestedClaims string encoded json object that specifies claims and optionally their properties which should be
	// included in the id_token or returned from the UserInfo Endpoint.
	//
	// More information: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter
	RequestedClaims json.RawMessage `json:"requested_claims" yaml:"requested_claims"`
}
