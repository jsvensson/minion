# Minion

[![GoDoc](https://godoc.org/github.com/jsvensson/minion?status.svg)](https://godoc.org/github.com/jsvensson/minion)
[![Build Status](https://travis-ci.org/jsvensson/minion.svg?branch=master)](https://travis-ci.org/jsvensson/minion)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsvensson/minion)](https://goreportcard.com/report/github.com/jsvensson/minion)
[![codecov](https://codecov.io/gh/jsvensson/minion/branch/master/graph/badge.svg)](https://codecov.io/gh/jsvensson/minion)

**Minion** 🍌 is a worker/dispatcher package for distributing jobs across a number of workers. The dispatcher creates the workers and returns a buffered channel of a given length that is used as the job queue. The incoming jobs are distributed to the available workers.

A job is created by implementing the `Job` interface:

``` go
type Job interface {
	// Perform runs the job.
	Perform()
}
```

The implementing struct can contain whatever additional fields it needs to perform the job, as seen in the example below.

This package is strongly inspired by [Marcio Castilho's article](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/).

## Example

``` go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/jsvensson/minion"
)

func main() {
    // Create dispatcher/queue with five workers and a queue size of 10
    dispatcher, jobQueue := minion.NewDispatcher(5, 10)

    // Start the dispatcher
    dispatcher.Run()

    // Timer and waitgroup to sync job completion
    start := time.Now()
    wg := &sync.WaitGroup{}

    // Create 50 calculation jobs
    for i := 1; i <= 50; i++ {
        wg.Add(1)
        fmt.Println("Creating job", i)
        job := Calculation{i, wg}
        
        // Add job to queue, blocks if the job queue is full
        jobQueue <- job
    }

    // Wait for all jobs to finish
    wg.Wait()
    fmt.Printf("Jobs executed in %.3f seconds\n", time.Since(start).Seconds())
}

// Calculation implements the Job interface so it can be sent to the job queue.
type Calculation struct {
    Value int
    wg    *sync.WaitGroup
}

// Perform runs the calculation job.
func (c Calculation) Perform() {
    // Do some time-consuming math
    time.Sleep(1 * time.Second)
    result := c.Value * c.Value

    fmt.Printf("Calculation %d done, result: %d\n", c.Value, result)
    c.wg.Done()
}
```
