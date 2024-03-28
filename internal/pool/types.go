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

func NewResult[T any](key uuid.UUID, value T) *Result[T] {
	r := new(Result[T])

	r.key = key
	r.Value = value

	return r
}

type Job struct {
	ID uuid.UUID

	command any
	results chan *Result[any]

	wg *sync.WaitGroup
}

func NewJob(command any, results chan *Result[any], wg *sync.WaitGroup) *Job {
	j := new(Job)

	j.ID = uuid.New()
	j.command = command
	j.results = results
	j.wg = wg

	return j
}
