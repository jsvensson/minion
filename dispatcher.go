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
func NewDispatcher(workers, queueSize int) *Dispatcher {
	return &Dispatcher{
		workerPool: make(chan chan Job, workers),
		jobQueue:   make(chan Job, queueSize),
		maxWorkers: workers,
		quit:       make(chan bool),
	}
}

// Run starts the dispatcher.
func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.workerPool)
		worker.Start()
	}

	go d.dispatch()
}

// Stop stops the dispatcher, preventing it from accepting new jobs. Any jobs currently in the job queue will continue
// to be dispatched to workers until the job queue is empty.
func (d *Dispatcher) Stop() {
	go func() {
		d.quit <- true
	}()
}

// Enqueue adds a job to the job queue. If the queue is full, the function will block until the queue has slots
// available.
func (d *Dispatcher) Enqueue(job Job) {
	d.jobQueue <- job
}

// TryEnqueue will try to enqueue a job, without blocking. It returns true if the job was enqueued, false if the job
// queue is full and the job was unable to be enqueued.
func (d *Dispatcher) TryEnqueue(job Job) bool {
	select {
	case d.jobQueue <- job:
		return true
	default:
		return false
	}
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
