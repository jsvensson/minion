package minion_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"time"

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
	time.Sleep(20 * time.Millisecond)
	tj.wg.Done()
}

func TestMinionTestSuite(t *testing.T) {
	suite.Run(t, new(MinionTestSuite))
}

func (t *MinionTestSuite) SetupTest() {
	t.counter = new(uint64)
}

func (t *MinionTestSuite) TestEnqueueSingleJob() {
	disp := minion.NewDispatcher(1, 1)
	disp.Run()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	disp.Enqueue(TestJob{wg, t.counter})
	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}

func (t *MinionTestSuite) TestEnqueueManyJobs() {
	disp := minion.NewDispatcher(3, 10)
	disp.Run()

	wg := &sync.WaitGroup{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		disp.Enqueue(TestJob{wg, t.counter})
	}

	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(50), actual)
}

func (t *MinionTestSuite) TestTryEnqueueSingleJob() {
	disp := minion.NewDispatcher(1, 1)
	disp.Run()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	disp.TryEnqueue(TestJob{wg, t.counter})
	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}

func (t *MinionTestSuite) TestTryEnqueueBlockedJob() {
	disp := minion.NewDispatcher(1, 1)

	wg := &sync.WaitGroup{}

	// Fist call gets enqueued
	wg.Add(1)
	enqueued := disp.TryEnqueue(TestJob{wg, t.counter})
	assert.True(t.T(), enqueued, "first job should not be blocked")

	// Second call never gets enqueued
	enqueued = disp.TryEnqueue(TestJob{wg, t.counter})
	assert.False(t.T(), enqueued, "second job should be blocked")

	disp.Run()

	wg.Wait()
	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}
