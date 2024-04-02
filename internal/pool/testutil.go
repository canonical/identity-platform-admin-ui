// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"sync"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func SetupMockSubmit(wp *MockWorkerPoolInterface, resultsChan chan *Result[any]) (*gomock.Call, chan *Result[any]) {
	key := uuid.New()
	var internalResultsChannel chan *Result[any]

	call := wp.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Do(
		func(command any, results chan *Result[any], wg *sync.WaitGroup) {
			var value any = true

			switch commandFunc := command.(type) {
			case func():
				commandFunc()
			case func() any:
				value = commandFunc()
			}

			result := NewResult[any](key, value)
			results <- result
			if resultsChan != nil {
				resultsChan <- result
			}

			wg.Done()

			internalResultsChannel = results
		},
	).Return(key.String(), nil)

	return call, internalResultsChannel
}
