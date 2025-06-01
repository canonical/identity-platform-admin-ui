// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/canonical/rebac-admin-ui-handlers/v1"
	"github.com/canonical/rebac-admin-ui-handlers/v1/resources"
	"github.com/google/uuid"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type Config struct {
	Name      string
	Namespace string
	KeyName   string
	K8s       coreV1.CoreV1Interface
}

type Service struct {
	cmName      string
	cmNamespace string
	keyName     string

	k8s   coreV1.CoreV1Interface
	authz AuthorizerInterface

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *Service) ListResources(ctx context.Context) ([]*Configuration, error) {
	ctx, span := s.tracer.Start(ctx, "idp.Service.ListResources")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return s.idpConfiguration(cm.Data)

}

func (s *Service) GetResource(ctx context.Context, providerID string) ([]*Configuration, error) {
	ctx, span := s.tracer.Start(ctx, "idp.Service.GetResource")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	idps, err := s.idpConfiguration(cm.Data)

	if err != nil {
		return nil, err
	}

	if idps == nil {
		return nil, nil
	}

	// TODO @shipperizer find a better way to index the idps
	for _, idp := range idps {
		if idp.ID == providerID {
			return []*Configuration{idp}, nil
		}
	}
	return []*Configuration{}, nil
}

func (s *Service) EditResource(ctx context.Context, providerID string, data *Configuration) ([]*Configuration, error) {
	ctx, span := s.tracer.Start(ctx, "idp.Service.EditResource")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	idps, err := s.idpConfiguration(cm.Data)

	if err != nil {
		return nil, err
	}

	if idps == nil {
		return nil, nil
	}

	var idp *Configuration
	// TODO @shipperizer find a better way to index the idps
	for _, i := range idps {
		if i.ID == providerID {
			i = s.mergeConfiguration(i, data)
			idp = i
		}
	}

	if idp == nil {
		return []*Configuration{}, fmt.Errorf("provider with ID %s not found", providerID)
	}

	rawIdps, err := json.Marshal(idps)

	if err != nil {
		return nil, err
	}

	cm.Data[s.keyName] = string(rawIdps)

	if _, err = s.k8s.ConfigMaps(s.cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {

		return nil, err
	}

	return []*Configuration{idp}, nil

}

func (s *Service) CreateResource(ctx context.Context, data *Configuration) ([]*Configuration, error) {
	ctx, span := s.tracer.Start(ctx, "idp.Service.CreateResource")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	idps, err := s.idpConfiguration(cm.Data)

	if err != nil {
		return nil, err
	}
	if idps == nil {
		return nil, nil
	}

	// assign random ID if empty
	if data.ID == "" {
		data.ID = uuid.NewString()
	}

	// TODO @shipperizer find a better way to index the idps
	for _, i := range idps {
		if i.ID == data.ID {
			return idps, fmt.Errorf("provider with same ID already exists")
		}
	}

	idps = append(idps, data)

	rawIdps, err := json.Marshal(idps)

	if err != nil {
		return nil, err
	}

	// catch if configmap is empty and initialize
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	cm.Data[s.keyName] = string(rawIdps)

	if _, err = s.k8s.ConfigMaps(s.cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {

		return nil, err
	}

	s.authz.SetCreateProviderEntitlements(ctx, data.ID)

	return []*Configuration{data}, nil
}

func (s *Service) DeleteResource(ctx context.Context, providerID string) error {
	ctx, span := s.tracer.Start(ctx, "idp.Service.DeleteResource")
	defer span.End()

	cm, err := s.k8s.ConfigMaps(s.cmNamespace).Get(ctx, s.cmName, metaV1.GetOptions{})

	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	var found bool
	idps, err := s.idpConfiguration(cm.Data)

	if err != nil {
		return err
	}

	if idps == nil {
		return nil
	}

	newIdps := make([]*Configuration, 0)

	// TODO @shipperizer find a better way to index the idps
	for _, i := range idps {

		if i.ID == providerID {
			found = true
		} else {
			newIdps = append(newIdps, i)
		}
	}

	if !found {
		return fmt.Errorf("provider with ID %s not found", providerID)
	}

	rawIdps, err := json.Marshal(newIdps)

	if err != nil {
		return err
	}

	cm.Data[s.keyName] = string(rawIdps)

	if _, err = s.k8s.ConfigMaps(s.cmNamespace).Update(ctx, cm, metaV1.UpdateOptions{}); err != nil {

		return err
	}
	s.authz.SetDeleteProviderEntitlements(ctx, providerID)

	return nil

}

// TODO @shipperizer ugly but safe, other way is to json/yaml Marshal/Unmarshal and use omitempty
func (s *Service) mergeConfiguration(base, update *Configuration) *Configuration {
	if update.Provider != "" {
		base.Provider = update.Provider
	}

	if update.Label != "" {
		base.Provider = update.Provider
	}

	if update.ClientID != "" {
		base.Provider = update.Provider
	}

	if update.ClientSecret != "" {
		base.ClientSecret = update.ClientSecret
	}

	if update.IssuerURL != "" {
		base.IssuerURL = update.IssuerURL
	}

	if update.AuthURL != "" {
		base.AuthURL = update.AuthURL
	}

	if update.TokenURL != "" {
		base.TokenURL = update.TokenURL
	}

	if update.Tenant != "" {
		base.Tenant = update.Tenant
	}

	if update.SubjectSource != "" {
		base.SubjectSource = update.SubjectSource
	}

	if update.TeamId != "" {
		base.TeamId = update.TeamId
	}

	if update.PrivateKeyId != "" {
		base.PrivateKeyId = update.PrivateKeyId
	}

	if update.PrivateKey != "" {
		base.PrivateKey = update.PrivateKey
	}

	if update.Scope != nil && len(update.Scope) > 0 {
		base.Scope = update.Scope
	}

	if update.Mapper != "" {
		base.Mapper = update.Mapper
	}

	if update.RequestedClaims != nil {
		base.RequestedClaims = update.RequestedClaims
	}

	return base
}

func (s *Service) idpConfiguration(idps map[string]string) ([]*Configuration, error) {

	idpConfig := make([]*Configuration, 0)

	rawIdps, ok := idps[s.keyName]
	if !ok {
		s.logger.Errorf("failed to find key %s in configMap %v", s.keyName, idps)
		return idpConfig, nil
	}

	err := json.Unmarshal([]byte(rawIdps), &idpConfig)

	if err != nil {
		s.logger.Errorf("failed unmarshalling %s - %v", rawIdps, err)
		return nil, fmt.Errorf("failed unmarshalling %s - %v", rawIdps, err)
	}

	return idpConfig, nil
}

func (s *Service) keyIDMapper(id, namespace string) string {
	return uuid.NewSHA1(uuid.Nil, []byte(fmt.Sprintf("%s.%s", id, namespace))).String()
}

// TODO @shipperizer analyze if providers IDs need to be what we use for path or if filename is the right one
func NewService(config *Config, authz AuthorizerInterface, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Service {
	s := new(Service)

	if config == nil {
		panic("empty config for IDP service")
	}

	s.k8s = config.K8s
	s.cmName = config.Name
	s.cmNamespace = config.Namespace
	s.authz = authz
	// TODO @shipperizer fetch it from the config.KeyName
	s.keyName = "idps.yaml"

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}

type V1Service struct {
	core *Service
}

// ListAvailableIdentityProviders returns the static list of supported identity providers.
func (s *V1Service) ListAvailableIdentityProviders(ctx context.Context, params *resources.GetAvailableIdentityProvidersParams) (*resources.PaginatedResponse[resources.AvailableIdentityProvider], error) {
	_, span := s.core.tracer.Start(ctx, "idp.V1Service.ListAvailableIdentityProviders")
	defer span.End()

	idps := make([]resources.AvailableIdentityProvider, 0)

	for _, i := range strings.Split(SUPPORTED_PROVIDERS, " ") {
		idps = append(
			idps,
			resources.AvailableIdentityProvider{Id: i, Name: &i},
		)
	}

	r := new(resources.PaginatedResponse[resources.AvailableIdentityProvider])
	r.Meta = resources.ResponseMeta{Size: len(idps)}
	r.Data = idps

	return r, nil
}

// ListIdentityProviders returns a list of registered identity providers configurations.
func (s *V1Service) ListIdentityProviders(ctx context.Context, params *resources.GetIdentityProvidersParams) (*resources.PaginatedResponse[resources.IdentityProvider], error) {
	ctx, span := s.core.tracer.Start(ctx, "idp.V1Service.ListIdentityProviders")
	defer span.End()

	idps, err := s.core.ListResources(ctx)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	r := new(resources.PaginatedResponse[resources.IdentityProvider])
	r.Data = make([]resources.IdentityProvider, 0)
	r.Meta = resources.ResponseMeta{Size: len(idps)}

	// this caters for nil slice and empty slice
	if len(idps) == 0 {
		return r, nil
	}

	for _, idp := range idps {
		if idp == nil {
			continue
		}

		r.Data = append(
			r.Data,
			// TODO @shipperizer see f any other field can match
			*s.castConfiguration(ctx, idp),
		)
	}

	return r, nil
}

// RegisterConfiguration register a new authentication provider configuration.
func (s *V1Service) RegisterConfiguration(ctx context.Context, provider *resources.IdentityProvider) (*resources.IdentityProvider, error) {
	_, span := s.core.tracer.Start(ctx, "idp.V1Service.ListIdentityProviders")
	defer span.End()

	// TODO @shipperizer mismatch between required information from kratos and what is available on payload
	// cannot be filled, need to rework v1 api
	return nil, v1.NewNotImplementedError("use /api/v0/idps endpoint")
}

// DeleteConfiguration removes an authentication provider configuration identified by `id`.
func (s *V1Service) DeleteConfiguration(ctx context.Context, id string) (bool, error) {
	ctx, span := s.core.tracer.Start(ctx, "idp.V1Service.DeleteConfiguration")
	defer span.End()

	err := s.core.DeleteResource(ctx, id)

	return err == nil, err
}

// GetConfiguration returns the authentication provider configuration identified by `id`.
func (s *V1Service) GetConfiguration(ctx context.Context, id string) (*resources.IdentityProvider, error) {
	ctx, span := s.core.tracer.Start(ctx, "idp.V1Service.GetConfiguration")
	defer span.End()

	idps, err := s.core.GetResource(ctx, id)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	// this caters for nil slice and empty slice
	if len(idps) == 0 {
		return nil, v1.NewNotFoundError(fmt.Sprintf("provider %s not found", id))
	}

	if len(idps) != 1 {
		return nil, v1.NewUnknownError("multiple providers with the same id found")
	}

	return s.castConfiguration(ctx, idps[0]), nil
}

// UpdateConfiguration update the authentication provider configuration identified by `id`.
func (s *V1Service) UpdateConfiguration(ctx context.Context, provider *resources.IdentityProvider) (*resources.IdentityProvider, error) {
	ctx, span := s.core.tracer.Start(ctx, "idp.V1Service.UpdateConfiguration")
	defer span.End()

	if provider == nil {
		return nil, v1.NewMissingRequestBodyError("missing provider payload")
	}

	idps, err := s.core.EditResource(
		ctx,
		*provider.Id,
		s.castProvider(ctx, provider),
	)

	if err != nil {
		return nil, v1.NewUnknownError(err.Error())
	}

	// this caters for nil slice and empty slice
	if len(idps) == 0 {
		return nil, v1.NewNotFoundError(fmt.Sprintf("provider %s not found", *provider.Id))
	}

	if len(idps) != 1 {
		return nil, v1.NewUnknownError("multiple providers with the same id found")
	}

	return s.castConfiguration(ctx, idps[0]), nil
}

func (s *V1Service) castProvider(_ context.Context, provider *resources.IdentityProvider) *Configuration {
	if provider == nil {
		return nil
	}

	cfg := new(Configuration)

	if provider.Id != nil {
		cfg.ID = *provider.Id
	}

	if provider.ClientID != nil {
		cfg.ClientID = *provider.ClientID
	}

	if provider.ClientSecret != nil {
		cfg.ClientSecret = *provider.ClientSecret
	}

	if provider.Name != nil {
		cfg.Label = *provider.Name
	}

	return cfg
}

func (s *V1Service) castConfiguration(_ context.Context, provider *Configuration) *resources.IdentityProvider {
	if provider == nil {
		return nil
	}

	enabled := true

	return &resources.IdentityProvider{
		Id:           &provider.ID,
		ClientID:     &provider.ClientID,
		ClientSecret: &provider.ClientSecret,
		Name:         &provider.Label,
		Enabled:      &enabled,
	}
}

func NewV1Service(svc *Service) *V1Service {
	s := new(V1Service)

	s.core = svc

	return s
}
