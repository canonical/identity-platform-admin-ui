// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package authorization

import (
	"fmt"
)

const (
	MEMBER_RELATION   = "member"
	ASSIGNEE_RELATION = "assignee"
	CAN_VIEW_RELATION = "can_view"
)

func UserForTuple(userId string) string {
	return fmt.Sprintf("user:%s", userId)
}

func UserWildcardForTuple() string {
	return fmt.Sprintf("user:*")
}

func RoleForTuple(roleId string) string {
	return fmt.Sprintf("role:%s", roleId)
}

func RoleAssigneeForTuple(roleId string) string {
	return fmt.Sprintf("role:%s#%s", roleId, ASSIGNEE_RELATION)
}

func GroupForTuple(groupId string) string {
	return fmt.Sprintf("group:%s", groupId)
}

func GroupMemberForTuple(groupId string) string {
	return fmt.Sprintf("group:%s#%s", groupId, MEMBER_RELATION)
}

func IdentityForTuple(identityId string) string {
	return fmt.Sprintf("identity:%s", identityId)
}

func SchemeForTuple(schemeId string) string {
	return fmt.Sprintf("scheme:%s", schemeId)
}

func ClientForTuple(clientId string) string {
	return fmt.Sprintf("client:%s", clientId)
}

func ProviderForTuple(providerId string) string {
	return fmt.Sprintf("provider:%s", providerId)
}

func RuleForTuple(ruleId string) string {
	return fmt.Sprintf("rule:%s", ruleId)
}

func ApplicationForTuple(applicationId string) string {
	return fmt.Sprintf("application:%s", applicationId)
}
