// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"context"
	"encoding/json"
	"fmt"
	reflect "reflect"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
	gomock "go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/canonical/rebac-admin-ui-handlers/v1/interfaces"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	"github.com/google/uuid"
)

//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_corev1.go k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,ConfigMapInterface

func TestListResourcesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	mockTracer.EXPECT().Start(ctx, "idp.Service.ListResources").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).ListResources(ctx)

	if !reflect.DeepEqual(is, idps) {
		t.Fatalf("expected providers to be %v not  %v", idps, is)

	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestListResourcesSuccessButEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = ""

	mockTracer.EXPECT().Start(ctx, "idp.Service.ListResources").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).ListResources(ctx)

	if len(is) != 0 {
		t.Fatalf("expected providers to be empty not  %v", is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestListResourcesFailsOnConfigMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	mockTracer.EXPECT().Start(ctx, "idp.Service.ListResources").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(nil, fmt.Errorf("broken"))
	mockLogger.EXPECT().Error(gomock.Any()).Times(1)
	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).ListResources(ctx)

	if is != nil {
		t.Fatalf("expected result to be nil not  %v", is)

	}

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}
}

func TestListResourcesFailsOnMissingKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data["random"] = string(rawIdps)

	mockTracer.EXPECT().Start(ctx, "idp.Service.ListResources").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).ListResources(ctx)

	if is != nil && len(is) > 0 {
		t.Fatalf("expected result to be an empty slice not  %v", is)

	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestGetResourceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	mockTracer.EXPECT().Start(ctx, "idp.Service.GetResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).GetResource(ctx, idps[0].ID)

	if !reflect.DeepEqual(is[0], idps[0]) {
		t.Fatalf("expected providers to be %v not  %v", idps[0], is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestGetResourceNotfound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	mockTracer.EXPECT().Start(ctx, "idp.Service.GetResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).GetResource(ctx, "fake")

	if len(is) != 0 {
		t.Fatalf("expected providers to be empty not  %v", is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestGetResourceSuccessButEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = ""

	mockTracer.EXPECT().Start(ctx, "idp.Service.GetResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).GetResource(ctx, "fake")

	if len(is) != 0 {
		t.Fatalf("expected providers to be empty not  %v", is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestGetResourceFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = ""

	mockTracer.EXPECT().Start(ctx, "idp.Service.GetResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).GetResource(ctx, "fake")

	if len(is) != 0 {
		t.Fatalf("expected providers to be empty not  %v", is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestEditResourceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	c := new(Configuration)
	c.ClientSecret = "secret-9"

	mockTracer.EXPECT().Start(ctx, "idp.Service.EditResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).EditResource(ctx, idps[0].ID, c)

	if is[0].ClientSecret != c.ClientSecret {
		t.Fatalf("expected provider secret to be %v not  %v", c.ClientSecret, is[0].ClientSecret)

	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestEditResourceNotfound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	c := new(Configuration)
	c.ClientSecret = "secret-9"

	mockTracer.EXPECT().Start(ctx, "idp.Service.EditResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).EditResource(ctx, "fake", c)

	if len(is) != 0 {
		t.Fatalf("expected providers to be empty not  %v", is)
	}

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}
}

func TestEditResourceFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	c := new(Configuration)
	c.ClientSecret = "secret-9"

	mockTracer.EXPECT().Start(ctx, "idp.Service.EditResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).Return(cm, fmt.Errorf("error"))

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).EditResource(ctx, idps[0].ID, c)

	if is != nil {
		t.Fatalf("expected providers to be nil, not %v", is)
	}

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}
}

func TestCreateResourceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	c := new(Configuration)
	c.ClientSecret = "secret-9"
	c.ID = "okta_347646e49b484037b83690b020f9f629"
	c.ClientID = "347646e4-9b48-4037-b836-90b020f9f629"
	c.Provider = "okta"
	c.Mapper = "file:///etc/config/kratos/okta_schema.jsonnet"
	c.Scope = []string{"email"}

	mockTracer.EXPECT().Start(ctx, "idp.Service.CreateResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockAuthz.EXPECT().SetCreateProviderEntitlements(gomock.Any(), gomock.Any())
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, configMap *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			i := make([]*Configuration, 0)

			rawIdps := configMap.Data[cfg.KeyName]

			_ = yaml.Unmarshal([]byte(rawIdps), &i)

			if len(i) != len(idps)+1 {
				t.Fatalf("expected providers to be %v not %v", len(idps)+1, len(i))
			}
			return cm, nil
		},
	)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).CreateResource(ctx, c)

	if is == nil {
		t.Fatalf("expected provider to be not nil %v", is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestCreateResourceSuccessSetsIDIfMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = ""

	c := new(Configuration)
	c.ClientSecret = "secret-9"
	c.ClientID = "347646e4-9b48-4037-b836-90b020f9f629"
	c.Provider = "okta"
	c.Mapper = "file:///etc/config/kratos/okta_schema.jsonnet"
	c.Scope = []string{"email"}

	mockTracer.EXPECT().Start(ctx, "idp.Service.CreateResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockAuthz.EXPECT().SetCreateProviderEntitlements(gomock.Any(), gomock.Any())
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, configMap *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			i := make([]*Configuration, 0)

			rawIdps, ok := configMap.Data[cfg.KeyName]

			if !ok {
				t.Fatalf("key is missing from the configmap")
			}

			_ = yaml.Unmarshal([]byte(rawIdps), &i)

			if len(i) != 1 {
				t.Fatalf("expected providers to be %v not %v", 1, len(i))
			}

			if i[0].ID == "" {
				t.Fatalf("expected ID to have defaulted to a uuid, not %s", i[0].ID)
			}

			return cm, nil
		},
	)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).CreateResource(ctx, c)

	if is == nil {
		t.Fatalf("expected provider to be not nil %v", is)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestCreateResourceFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	c := new(Configuration)
	c.ClientSecret = "secret-9"
	c.ID = "okta_347646e49b484037b83690b020f9f629"
	c.ClientID = "347646e4-9b48-4037-b836-90b020f9f629"
	c.Provider = "okta"
	c.Mapper = "file:///etc/config/kratos/okta_schema.jsonnet"
	c.Scope = []string{"email"}

	mockTracer.EXPECT().Start(ctx, "idp.Service.CreateResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, configMap *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			i := make([]*Configuration, 0)

			rawIdps := configMap.Data[cfg.KeyName]

			_ = yaml.Unmarshal([]byte(rawIdps), &i)

			if len(i) != len(idps)+1 {
				t.Fatalf("expected providers to be %v not %v", len(idps)+1, len(i))
			}
			return nil, fmt.Errorf("error")
		},
	)

	is, err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).CreateResource(ctx, c)

	if is != nil {
		t.Fatalf("expected provider to be nil not %v", is)
	}

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}
}

func TestDeleteResourceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	mockTracer.EXPECT().Start(ctx, "idp.Service.DeleteResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockAuthz.EXPECT().SetDeleteProviderEntitlements(gomock.Any(), idps[0].ID)
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, configMap *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			i := make([]*Configuration, 0)

			rawIdps := configMap.Data[cfg.KeyName]

			_ = yaml.Unmarshal([]byte(rawIdps), &i)

			if len(i) != len(idps)-1 {
				t.Fatalf("expected providers to be %v not %v", len(idps)+1, len(i))
			}
			return cm, nil
		},
	)

	err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).DeleteResource(ctx, idps[0].ID)

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestDeleteResourceFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)
	ctx := context.Background()

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	rawIdps, _ := json.Marshal(idps)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[cfg.KeyName] = string(rawIdps)

	mockTracer.EXPECT().Start(ctx, "idp.Service.DeleteResource").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)

	err := NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger).DeleteResource(ctx, "fake")

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}
}

func TestV1ServiceImplementsRebacServiceInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var svc interface{} = new(V1Service)

	if _, ok := svc.(interfaces.IdentityProvidersService); !ok {
		t.Fatalf("V1Service doesnt implement interfaces.IdentityProvidersService")
	}
}

func TestV1ServiceListIdentityProviders(t *testing.T) {
	type expected struct {
		err  error
		idps []resources.IdentityProvider
	}

	k8sIdps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	idps := make([]resources.IdentityProvider, 0)
	enabled := true

	for _, idp := range k8sIdps {
		idps = append(idps,
			resources.IdentityProvider{
				Id:           &idp.ID,
				ClientID:     &idp.ClientID,
				ClientSecret: &idp.ClientSecret,
				Name:         &idp.Label,
				Enabled:      &enabled,
			},
		)
	}

	tests := []struct {
		name     string
		expected expected
	}{
		{
			name: "empty result",
			expected: expected{
				idps: []resources.IdentityProvider{},
				err:  nil,
			},
		},
		{
			name: "error",
			expected: expected{
				idps: nil,
				err:  fmt.Errorf("Internal Server Error: error"),
			},
		},
		{
			name: "full result",
			expected: expected{
				idps: idps,
				err:  nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.KeyName = "idps.yaml"
			cfg.Name = "idps"
			cfg.Namespace = "default"

			svc := NewV1Service(
				NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger),
			)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
			mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, name string, opts metaV1.GetOptions) (*v1.ConfigMap, error) {
					if test.expected.err != nil {
						return nil, test.expected.err
					}

					cm := new(v1.ConfigMap)
					cm.Data = make(map[string]string)

					rawIdps := make([]*Configuration, 0)
					for _, i := range test.expected.idps {
						rawIdps = append(rawIdps, svc.castProvider(ctx, &i))
					}

					rawIdpsJson, _ := json.Marshal(rawIdps)
					cm.Data[cfg.KeyName] = string(rawIdpsJson)

					return cm, nil
				},
			)

			r, err := svc.ListIdentityProviders(
				ctx,
				&resources.GetIdentityProvidersParams{},
			)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if !reflect.DeepEqual(r.Data, test.expected.idps) {
				t.Errorf("expected idps to be %v not %v", test.expected.idps, r.Data)
			}
		})
	}
}

