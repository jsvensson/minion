package minion

import "gopkg.in/tomb.v2"

// Worker waits for incoming jobs on its channel and performs them.
type Worker struct {
	workerPool chan chan interface{}
	jobChannel chan interface{}
	handler    func(interface{})
	tomb       *tomb.Tomb
}

// NewWorker creates a new worker connected to the provided worker pool.
func NewWorker(handler func(interface{}), pool chan chan interface{}, tomb *tomb.Tomb) Worker {
	return Worker{
		workerPool: pool,
		jobChannel: make(chan interface{}),
		handler:    handler,
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
			w.handler(job)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		}
	}
}
