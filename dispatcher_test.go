package minion_test

import (
	"sync"
	"testing"

	"sync/atomic"

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

func (t *MinionTestSuite) TestDispatcherSingleJob() {
	disp, jobChan := minion.NewDispatcher(1, 1)
	disp.Run()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	jobChan <- TestJob{wg, t.counter}
	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(1), actual)
}

func (t *MinionTestSuite) TestDispatcherManyJobs() {
	disp, jobChan := minion.NewDispatcher(3, 10)
	disp.Run()

	wg := &sync.WaitGroup{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		jobChan <- TestJob{wg, t.counter}
	}

	wg.Wait()

	actual := atomic.LoadUint64(t.counter)
	assert.Equal(t.T(), uint64(50), actual)
}
