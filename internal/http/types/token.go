// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package types

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/metadata"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

const (
	PAGINATION_HEADER = "X-Token-Pagination"
	// GRPC_PAGINATION_METADATA  all gRPC metadata are lowercase
	GRPC_PAGINATION_METADATA = "x-token-pagination"
)

// TODO @shipperizer move this under openfga package or at least change name to reflect this is used for openfga
// related endpoints

type TokenPaginator struct {
	tokens map[string]string

	tracer tracing.TracingInterface
	logger logging.LoggerInterface
}

// LoadFromString populates the TokenPaginator struct with pagination tokens from a string
func (p *TokenPaginator) LoadFromString(ctx context.Context, s string) error {
	tokenMap, err := base64.StdEncoding.DecodeString(s)

	if err != nil {
		p.logger.Errorf("issues decoding header: %s", err)
		return err
	}

	tokens := map[string]string{}

	err = json.Unmarshal(tokenMap, &tokens)

	if err != nil {
		p.logger.Errorf("issues parsing header: %s", err)
		return err
	}

	p.SetTokens(context.TODO(), tokens)

	return nil
}

// LoadFromRequest populates the TokenPaginator struct with pagination tokens from the r request
func (p *TokenPaginator) LoadFromRequest(ctx context.Context, r *http.Request) error {
	_, span := p.tracer.Start(ctx, "types.TokenPaginator.LoadFromRequest")
	defer span.End()

	header := r.Header.Get(PAGINATION_HEADER)

	if header == "" {
		return nil
	}

	return p.LoadFromString(ctx, header)
}

// LoadFromGRPCContext populates the TokenPaginator struct with pagination tokens from the gRPC context metadata
func (p *TokenPaginator) LoadFromGRPCContext(ctx context.Context) error {
	_, span := p.tracer.Start(ctx, "types.TokenPaginator.LoadFromGRPCContext")
	defer span.End()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	var header []string
	if header, ok = md[GRPC_PAGINATION_METADATA]; !ok || len(header) == 0 || header[0] == "" {
		return nil
	}

	return p.LoadFromString(ctx, header[0])
}

// SetToken sets a pagination token value for the specified type represented by key
// if the pagination token is an empty string, SetToken will be a noop
func (p *TokenPaginator) SetToken(ctx context.Context, key, value string) {
	if value != "" {
		p.tokens[key] = value
	}
}

// SetTokens sets its internal pagination tokens map to a copy of the provided map
// if any of the pagination tokens is an empty string, it will not be set
func (p *TokenPaginator) SetTokens(ctx context.Context, tokens map[string]string) {
	if tokens == nil {
		return
	}

	for key := range p.tokens {
		delete(p.tokens, key)
	}

	for key, value := range tokens {
		if value != "" {
			p.tokens[key] = value
		}
	}
}

// GetToken returns the token value mapped to type "key", or empty string if key is not present
func (p *TokenPaginator) GetToken(ctx context.Context, key string) string {
	if token, ok := p.tokens[key]; ok {
		return token
	}

	return ""
}

func (p *TokenPaginator) GetAllTokens(ctx context.Context) map[string]string {
	return p.tokens
}

// PaginationHeader returns a composite pagination token string to use as a header
func (p *TokenPaginator) PaginationHeader(ctx context.Context) (string, error) {
	_, span := p.tracer.Start(ctx, "types.TokenPaginator.PaginationHeader")
	defer span.End()

	if len(p.tokens) == 0 {
		return "", nil
	}

	tokenMap, err := json.Marshal(p.tokens)

	if err != nil {
		p.logger.Errorf("issues parsing tokens: %s", err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(tokenMap), nil
}

func NewTokenPaginator(tracer tracing.TracingInterface, logger logging.LoggerInterface) *TokenPaginator {
	p := new(TokenPaginator)

	p.logger = logger
	p.tracer = tracer
	p.tokens = make(map[string]string)

	return p

}
