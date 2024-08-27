// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package openfga

type listPermissionsResult struct {
	permissions []Permission
	token       string
	ofgaType    string
	err         error
}

type Permission struct {
	Relation string `json:"relation" validate:"required"`
	Object   string `json:"object" validate:"required"`
}

// Tuple is simply a wrapper around openfga TupleKey
// reason to have it is to hide underlying library complexity
// in case we want to swap it
type Tuple struct {
	User     string
	Relation string
	Object   string
}

func (t *Tuple) Values() (string, string, string) {
	return t.User, t.Relation, t.Object
}

func NewTuple(user, relation, object string) *Tuple {
	t := new(Tuple)

	t.User = user
	t.Relation = relation
	t.Object = object

	return t
}
