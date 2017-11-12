package minion

// Job describes a job to perform. The implementing struct can contain whatever additional fields
// it requires to perform its job when Perform() runs.
type Job interface {
	// Perform runs the job.
	Perform()
}

// Worker waits for incoming jobs on its channel and performs them.
type Worker struct {
	workerPool chan chan Job
	jobChannel chan Job
	quit       chan bool
}

// NewWorker creates a new worker connected to the provided worker pool.
func NewWorker(pool chan chan Job) Worker {
	return Worker{
		workerPool: pool,
		jobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

// Start starts the worker, waiting for incoming jobs on its job channel.
func (w Worker) Start() {
	go func() {
		for {
			// Register worker in pool
			w.workerPool <- w.jobChannel

			select {
			case job := <-w.jobChannel:
				job.Perform()
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop stops the worker.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
