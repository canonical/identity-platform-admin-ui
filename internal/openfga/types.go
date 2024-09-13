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

type TokenMapFilter struct {
	tokens map[string]string
}

func (f *TokenMapFilter) WithFilter() any {
	return f.tokens
}

func NewTokenMapFilter(tokens map[string]string) *TokenMapFilter {
	f := new(TokenMapFilter)
	f.tokens = tokens

	return f
}

type TypesFilter struct {
	resourceTypes []string
}

func (f *TypesFilter) WithFilter() any {
	return f.resourceTypes
}

func NewTypesFilter(resourceTypes ...string) *TypesFilter {
	f := new(TypesFilter)

	f.resourceTypes = make([]string, 0)

	for _, r := range resourceTypes {
		f.resourceTypes = append(f.resourceTypes, r)
	}

	return f
}

type RelationFilter struct {
	relation string
}

func (f *RelationFilter) WithFilter() any {
	return f.relation
}

func NewRelationFilter(relation string) *RelationFilter {
	f := new(RelationFilter)

	f.relation = relation

	return f
}

type listPermissionsOpts struct {
	TokenMap       map[string]string
	TypesFilter    []string
	RelationFilter string
}