func TestV1ServiceListAvailableIdentityProviders(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	svc := NewV1Service(
		NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger),
	)

	r, err := svc.ListAvailableIdentityProviders(
		ctx,
		&resources.GetAvailableIdentityProvidersParams{},
	)

	idps := make([]string, 0)

	for _, i := range r.Data {
		idps = append(idps, i.Id)
	}

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}

	if strings.Join(idps, " ") != SUPPORTED_PROVIDERS {
		t.Fatalf("expected providers to be %s not  %v", SUPPORTED_PROVIDERS, idps)
	}
}

func TestV1ServiceRegisterConfiguration(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockAuthz := NewMockAuthorizerInterface(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))

	cfg := new(Config)
	cfg.K8s = mockCoreV1
	cfg.KeyName = "idps.yaml"
	cfg.Name = "idps"
	cfg.Namespace = "default"

	svc := NewV1Service(
		NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger),
	)

	r, err := svc.RegisterConfiguration(
		ctx,
		&resources.IdentityProvider{},
	)

	if err == nil {
		t.Fatalf("expected error not to be nil %v", err)
	}

	if r != nil {
		t.Fatalf("expected result to be nil not %v", r)
	}
}

func TestV1ServiceDeleteConfiguration(t *testing.T) {
	type expected struct {
		err error
		ok  bool
	}

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	tests := []struct {
		name     string
		expected expected
	}{
		{
			name: "success",
			expected: expected{
				ok:  true,
				err: nil,
			},
		},
		{
			name: "error",
			expected: expected{
				ok:  false,
				err: fmt.Errorf("Internal Server Error: error"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.KeyName = "idps.yaml"
			cfg.Name = "idps"
			cfg.Namespace = "default"

			svc := NewV1Service(
				NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger),
			)

			rawIdps, _ := json.Marshal(idps)
			cm := new(v1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[cfg.KeyName] = string(rawIdps)

			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockAuthz.EXPECT().SetDeleteProviderEntitlements(gomock.Any(), idps[0].ID).MinTimes(0).MaxTimes(1)
			mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(2).Return(mockConfigMapV1)
			mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).Return(cm, nil)
			mockConfigMapV1.EXPECT().Update(gomock.Any(), cm, gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, configMap *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
					i := make([]*Configuration, 0)

					rawIdps := configMap.Data[cfg.KeyName]

					_ = yaml.Unmarshal([]byte(rawIdps), &i)

					if test.expected.err != nil {
						return cm, test.expected.err
					}

					if len(i) != len(idps)-1 {
						t.Fatalf("expected providers to be %v not %v", len(idps)+1, len(i))
					}

					return cm, nil
				},
			)

			ok, err := svc.DeleteConfiguration(
				ctx,
				idps[0].ID,
			)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not %v", test.expected.err, err)
			}

			if test.expected.ok != ok {
				t.Errorf("expected result to be %v not %v", test.expected.ok, ok)
			}
		})
	}
}

