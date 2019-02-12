package minion_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jsvensson/minion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MinionTestSuite struct {
	suite.Suite
	counter *uint64
}

type TestJob struct {
	wg      *sync.WaitGroup
	counter *uint64
}

// Perform increments the test counter.
func (tj TestJob) Perform() {
	atomic.AddUint64(tj.counter, 1)
	tj.wg.Done()
}

func TestMinionTestSuite(t *testing.T) {
	suite.Run(t, new(MinionTestSuite))
}

func (t *MinionTestSuite) SetupTest() {
	t.counter = new(uint64)
}

func (t *MinionTestSuite) TestEnqueueSingleJob() {
	d := minion.NewDispatcher(1, 1)
	d.Start()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	d.Enqueue(TestJob{wg, t.counter})
	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}

func (t *MinionTestSuite) TestEnqueueManyJobs() {
	d := minion.NewDispatcher(3, 10)
	d.Start()

	wg := &sync.WaitGroup{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		d.Enqueue(TestJob{wg, t.counter})
	}

	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(50), actual)
}

func (t *MinionTestSuite) TestTryEnqueueSingleJob() {
	d := minion.NewDispatcher(1, 1)
	d.Start()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	d.TryEnqueue(TestJob{wg, t.counter})
	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}

func (t *MinionTestSuite) TestTryEnqueueBlockedJob() {
	d := minion.NewDispatcher(1, 1)

	wg := &sync.WaitGroup{}

	// First call gets enqueued
	wg.Add(1)
	enqueued := d.TryEnqueue(TestJob{wg, t.counter})
	assert.True(t.T(), enqueued, "first job should not be blocked")

	// Second call never gets enqueued
	enqueued = d.TryEnqueue(TestJob{wg, t.counter})
	assert.False(t.T(), enqueued, "second job should be blocked")

	d.Start()

	wg.Wait()
	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}
