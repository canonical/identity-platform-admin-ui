// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package types

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tomnomnom/linkheader"
)

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Meta    *Pagination `json:"_meta,omitempty"`
}

// NavigationTokens are parameters used to navigate `list` result endpoints
type NavigationTokens struct {
	// serialization only
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// Pagination object is used to serialize and deserialize pagination parameters
// it will populate the `meta` part for the `Response` struct
type Pagination struct {
	PageToken string `json:"page_token,omitempty"` // deserialization only
	Size      int64  `json:"size"`

	// serialization only
	NavigationTokens
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

// ParseLinkTokens accepts a request/response headers and will parse the Link
// headers, it returns quickly in case of error with a default NavigationTokens object
func ParseLinkTokens(headers http.Header) (NavigationTokens, error) {
	links := linkheader.Parse(headers.Get("Link"))

	pagination := NavigationTokens{}

	for _, link := range links {
		token, err := parseLinkToken(link.URL)

		if err != nil {
			return NavigationTokens{}, err
		}

		switch link.Rel {
		case "next":
			pagination.Next = token
		case "prev":
			pagination.Prev = token
		}
	}

	return pagination, nil
}

func parseLinkToken(linkURL string) (string, error) {
	u, err := url.Parse(linkURL)

	if err != nil {
		return "", fmt.Errorf("failed to parse link header successfully: %s", err)
	}

	return u.Query().Get("page_token"), nil
}
