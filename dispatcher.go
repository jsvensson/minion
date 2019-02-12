// Package minion is a worker/dispatcher package for distributing jobs across a number of workers. Jobs are created
// by implementing the Job interface.
//
// See https://github.com/jsvensson/minion for a complete example.
package minion

import (
	"gopkg.in/tomb.v2"
)

// Dispatcher contains a worker pool and dispatches jobs to available workers.
type Dispatcher struct {
	workerPool chan chan interface{}
	jobQueue   chan interface{}
	maxWorkers int
	handler    func(interface{})
	tomb       *tomb.Tomb
}

// NewDispatcher creates a new work dispatcher with the provided number of workers and buffer size for the job queue.
// The handler function receives the job data and processes it.
func NewDispatcher(handler func(interface{}), workers, queueSize int) *Dispatcher {
	return &Dispatcher{
		workerPool: make(chan chan interface{}, workers),
		jobQueue:   make(chan interface{}, queueSize),
		maxWorkers: workers,
		handler:    handler,
		tomb:       &tomb.Tomb{},
	}
}

// Start starts the dispatcher. This function does not block.
func (d *Dispatcher) Start() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.handler, d.workerPool, d.tomb)
		d.tomb.Go(worker.Run)
	}

	d.tomb.Go(d.dispatch)
}

// Stop stops the dispatcher, preventing it from accepting new jobs. Any jobs currently in the job queue will continue
// to be dispatched to workers until the job queue is empty.
func (d *Dispatcher) Stop() error {
	d.tomb.Kill(nil)
	close(d.jobQueue)
	return d.tomb.Wait()
}

// Enqueue adds a job to the job queue. If the queue is full, the function will block until the queue has slots
// available. Panics if the dispatcher has been stopped.
func (d *Dispatcher) Enqueue(job interface{}) {
	d.jobQueue <- job
}

// TryEnqueue will try to enqueue a job, without blocking. It returns true if the job was enqueued, false otherwise.
func (d *Dispatcher) TryEnqueue(job interface{}) bool {
	select {
	case d.jobQueue <- job:
		return true
	default:
		return false
	}
}

func (d *Dispatcher) dispatch() error {
	for {
		select {
		case job := <-d.jobQueue:
			<-d.workerPool <- job
		case <-d.tomb.Dying():
			return tomb.ErrDying
		}
	}
}
