// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0-only

package pool

import (
	"sync"
)

type WorkerPoolInterface interface {
	Submit(any, chan *Result[any], *sync.WaitGroup) (string, error)
}
