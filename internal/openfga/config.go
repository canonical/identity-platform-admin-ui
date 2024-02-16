// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package openfga

import (
	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type Config struct {
	ApiScheme   string
	ApiHost     string
	StoreID     string
	ApiToken    string
	AuthModelID string
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

	return c
}
