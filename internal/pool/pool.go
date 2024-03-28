// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"github.com/canonical/identity-platform-admin-ui/internal/monitoring"
	"github.com/canonical/identity-platform-admin-ui/internal/tracing"
)

type WorkerPool struct {
	workers int

	jobs chan *job

	shutdownCtx  context.Context
	shutdownFunc context.CancelCauseFunc

	wg sync.WaitGroup

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (p *WorkerPool) Stop() {
	p.shutdownFunc(fmt.Errorf("shutting down"))
	p.wg.Wait()
}

func (p *WorkerPool) Submit(command any, results chan *Result[any], wg *sync.WaitGroup) (string, error) {
	_job := newJob(command, results, wg)
	select {
	case p.jobs <- _job:
		return _job.ID(), nil
	default:
		return "", fmt.Errorf("WorkerPool queue is full")
	}
}

func (p *WorkerPool) consume(ID uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Debug("Recovered in consume ", ID.String(), " ", r)
			p.wg.Done()
			// TODO @shipperizer start another worker
			return
		}
	}()

	for {
		select {
		case <-p.shutdownCtx.Done():
			p.logger.Info(ID, " going down")
			p.wg.Done()
			return
		case job := <-p.jobs:
			p.execute(job.id, job.command, job.results, job.wg)
		}

	}

}

func (p *WorkerPool) execute(jobID uuid.UUID, command any, results chan *Result[any], wg *sync.WaitGroup) {

	defer wg.Done()

	select {
	case <-p.shutdownCtx.Done():
		p.logger.Info(jobID, " aborting execution")
	default:
		switch commandFunc := command.(type) {
		case func():
			commandFunc()
			results <- NewResult[any](jobID, true)
		case func() any:
			results <- NewResult[any](jobID, commandFunc())
		}
	}
}
func (p *WorkerPool) start() {
	p.wg.Add(p.workers)

	for i := 0; i < p.workers; i++ {
		go p.consume(uuid.New())
	}
}

func NewWorkerPool(workers int, tracer tracing.TracingInterface, monitor monitoring.MonitorInterface, logger logging.LoggerInterface) *WorkerPool {
	p := new(WorkerPool)
	p.logger = logger
	p.monitor = monitor
	p.tracer = tracer

	p.workers = workers

	p.shutdownCtx, p.shutdownFunc = context.WithCancelCause(context.Background())
	p.jobs = make(chan *job, 2*workers)

	go p.start()

	return p
}
