// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type TracingInterface interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}
