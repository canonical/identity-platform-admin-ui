// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	pageQueryKey      = "page"
	nextTokenQueryKey = "nextToken"
)

// Next contains data used to retrieve the next available set of results
type Next struct {
	Page      *int    `json:"page,omitempty"`
	PageToken *string `json:"pageToken,omitempty"`
}

// PaginatedResponse contains info about a page of results and possibly the next page to retrieve
type PaginatedResponse[T any] struct {
	Meta ResponseMeta
	Next Next
	Data []T
}

// populateQuery populates query parameters for paginated responses based on the next page available
// it overrides existing page or nextToken query param, without changing existing parameters in the original query
func (p *PaginatedResponse[T]) populateQuery(q url.Values) {
	q.Del(pageQueryKey)
	q.Del(nextTokenQueryKey)

	key := p.nextPageKey()
	if key == "" {
		return
	}

	q.Set(key, p.nextPage())
}

func (p *PaginatedResponse[T]) nextPageKey() string {
	key := ""
	if p.hasPageNumber() {
		key = pageQueryKey
	} else if p.hasNextPageToken() {
		key = nextTokenQueryKey
	}

	return key
}

// nextPage returns a string representing the next page (token or integer) or an empty string if there's no next page
func (p *PaginatedResponse[T]) nextPage() string {
	nextPage := ""
	if p.hasPageNumber() {
		nextPage = strconv.Itoa(*p.Next.Page)
	} else if p.hasNextPageToken() {
		nextPage = *p.Next.PageToken
	}

	return nextPage
}

func (p *PaginatedResponse[T]) hasNextPageToken() bool {
	return p.Next.PageToken != nil
}

func (p *PaginatedResponse[T]) hasPageNumber() bool {
	return p.Next.Page != nil
}

// NewResponseLinks returns a resources.ResponseLinks object with the href to retreive the next set of results, if any
func NewResponseLinks[T any](u *url.URL, p *PaginatedResponse[T]) ResponseLinks {
	query := u.Query()

	p.populateQuery(query)

	ret := ResponseLinks{}
	if queryParams := strings.TrimSpace(query.Encode()); queryParams != "" {
		ret.Next.Href = fmt.Sprintf("%s?%s", u.Path, queryParams)
	}

	return ret
}
