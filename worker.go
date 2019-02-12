package minion

import "gopkg.in/tomb.v2"

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
	tomb       *tomb.Tomb
}

// NewWorker creates a new worker connected to the provided worker pool.
func NewWorker(pool chan chan Job, tomb *tomb.Tomb) Worker {
	return Worker{
		workerPool: pool,
		jobChannel: make(chan Job),
		tomb:       tomb,
	}
}

// Run starts the worker, waiting for incoming jobs on its job channel.
func (w Worker) Run() error {
	for {
		// Register worker in pool
		w.workerPool <- w.jobChannel

		select {
		case job := <-w.jobChannel:
			job.Perform()
		case <-w.tomb.Dying():
			return tomb.ErrDying
		}
	}
}
