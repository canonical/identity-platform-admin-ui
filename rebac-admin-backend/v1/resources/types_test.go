// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPaginatedResponse_PopulateQuery(t *testing.T) {
	c := qt.New(t)

	mockNextPageNumber := 42
	mockNextPageToken := "mock-next-page-token"
	mockURL, err := url.Parse("http://localhost/mock-endpoint?filter=mock-filter")
	c.Assert(err, qt.IsNil)

	type queryTest struct {
		title string
		key   string
		value string
		p     PaginatedResponse[string]
	}

	for _, test := range []queryTest{
		{
			title: "PageNumber",
			key:   "page",
			value: "42",
			p: PaginatedResponse[string]{
				Next: Next{Page: &mockNextPageNumber},
			},
		},
		{
			title: "PageToken",
			key:   "nextToken",
			value: "mock-next-page-token",
			p: PaginatedResponse[string]{
				Next: Next{PageToken: &mockNextPageToken},
			},
		},
		{
			title: "Empty",
			p:     PaginatedResponse[string]{},
		},
	} {
		tt := test
		c.Run(fmt.Sprintf("TestPaginatedResponse_PopulateQuery%s", tt.title), func(c *qt.C) {
			expectedQuery := url.Values{}
			expectedQuery.Set("filter", "mock-filter")
			if tt.key != "" {
				expectedQuery.Set(tt.key, tt.value)
			}

			query := mockURL.Query()
			tt.p.populateQuery(&query)

			c.Assert(query, qt.DeepEquals, expectedQuery)
		})
	}
}

func TestNewResponseLinks(t *testing.T) {
	c := qt.New(t)

	type responseLinksTest struct {
		title    string
		expected ResponseLinks
		url      *url.URL
		p        *PaginatedResponse[string]
	}

	getURL := func(rawURL string) *url.URL {
		u, err := url.Parse(rawURL)
		c.Assert(err, qt.IsNil)
		return u
	}

	mockPageNumber := 41
	mockNextPageNumber := 42
	mockPageToken := "mock-token"
	mockNextPageToken := "mock-next-token"

	for _, test := range []responseLinksTest{
		{
			title:    "NextPageNumber",
			expected: ResponseLinks{ResponseLinksNext{Href: "/endpoint/mock?page=42"}},
			url:      getURL("https://localhost:8080/endpoint/mock?page=41"),
			p: &PaginatedResponse[string]{
				Meta: ResponseMeta{
					Page: &mockPageNumber,
				},
				Next: Next{Page: &mockNextPageNumber},
			},
		},
		{
			title:    "NextPageToken",
			expected: ResponseLinks{ResponseLinksNext{Href: "/endpoint/test?nextToken=mock-next-token"}},
			url:      getURL("https://localhost:8080/endpoint/test?nextToken=mock-token"),
			p: &PaginatedResponse[string]{
				Meta: ResponseMeta{
					PageToken: &mockPageToken,
				},
				Next: Next{PageToken: &mockNextPageToken},
			},
		},
		{
			title:    "NoNextPageNumber",
			expected: ResponseLinks{},
			url:      getURL("https://localhost:8080/endpoint/test?page=42"),
			p: &PaginatedResponse[string]{
				Meta: ResponseMeta{
					Page: &mockPageNumber,
				},
			},
		},
		{
			title:    "NoNextPageToken",
			expected: ResponseLinks{},
			url:      getURL("https://localhost:8080/endpoint/test?nextToken=mock-token"),
			p: &PaginatedResponse[string]{
				Meta: ResponseMeta{
					PageToken: &mockPageToken,
				},
			},
		},
	} {
		tt := test
		c.Run(tt.title, func(c *qt.C) {
			response := NewResponseLinks(tt.url, tt.p)

			c.Assert(response, qt.DeepEquals, tt.expected)
		})
	}
}
