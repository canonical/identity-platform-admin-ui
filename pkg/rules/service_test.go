// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	reflect "reflect"
	"testing"

	oathkeeper "github.com/ory/oathkeeper-client-go"
	trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_interfaces.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_corev1.go k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,ConfigMapInterface
//go:generate mockgen -build_flags=--mod=mod -package rules -destination ./mock_oathkeeper.go github.com/ory/oathkeeper-client-go ApiApi

func TestListRulesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id := "mocked_rule1:allow"
	string_authorizer_handler := "allow"

	listRulesRequest := oathkeeper.ApiApiListRulesRequest{
		ApiService: mockOathkeeperApiApi,
	}

	mock_rules := []oathkeeper.Rule{
		{
			Authenticators: []oathkeeper.RuleHandler{
				{
					Handler: &string_athenticator_handler,
				},
			},
			Mutators: []oathkeeper.RuleHandler{
				{
					Handler: &string_mutator_handler,
				},
			},
			Id: &string_id,
			Authorizer: &oathkeeper.RuleHandler{
				Handler: &string_authorizer_handler,
			},
			Match: &oathkeeper.RuleMatch{
				Methods: []string{
					"GET",
					"POST",
				},
			},
		},
	}

	mockTracer.EXPECT().Start(ctx, "rules.Service.ListRules").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockOathkeeperApiApi.EXPECT().ListRules(ctx).Times(1).Return(listRulesRequest)
	mockOathkeeperApiApi.EXPECT().ListRulesExecute(gomock.Any()).Times(1).Return(mock_rules, new(http.Response), nil)

	rules, err := NewService(&config, mockTracer, mockMonitor, mockLogger).ListRules(ctx, 1, 100)

	if !reflect.DeepEqual(rules, mock_rules) {
		t.Fatalf("expected identities to be %v not  %v", mock_rules, rules)
	}
	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestListRulesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	listRulesRequest := oathkeeper.ApiApiListRulesRequest{
		ApiService: mockOathkeeperApiApi,
	}

	mock_rules := make([]oathkeeper.Rule, 0)

	mockTracer.EXPECT().Start(ctx, "rules.Service.ListRules").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockOathkeeperApiApi.EXPECT().ListRules(ctx).Times(1).Return(listRulesRequest)
	mockOathkeeperApiApi.EXPECT().ListRulesExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r oathkeeper.ApiApiListRulesRequest) ([]oathkeeper.Rule, *http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusInternalServerError,
						"details": map[string]interface{}{},
						"message": "error",
						"reason":  "error",
						"request": "mock_request",
						"status":  "Not Found",
					},
				},
			)

			return mock_rules, rr.Result(), fmt.Errorf("error")
		},
	)
	mockLogger.EXPECT().Error(gomock.Any()).Times(1)

	rules, err := NewService(&config, mockTracer, mockMonitor, mockLogger).ListRules(ctx, 1, 100)

	if len(rules) != 0 {
		t.Fatalf("expected rules to be empty list")
	}
	if err == nil {
		t.Fatal("expected error to be not nil")
	}

}

func TestGetRuleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	string_athenticator_handler := "cookie_session"
	string_mutator_handler := "header"
	string_id := "mocked_rule1:allow"
	string_authorizer_handler := "allow"

	getRulesRequest := oathkeeper.ApiApiGetRuleRequest{
		ApiService: mockOathkeeperApiApi,
	}

	mock_rule := oathkeeper.Rule{
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &string_athenticator_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &string_mutator_handler,
			},
		},
		Id: &string_id,
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &string_authorizer_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{
				"GET",
				"POST",
			},
		},
	}

	mock_rule_list := []oathkeeper.Rule{mock_rule}

	mockTracer.EXPECT().Start(ctx, "rules.Service.GetRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockOathkeeperApiApi.EXPECT().GetRule(ctx, string_id).Times(1).Return(getRulesRequest)
	mockOathkeeperApiApi.EXPECT().GetRuleExecute(gomock.Any()).Times(1).Return(&mock_rule, new(http.Response), nil)

	rules, err := NewService(&config, mockTracer, mockMonitor, mockLogger).GetRule(ctx, string_id)

	if !reflect.DeepEqual(rules, mock_rule_list) {
		t.Fatalf("expected identities to be %v not  %v", mock_rule_list, rules)
	}
	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestGetRuleFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	string_id := "mocked_rule1:allow"

	getRulesRequest := oathkeeper.ApiApiGetRuleRequest{
		ApiService: mockOathkeeperApiApi,
	}

	mock_rule := oathkeeper.Rule{}

	mockTracer.EXPECT().Start(ctx, "rules.Service.GetRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockOathkeeperApiApi.EXPECT().GetRule(ctx, string_id).Times(1).Return(getRulesRequest)
	mockOathkeeperApiApi.EXPECT().GetRuleExecute(gomock.Any()).Times(1).DoAndReturn(
		func(r oathkeeper.ApiApiGetRuleRequest) (*oathkeeper.Rule, *http.Response, error) {
			rr := httptest.NewRecorder()
			rr.Header().Set("Content-Type", "application/json")
			rr.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(rr).Encode(
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusInternalServerError,
						"details": map[string]interface{}{},
						"message": "error",
						"reason":  "error",
						"request": "mock_request",
						"status":  "Not Found",
					},
				},
			)

			return &mock_rule, rr.Result(), fmt.Errorf("error")
		},
	)
	mockLogger.EXPECT().Error(gomock.Any()).Times(1)

	_, err := NewService(&config, mockTracer, mockMonitor, mockLogger).GetRule(ctx, string_id)

	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

func TestUpdateRuleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	rules2_id := "mocked_rule2:deny"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"
	author_deny := "deny"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	rules2 := oathkeeper.Rule{
		Id: &rules2_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_deny,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"PUT", "DELETE"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1, rules2)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	ruleUpdate := oathkeeper.Rule{
		Id: &rules2_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_deny,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	ruleUpdatedList := make([]oathkeeper.Rule, 0)
	ruleUpdatedList = append(ruleUpdatedList, rules1, ruleUpdate)
	ruleUpdatedRaw, _ := json.Marshal(ruleUpdatedList)

	mockTracer.EXPECT().Start(ctx, "rules.Service.UpdateRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(config.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, "mock_config", gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, cm *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			rules := cm.Data[config.File]

			if ruleIncludedInMarshalledList(ruleUpdate, rules) {
				t.Fatalf("expected result to be %v not %v", string(ruleUpdatedRaw), rules)
			}

			return cm, nil
		},
	)

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).UpdateRule(ctx, rules2_id, ruleUpdate)

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}

}

func TestUpdateRuleNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	rules2_id := "mocked_rule2:deny"
	rules3_id := "mocked_rule3:deny"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"
	author_deny := "deny"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}
	rules2 := oathkeeper.Rule{
		Id: &rules2_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_deny,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"PUT", "DELETE"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1, rules2)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	ruleUpdate := oathkeeper.Rule{
		Id: &rules3_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	mockTracer.EXPECT().Start(ctx, "rules.Service.UpdateRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(config.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, "mock_config", gomock.Any()).Times(1).Return(cm, nil)

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).UpdateRule(ctx, rules3_id, ruleUpdate)

	expectedError := fmt.Sprintf("rule with ID %s not found", *ruleUpdate.Id)
	if err.Error() != expectedError {
		t.Fatalf("expected error to be %v not  %v", expectedError, err)
	}
}

func TestUpdateRuleIdMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	rules2_id := "mocked_rule2:deny"
	rules3_id := "mocked_rule3:deny"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"
	author_deny := "deny"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}
	rules2 := oathkeeper.Rule{
		Id: &rules2_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_deny,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"PUT", "DELETE"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1, rules2)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	rule_update := oathkeeper.Rule{
		Id: &rules3_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	mockTracer.EXPECT().Start(ctx, "rules.Service.UpdateRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).UpdateRule(ctx, rules1_id, rule_update)

	expectedError := fmt.Sprintf("The URL parameter id %s is different from payload rule id %s", rules1_id, *rule_update.Id)
	if err.Error() != expectedError {
		t.Fatalf("expected error to be %v not  %v", expectedError, err)
	}
}

func TestCreateRuleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	rules2_id := "mocked_rule2:deny"
	rules3_id := "mocked_rule3:deny"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"
	author_deny := "deny"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}
	rules2 := oathkeeper.Rule{
		Id: &rules2_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_deny,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"PUT", "DELETE"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1, rules2)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	ruleCreate := oathkeeper.Rule{
		Id: &rules3_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	ruleCreateList := make([]oathkeeper.Rule, 0)
	ruleCreateList = append(ruleCreateList, rules1, rules2, ruleCreate)
	ruleCreatedRaw, _ := json.Marshal(ruleCreateList)

	mockTracer.EXPECT().Start(ctx, "rules.Service.CreateRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(config.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, "mock_config", gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, cm *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			rules := cm.Data[config.File]
			if ruleIncludedInMarshalledList(ruleCreate, rules) {
				t.Fatalf("expected result to be %v not %v", string(ruleCreatedRaw), rules)
			}

			return cm, nil
		},
	)

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).CreateRule(ctx, ruleCreate)

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestCreateRuleAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	rules2_id := "mocked_rule2:deny"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"
	author_deny := "deny"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}
	rules2 := oathkeeper.Rule{
		Id: &rules2_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_deny,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"PUT", "DELETE"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1, rules2)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	ruleCreate := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	mockTracer.EXPECT().Start(ctx, "rules.Service.CreateRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(config.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, "mock_config", gomock.Any()).Times(1).Return(cm, nil)

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).CreateRule(ctx, ruleCreate)

	expected_error := fmt.Sprintf("rule with ID %s already exists", *ruleCreate.Id)
	if err.Error() != expected_error {
		t.Fatalf("expected error to be %s not  %s", expected_error, err.Error())
	}
}

func TestDeleteRuleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	mockTracer.EXPECT().Start(ctx, "rules.Service.DeleteRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(config.Namespace).Times(2).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, "mock_config", gomock.Any()).Times(1).Return(cm, nil)
	mockConfigMapV1.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, cm *v1.ConfigMap, opts metaV1.UpdateOptions) (*v1.ConfigMap, error) {
			rules := cm.Data[config.File]
			if !isMarshalledRuleListEmpty(rules) {
				t.Fatalf("expected rule %s to contain empty list, not %s", config.File, rules)
			}

			return cm, nil
		},
	)

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).DeleteRule(ctx, rules1_id)

	if err != nil {
		t.Fatalf("expected error to be nil not  %v", err)
	}
}

func TestDeleteRuleFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockLoggerInterface(ctrl)
	mockTracer := NewMockTracer(ctrl)
	mockMonitor := NewMockMonitorInterface(ctrl)
	mockOathkeeperApiApi := NewMockApiApi(ctrl)
	mockCoreV1 := NewMockCoreV1Interface(ctrl)
	mockConfigMapV1 := NewMockConfigMapInterface(ctrl)

	ctx := context.Background()
	config := Config{
		Name:      "mock_config",
		File:      "admin_ui_rules.json",
		Namespace: "mock_namespace",
		K8s:       mockCoreV1,
		OkClient:  mockOathkeeperApiApi,
	}

	rules1_id := "mocked_rule1:allow"
	authen_handler := "cookie_session"
	mutator_header := "header"
	author_handler := "allow"

	rules1 := oathkeeper.Rule{
		Id: &rules1_id,
		Authenticators: []oathkeeper.RuleHandler{
			{
				Handler: &authen_handler,
			},
		},
		Mutators: []oathkeeper.RuleHandler{
			{
				Handler: &mutator_header,
			},
		},
		Authorizer: &oathkeeper.RuleHandler{
			Handler: &author_handler,
		},
		Match: &oathkeeper.RuleMatch{
			Methods: []string{"GET", "POST"},
		},
	}

	ruleList := make([]oathkeeper.Rule, 0)
	ruleList = append(ruleList, rules1)

	rawRuleList, _ := json.Marshal(ruleList)
	cm := new(v1.ConfigMap)
	cm.Data = make(map[string]string)
	cm.Data[config.File] = string(rawRuleList)

	ruleForDeletion := "mocked_rule3:deny"

	mockTracer.EXPECT().Start(ctx, "rules.Service.DeleteRule").Times(1).Return(ctx, trace.SpanFromContext(ctx))
	mockCoreV1.EXPECT().ConfigMaps(config.Namespace).Times(1).Return(mockConfigMapV1)
	mockConfigMapV1.EXPECT().Get(ctx, "mock_config", gomock.Any()).Times(1).Return(cm, nil)

	err := NewService(&config, mockTracer, mockMonitor, mockLogger).DeleteRule(ctx, ruleForDeletion)

	expected_error := fmt.Sprintf("rule with ID %s not found", ruleForDeletion)
	if err.Error() != expected_error {
		t.Fatalf("expected error to be %v not  %v", expected_error, err.Error())
	}
}

func ruleIncludedInMarshalledList(rule oathkeeper.Rule, ruleList string) bool {
	rules := make([]oathkeeper.Rule, 0)
	err := json.Unmarshal([]byte(ruleList), &rules)

	if err != nil {
		return false
	}

	for _, v := range rules {
		if v.Id == rule.Id {
			return true
		}
	}

	return false
}

func isMarshalledRuleListEmpty(ruleList string) bool {
	rules := make([]oathkeeper.Rule, 0)
	err := json.Unmarshal([]byte(ruleList), &rules)

	if err != nil {
		return false
	}

	return len(rules) == 0
}
