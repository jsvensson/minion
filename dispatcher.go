// Package minion is a worker/dispatcher package for distributing jobs across a number of workers. Jobs are created
// by implementing the Job interface.
//
// See https://github.com/jsvensson/minion for a complete example.
package minion

// Dispatcher contains a worker pool and dispatches jobs to available workers.
type Dispatcher struct {
	workerPool chan chan Job
	jobQueue   chan Job
	maxWorkers int
	quit       chan bool
}

// NewDispatcher creates a new work dispatcher with the provided number of workers and buffer size for the job queue.
func NewDispatcher(workers, queueSize int) (*Dispatcher, chan<- Job) {
	queue := make(chan Job, queueSize)

	dispatcher := &Dispatcher{
		workerPool: make(chan chan Job, workers),
		jobQueue:   queue,
		maxWorkers: workers,
		quit:       make(chan bool),
	}

	return dispatcher, queue
}

// Run starts the dispatcher.
func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.workerPool)
		worker.Start()
	}

	go d.dispatch()
}

// Stop stops the dispatcher. Any currently running workers will finish their current job.
func (d *Dispatcher) Stop() {
	go func() {
		d.quit <- true
	}()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			jobChan := <-d.workerPool
			jobChan <- job
		case <-d.quit:
			return
		}
	}
}
