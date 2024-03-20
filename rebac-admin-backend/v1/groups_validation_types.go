// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"github.com/canonical/identity-platform-admin-ui/rebac-admin-backend/v1/resources"
)

func validateGroup(value *resources.Group) error {
	if len(value.Name) == 0 {
		return NewRequestBodyValidationError("empty group name")
	}
	return nil
}

func validateGroupEntitlementsPatchRequestBody(value *resources.GroupEntitlementsPatchRequestBody) error {
	if len(value.Patches) == 0 {
		return NewRequestBodyValidationError("empty patch array")
	}
	return validateSlice[resources.GroupEntitlementsPatchItem](value.Patches, validateGroupEntitlementsPatchItem)
}

func validateGroupEntitlementsPatchItem(value *resources.GroupEntitlementsPatchItem) error {
	if err := validateEntityEntitlement(&value.Entitlement); err != nil {
		return err
	}
	return validateStringEnum("op", value.Op, resources.GroupEntitlementsPatchItemOpAdd, resources.GroupEntitlementsPatchItemOpRemove)
}

func validateEntityEntitlement(value *resources.EntityEntitlement) error {
	if len(value.EntitlementType) == 0 {
		return NewRequestBodyValidationError("empty entitlement type")
	}
	if len(value.EntityName) == 0 {
		return NewRequestBodyValidationError("empty entity name")
	}
	if len(value.EntityType) == 0 {
		return NewRequestBodyValidationError("empty entity type")
	}
	return nil
}

func validateGroupIdentitiesPatchRequestBody(value *resources.GroupIdentitiesPatchRequestBody) error {
	if len(value.Patches) == 0 {
		return NewRequestBodyValidationError("empty patch array")
	}
	return validateSlice[resources.GroupIdentitiesPatchItem](value.Patches, validateGroupIdentitiesPatchItem)
}

func validateGroupIdentitiesPatchItem(value *resources.GroupIdentitiesPatchItem) error {
	if len(value.Identity) == 0 {
		return NewRequestBodyValidationError("empty identity name")
	}
	return validateStringEnum("op", value.Op, resources.GroupIdentitiesPatchItemOpAdd, resources.GroupIdentitiesPatchItemOpRemove)
}

func validateGroupRolesPatchRequestBody(value *resources.GroupRolesPatchRequestBody) error {
	if len(value.Patches) == 0 {
		return NewRequestBodyValidationError("empty patch array")
	}
	return validateSlice[resources.GroupRolesPatchItem](value.Patches, validateGroupRolesPatchItem)
}

func validateGroupRolesPatchItem(value *resources.GroupRolesPatchItem) error {
	if len(value.Role) == 0 {
		return NewRequestBodyValidationError("empty role name")
	}
	return validateStringEnum("op", value.Op, resources.GroupRolesPatchItemOpAdd, resources.GroupRolesPatchItemOpRemove)
}
