// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package pool -destination ./mock_logger.go -source=../logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package pool -destination ./mock_monitor.go -source=../monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package pool -destination ./mock_tracing.go go.opentelemetry.io/otel/trace Tracer

func TestWorkerPool_Submit(t *testing.T) {
	for _, tt := range []struct {
		name           string
		command        func(i int) any
		expectedValues []string
	}{
		{
			name: "Callable",
			command: func(i int) any {
				return func() any {
					return strconv.Itoa(i)
				}
			},
			expectedValues: []string{
				"0",
				"1",
				"2",
				"3",
			},
		},
		{
			name: "Runnable",
			command: func(i int) any {
				return func() {
					strconv.Itoa(i)
				}
			},
			expectedValues: []string{
				"true",
				"true",
				"true",
				"true",
			},
		}} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tracer := NewMockTracer(ctrl)
			monitor := NewMockMonitorInterface(ctrl)
			logger := NewMockLoggerInterface(ctrl)

			tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			logger.EXPECT().Info(gomock.Any()).AnyTimes()
			logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

			expectedResultsMap := make(map[string]string, 4)

			wpool := NewWorkerPool(
				4,
				tracer,
				monitor,
				logger,
			)
			time.Sleep(time.Millisecond * 100)
			var wg sync.WaitGroup

			iterations := 4
			results := make(chan *Result[any], iterations)

			for i := 0; i < iterations; i++ {
				i := i
				wg.Add(1)
				taskID, err := wpool.Submit(tt.command(i), results, &wg)
				if err != nil {
					t.Fatalf("Unable to submit task")
				}

				expectedResultsMap[taskID] = strconv.Itoa(i)
			}

			resultsMap := make(map[string]any, 4)

			select {
			case <-time.After(time.Millisecond * 500):
				t.Fatalf("Timeout occurred")
			default:
				wg.Wait()
				close(results)
				for result := range results {
					var value any
					switch result.Value.(type) {
					case string:
						value = result.Value.(string)
					case bool:
						value = "true"
					}

					resultsMap[result.ID()] = value
				}
			}

			resultsValues := make([]string, 0)
			for _, value := range resultsMap {
				resultsValues = append(resultsValues, value.(string))
			}

			sort.Strings(resultsValues)

			if !reflect.DeepEqual(tt.expectedValues, resultsValues) {
				t.Fatalf("Results maps don't match")
			}
		})
	}
}

func TestWorkerPool_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracer := NewMockTracer(ctrl)
	monitor := NewMockMonitorInterface(ctrl)
	logger := NewMockLoggerInterface(ctrl)

	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(context.TODO(), trace.SpanFromContext(context.TODO()))
	logger.EXPECT().Info(gomock.Any()).AnyTimes()
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	wpool := NewWorkerPool(
		1,
		tracer,
		monitor,
		logger,
	)

	time.Sleep(time.Millisecond * 100)

	select {
	case <-time.After(time.Millisecond * 200):
		t.Fatalf("Timeout occurred")
	default:
		wpool.Stop()
		wpool.wg.Wait()
	}

}
