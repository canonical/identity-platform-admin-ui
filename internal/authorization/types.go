// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package authorization

import (
	"fmt"
	"strings"
)

const (
	PERMISSION_SEPARATOR = "::"
)

type URN struct {
	relation string
	object   string
}

func (a *URN) ID() string {
	return fmt.Sprintf("%s%s%s", a.relation, PERMISSION_SEPARATOR, a.object)
}

func (a *URN) Relation() string {
	return a.relation
}

func (a *URN) Object() string {
	return a.object
}

func NewURN(relation, object string) *URN {
	u := new(URN)

	u.relation = relation
	u.object = object

	return u
}

func NewURNFromURLParam(ID string) *URN {
	values := strings.Split(ID, PERMISSION_SEPARATOR)

	if len(values) < 2 {
		// not a valid URN
		return nil
	}

	// use only first two elements
	return NewURN(values[0], values[1])
}
