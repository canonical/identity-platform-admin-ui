// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package types

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
	"go.opentelemetry.io/otel/trace"
)

const (
	PAGINATION_HEADER = "X-Token-Pagination"
)

type TokenPaginator struct {
	tokens map[string]string

	tracer tracing.TracingInterface
	logger logging.LoggerInterface
}

func (p *TokenPaginator) LoadFromRequest(ctx context.Context, r *http.Request) error {
	_, span := p.tracer.Start(ctx, "types.TokenPaginator.LoadFromRequest")
	defer span.End()

	header := r.Header.Get(PAGINATION_HEADER)

	if header == "" {
		return nil
	}

	tokenMap, err := base64.StdEncoding.DecodeString(header)

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

	p.tokens = tokens

	return nil
}

func (p *TokenPaginator) SetToken(ctx context.Context, key, value string) {
	p.tokens[key] = value
}

func (p *TokenPaginator) GetToken(ctx context.Context, key string) string {
	if token, ok := p.tokens[key]; ok {
		return token
	}

	return ""
}

func (p *TokenPaginator) GetAllTokens(ctx context.Context) map[string]string {
	return p.tokens
}

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

func NewTokenPaginator(tracer trace.Tracer, logger logging.LoggerInterface) *TokenPaginator {
	p := new(TokenPaginator)

	p.logger = logger
	p.tracer = tracer
	p.tokens = make(map[string]string)

	return p

}
