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

type Urn struct {
	relation string
	object   string
}

func (a *Urn) ID() string {
	return fmt.Sprintf("%s%s%s", a.relation, PERMISSION_SEPARATOR, a.object)
}

func (a *Urn) Relation() string {
	return a.relation
}

func (a *Urn) Object() string {
	return a.object
}

func NewUrn(relation, object string) *Urn {
	u := new(Urn)

	u.relation = relation
	u.object = object

	return u
}

func NewUrnFromURLParam(ID string) *Urn {
	values := strings.Split(ID, PERMISSION_SEPARATOR)

	if len(values) < 2 {
		// not a valid Urn
		return nil
	}

	// use only first two elements
	return NewUrn(values[0], values[1])
}
