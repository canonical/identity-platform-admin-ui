// Copyright 2024 Canonical Ltd
// SPDX-License-Identifier: AGPL

package types

import (
	"net/url"
	"strconv"
)

type Response struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Meta    *Pagination `json:"_meta"`
}

type Pagination struct {
	Page      int64  `json:"page"`
	PageToken string `json:"page_token"`
	Size      int64  `json:"size"`

	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
	First string `json:"first,omitempty"`
}

func NewPaginationWithDefaults() *Pagination {
	p := new(Pagination)

	p.PageToken = ""
	p.Size = 100

	return p
}

func ParsePagination(q url.Values) *Pagination {

	p := NewPaginationWithDefaults()

	// TODO @shipperizer introduce go-playground/validator
	if size, err := strconv.ParseInt(q.Get("size"), 10, 64); err == nil && size > 0 {
		p.Size = size
	}

	if token := q.Get("page_token"); token != "" {
		p.PageToken = token
	}

	return p
}
