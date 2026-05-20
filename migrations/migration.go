// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0-only

package migrations

import "embed"

//go:embed *.sql
var EmbedMigrations embed.FS
