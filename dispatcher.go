package minion

// Dispatcher contains a worker pool and dispatches jobs to available workers.
type Dispatcher struct {
	WorkerPool chan chan Job
	jobQueue   chan Job
	maxWorkers int
	quit       chan bool
}

// NewDispatcher creates a new work dispatcher with the provided number of workers and size for the job queue.
func NewDispatcher(workers, queueSize int) (*Dispatcher, chan<- Job) {
	queue := make(chan Job, queueSize)

	dispatcher := &Dispatcher{
		WorkerPool: make(chan chan Job, workers),
		jobQueue:   queue,
		maxWorkers: workers,
		quit:       make(chan bool),
	}

	return dispatcher, queue
}

// Run starts the dispatcher.
func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
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
			jobChan := <-d.WorkerPool
			jobChan <- job
		case <-d.quit:
			return
		}
	}
}