func TestV1ServiceGetConfiguration(t *testing.T) {
	type expected struct {
		err error
		idp *resources.IdentityProvider
	}

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	enabled := true

	tests := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "empty result",
			input: uuid.NewString(),
			expected: expected{
				idp: nil,
				err: fmt.Errorf("Internal Server Error: error"),
			},
		},
		{
			name:  "success",
			input: idps[0].ID,
			expected: expected{
				idp: &resources.IdentityProvider{
					Id:           &idps[0].ID,
					ClientID:     &idps[0].ClientID,
					ClientSecret: &idps[0].ClientSecret,
					Name:         &idps[0].Label,
					Enabled:      &enabled,
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.KeyName = "idps.yaml"
			cfg.Name = "idps"
			cfg.Namespace = "default"

			svc := NewV1Service(
				NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger),
			)

			rawIdps, _ := json.Marshal(idps)
			cm := new(v1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[cfg.KeyName] = string(rawIdps)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).Times(1).Return(mockConfigMapV1)
			mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, name string, opts metaV1.GetOptions) (*v1.ConfigMap, error) {
					cm := new(v1.ConfigMap)
					cm.Data = make(map[string]string)
					rawIdpsJson, _ := json.Marshal(idps)
					cm.Data[cfg.KeyName] = string(rawIdpsJson)

					return cm, nil
				},
			)

			r, err := svc.GetConfiguration(
				ctx,
				test.input,
			)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if !reflect.DeepEqual(r, test.expected.idp) {
				t.Errorf("expected idp to be %v not %v", test.expected.idp, r)
			}
		})
	}
}

