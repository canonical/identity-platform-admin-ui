// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package status

import (
	"context"
	"testing"

	v0Status "github.com/canonical/identity-platform-api/v0/status"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package status -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package status -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package status -destination ./mock_tracer.go 	go.opentelemetry.io/otel/trace Tracer

func TestGrpcHandler_GetStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		want *v0Status.Status
	}{
		{
			name: "Status OK",
			want: &v0Status.Status{
				Status: "Ok",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSpan := trace.SpanFromContext(context.TODO())

			mockTracer.EXPECT().Start(gomock.Any(), "status.GrpcHandler.GetStatus").Return(context.TODO(), mockSpan)

			g := NewGrpcHandler(mockTracer, mockMonitor, mockLogger)

			got, _ := g.GetStatus(context.TODO(), nil)

			if got.GetStatus() != tt.want.GetStatus() {
				t.Errorf("expected status: %v, got %v", tt.want.GetStatus(), got.GetStatus())
			}
		})
	}
}
