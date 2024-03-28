// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"sync"
)

type WorkerPoolInterface interface {
	Submit(any, chan *Result[any], *sync.WaitGroup) (string, error)
}
