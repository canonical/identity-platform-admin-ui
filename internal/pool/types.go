// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package pool

import (
	"sync"

	"github.com/google/uuid"
)

type Result[T any] struct {
	key   uuid.UUID
	Value T
}

func (r *Result[T]) ID() string {
	return r.key.String()
}

func NewResult[T any](key uuid.UUID, value T) *Result[T] {
	r := new(Result[T])

	r.key = key
	r.Value = value

	return r
}

type job struct {
	id uuid.UUID

	command any
	results chan *Result[any]

	wg *sync.WaitGroup
}

func (j *job) ID() string {
	return j.id.String()
}

func newJob(command any, results chan *Result[any], wg *sync.WaitGroup) *job {
	j := new(job)

	j.id = uuid.New()
	j.command = command
	j.results = results
	j.wg = wg

	return j
}
