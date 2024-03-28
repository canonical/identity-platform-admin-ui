// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL

package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type WorkerPool struct {
	workers int

	jobs chan *Job

	shutdownCtx  context.Context
	shutdownFunc context.CancelCauseFunc

	wg sync.WaitGroup
}

func (p *WorkerPool) Stop() {
	p.shutdownFunc(fmt.Errorf("shutting down"))
	p.wg.Wait()
}

func (p *WorkerPool) Submit(command any, results chan *Result[any], wg *sync.WaitGroup) error {
	select {
	case p.jobs <- NewJob(command, results, wg):
		return nil
	default:
		return fmt.Errorf("WorkerPool queue is full")
	}
}

func (p *WorkerPool) consume(ID uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in consume ", ID.String(), " ", r)
			p.wg.Done()
			// TODO @shipperizer start another worker
			return
		}
	}()

	for {
		select {
		case <-p.shutdownCtx.Done():
			fmt.Println(ID, " going down")
			p.wg.Done()
			return
		case job := <-p.jobs:
			p.execute(job.ID, job.command, job.results, job.wg)
		}

	}

}

func (p *WorkerPool) execute(jobID uuid.UUID, command any, results chan *Result[any], wg *sync.WaitGroup) {

	defer wg.Done()

	select {
	case <-p.shutdownCtx.Done():
		fmt.Println("aborting")
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

// TODO @shipperizer add logger, monitor and tracer
func NewWorkerPool(workers int) *WorkerPool {
	p := new(WorkerPool)

	p.workers = workers

	p.shutdownCtx, p.shutdownFunc = context.WithCancelCause(context.Background())
	p.jobs = make(chan *Job)

	go p.start()

	return p
}
