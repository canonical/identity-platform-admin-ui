// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package openfga

import (
	validator "github.com/go-playground/validator/v10"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type Config struct {
	ApiScheme   string `validate:"required"`
	ApiHost     string `validate:"required"`
	StoreID     string `validate:"required"`
	ApiToken    string `validate:"required"`
	AuthModelID string `validate:"required"`
	Debug       bool

	Tracer  tracing.TracingInterface
	Monitor monitoring.MonitorInterface
	Logger  logging.LoggerInterface
}

func NewConfig(apiScheme, apiHost, storeID, apiToken, authModelID string, debug bool, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *Config {
	c := new(Config)

	c.ApiScheme = apiScheme
	c.ApiHost = apiHost
	c.StoreID = storeID
	c.ApiToken = apiToken
	c.AuthModelID = authModelID
	c.Debug = debug

	c.Monitor = monitor
	c.Tracer = tracer
	c.Logger = logger

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(c); err != nil {
		logger.Errorf("invalid config object: %s", err)

		return nil
	}

	return c
}
