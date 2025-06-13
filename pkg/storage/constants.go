// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package storage

// Roles
const (
	ROLE_TABLE    = "role"
	ROLE_NAME     = "role.name"
	ROLE_ID       = "role.id"
	ROLE_OWNER    = "role.owner"
	ROLE_OWNER_ID = "role.owner_id"
)

// Groups
const (
	GROUP_TABLE    = "group"
	GROUP_ID       = "group.id"
	GROUP_NAME     = "group.name"
	GROUP_OWNER    = "group.owner"
	GROUP_OWNER_ID = "group.owner_id"
)

// Join tables
const (
	GROUP_ROLE_TABLE    = "group_role"
	GROUP_ROLE_GROUP_ID = "group_role.group_id"
	GROUP_ROLE_ROLE_ID  = "group_role.role_id"
)
