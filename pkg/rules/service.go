// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package rules

import (
	"context"
	"encoding/json"
	"fmt"

	oathkeeper "github.com/ory/oathkeeper-client-go"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Config struct {
	Name      string
	File      string
	Namespace string
	K8s       coreV1.CoreV1Interface
	OkClient  oathkeeper.ApiApi
}

func NewConfig(cmName, cmFile, cmNamespace string, k8s coreV1.CoreV1Interface, oathkeeper oathkeeper.ApiApi) *Config {
	rulesConfig := Config{
		K8s:       k8s,
		Name:      cmName,
		File:      cmFile,
		Namespace: cmNamespace,
		OkClient:  oathkeeper,
	}

	return &rulesConfig
}

type Service struct {
	oathkeeper oathkeeper.ApiApi
	authz      AuthorizerInterface

	cmName      string
	cmFileName  string
	cmNamespace string

	k8s coreV1.CoreV1Interface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *Service) ListRules(ctx context.Context, offset, size int64) ([]oathkeeper.Rule, error) {
	ctx, span := s.tracer.Start(ctx, "rules.Service.ListRules")
	defer span.End()

	rules, _, err := s.oathkeeper.ListRules(ctx).Limit(size).Offset(offset).Execute()

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return rules, nil
}

func (s *Service) GetRule(ctx context.Context, id string) ([]oathkeeper.Rule, error) {
	ctx, span := s.tracer.Start(ctx, "rules.Service.GetRule")
	defer span.End()

	rule, _, err := s.oathkeeper.GetRule(ctx, id).Execute()

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	rules := make([]oathkeeper.Rule, 0)
	rules = append(rules, *rule)
	return rules, nil
}

func (s *Service) UpdateRule(ctx context.Context, id string, updatedRule oathkeeper.Rule) error {
	ctx, span := s.tracer.Start(ctx, "rules.Service.UpdateRule")
	defer span.End()

	//check for inconsistency between id url parameter and the id in payload object
	if id != *updatedRule.Id {
		return fmt.Errorf("The URL parameter id %s is different from payload rule id %s", id, *updatedRule.Id)
	}

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	rules, err := s.extractAdminRules(cm.Data)
	if err != nil {
		return err
	}

	if _, ok := rules[id]; !ok {
		return fmt.Errorf("rule with ID %s not found", id)
	}

	rules[id] = &updatedRule

	rawRuleList, err := s.marshalRuleMap(rules)
	if err != nil {
		return err
	}

	cm.Data[s.cmFileName] = rawRuleList

	if _, err = s.k8s.ConfigMaps(s.cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateRule(ctx context.Context, newRule oathkeeper.Rule) error {
	ctx, span := s.tracer.Start(ctx, "rules.Service.CreateRule")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	//Here I need to check the ID variable of all rules to make sure there's no collusion
	rulesReadable, err := s.extractReadableRules(cm.Data)
	if _, ok := rulesReadable[*newRule.Id]; ok {
		return fmt.Errorf("rule with ID %s already exists", *newRule.Id)
	}

	rulesWritable, err := s.extractAdminRules(cm.Data)
	if err != nil {
		return err
	}

	rulesWritable[*newRule.Id] = &newRule

	rawRuleList, err := s.marshalRuleMap(rulesWritable)
	if err != nil {
		return err
	}

	cm.Data[s.cmFileName] = rawRuleList

	if _, err = s.k8s.ConfigMaps(s.cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {
		return err
	}

	s.authz.SetCreateRuleEntitlements(ctx, *newRule.Id)

	return nil
}

func (s *Service) DeleteRule(ctx context.Context, id string) error {
	ctx, span := s.tracer.Start(ctx, "rules.Service.DeleteRule")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	rules, err := s.extractAdminRules(cm.Data)
	if err != nil {
		return err
	}

	if _, ok := rules[id]; !ok {
		return fmt.Errorf("rule with ID %s not found", id)
	}

	delete(rules, id)

	rawRuleList, err := s.marshalRuleMap(rules)
	if err != nil {
		return err
	}

	cm.Data[s.cmFileName] = rawRuleList

	if _, err = s.k8s.ConfigMaps(s.cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {
		return err
	}

	s.authz.SetDeleteRuleEntitlements(ctx, id)

	return nil
}

// The input parameter is always cm.Data
// Output map includes the rules in ADMIN_RULE_FILE, not all rules in the configmap.
func (s *Service) extractAdminRules(data map[string]string) (map[string]*oathkeeper.Rule, error) {

	ruleMap := make(map[string]*oathkeeper.Rule)

	ruleList := make([]*oathkeeper.Rule, 0)
	rawRuleList, ok := data[s.cmFileName]
	if !ok {
		return ruleMap, nil
	}

	err := json.Unmarshal([]byte(rawRuleList), &ruleList)

	if err != nil {
		s.logger.Errorf("failed unmarshalling %s - %v", rawRuleList, err)
		return nil, err
	}

	for _, v := range ruleList {
		ruleMap[*v.Id] = v
	}

	return ruleMap, nil
}

// The input parameter is always cm.Data
// Output map includes the all rules in the configmap, do not use the output for creating/editing/deleting rules in the configmap.
func (s *Service) extractReadableRules(data map[string]string) (map[string]*oathkeeper.Rule, error) {

	ruleMap := make(map[string]*oathkeeper.Rule)

	for file, rawRuleList := range data {

		ruleList := make([]*oathkeeper.Rule, 0)
		err := json.Unmarshal([]byte(rawRuleList), &ruleList)

		if err != nil {
			s.logger.Errorf("failed unmarshalling ruleset from file %s: %s - %v", file, rawRuleList, err)
			return nil, err
		}

		for _, v := range ruleList {
			ruleMap[*v.Id] = v
		}
	}
	return ruleMap, nil
}

// rules are stored in the configmap as a list of oathkeeper rules
func (s *Service) marshalRuleMap(rules map[string]*oathkeeper.Rule) (string, error) {
	ruleList := make([]*oathkeeper.Rule, 0)
	for _, v := range rules {
		ruleList = append(ruleList, v)
	}

	rawRuleList, err := json.Marshal(ruleList)

	if err != nil {
		s.logger.Errorf("failed marshalling %s - %v", rawRuleList, err)
		return "", err
	}

	return string(rawRuleList), nil
}

func NewService(config *Config, authz AuthorizerInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	if config == nil {
		panic("empty config for rules service")
	}

	s.oathkeeper = config.OkClient
	s.k8s = config.K8s
	s.cmName = config.Name
	s.cmFileName = config.File
	s.cmNamespace = config.Namespace
	s.authz = authz

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