func TestV1ServiceUpdateConfiguration(t *testing.T) {
	type expected struct {
		err error
		idp *resources.IdentityProvider
	}

	idps := []*Configuration{
		{
			ID:           "microsoft_af675f353bd7451588e2b8032e315f6f",
			ClientID:     "af675f35-3bd7-4515-88e2-b8032e315f6f",
			Provider:     "microsoft",
			ClientSecret: "secret-1",
			Tenant:       "e1574293-28de-4e94-87d5-b61c76fc14e1",
			Mapper:       "file:///etc/config/kratos/microsoft_schema.jsonnet",
			Scope:        []string{"email"},
		},
		{
			ID:           "google_18fa2999e6c9475aa49515d933d8e8ce",
			ClientID:     "18fa2999-e6c9-475a-a495-15d933d8e8ce",
			Provider:     "google",
			ClientSecret: "secret-2",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"email", "profile"},
		},
		{
			ID:           "aws_18fa2999e6c9475aa49589d941d8e1zy",
			ClientID:     "18fa2999-e6c9-475a-a495-89d941d8e1zy",
			Provider:     "aws",
			ClientSecret: "secret-3",
			Mapper:       "file:///etc/config/kratos/google_schema.jsonnet",
			Scope:        []string{"address", "profile"},
		},
	}

	enabled := true
	newSecret := "new-secret-1"
	randomUUID := uuid.NewString()

	tests := []struct {
		name     string
		input    *resources.IdentityProvider
		expected expected
	}{
		{
			name: "empty result",
			input: &resources.IdentityProvider{
				Id:           &randomUUID,
				ClientID:     &idps[0].ClientID,
				ClientSecret: &idps[0].ClientSecret,
			},
			expected: expected{
				idp: nil,
				err: fmt.Errorf("Internal Server Error: error"),
			},
		},
		{
			name: "success",
			input: &resources.IdentityProvider{
				Id:           &idps[0].ID,
				ClientID:     &idps[0].ClientID,
				ClientSecret: &newSecret,
			},
			expected: expected{
				idp: &resources.IdentityProvider{
					Id:           &idps[0].ID,
					ClientID:     &idps[0].ClientID,
					ClientSecret: &newSecret,
					Name:         &idps[0].Label,
					Enabled:      &enabled,
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracer(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockAuthz := NewMockAuthorizerInterface(ctrl)
			mockCoreV1 := NewMockCoreV1Interface(ctrl)
			mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

			cfg := new(Config)
			cfg.K8s = mockCoreV1
			cfg.KeyName = "idps.yaml"
			cfg.Name = "idps"
			cfg.Namespace = "default"

			svc := NewV1Service(
				NewService(cfg, mockAuthz, mockTracer, mockMonitor, mockLogger),
			)

			rawIdps, _ := json.Marshal(idps)
			cm := new(v1.ConfigMap)
			cm.Data = make(map[string]string)
			cm.Data[cfg.KeyName] = string(rawIdps)

			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockTracer.EXPECT().Start(ctx, gomock.Any()).AnyTimes().Return(ctx, trace.SpanFromContext(ctx))
			mockCoreV1.EXPECT().ConfigMaps(cfg.Namespace).MinTimes(1).MaxTimes(2).Return(mockConfigMapV1)

			mockConfigMapV1.EXPECT().Get(ctx, cfg.Name, gomock.Any()).Times(1).DoAndReturn(
				func(ctx context.Context, name string, opts metaV1.GetOptions) (*v1.ConfigMap, error) {
					cm := new(v1.ConfigMap)
					cm.Data = make(map[string]string)
					rawIdpsJson, _ := json.Marshal(idps)
					cm.Data[cfg.KeyName] = string(rawIdpsJson)

					return cm, nil
				},
			)
			mockConfigMapV1.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).MinTimes(0).MaxTimes(1).DoAndReturn(
				func(ctx context.Context, configMap *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {

					idpConfig := make([]*Configuration, 0)
					_ = json.Unmarshal([]byte(configMap.Data[cfg.KeyName]), &idpConfig)

					if idpConfig[0].ClientSecret != newSecret {
						t.Errorf("expected client secret to have change")
					}

					idps[0].ClientSecret = newSecret
					cm := new(v1.ConfigMap)
					cm.Data = make(map[string]string)
					rawIdpsJson, _ := json.Marshal(idps)
					cm.Data[cfg.KeyName] = string(rawIdpsJson)

					return cm, nil
				},
			)

			r, err := svc.UpdateConfiguration(
				ctx,
				test.input,
			)

			if test.expected.err != nil && err == nil {
				t.Errorf("expected error to be %v not %v", test.expected.err, err)
			}

			if test.expected.err != nil {
				return
			}

			if !reflect.DeepEqual(r, test.expected.idp) {
				t.Errorf("expected idp to be %v not %v", test.expected.idp, r)
			}
		})
	}
}
