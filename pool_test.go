package grpools

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	mapValues sync.Map
)

func TestPoolShouldIncrementGettingFromFor(t *testing.T) {
	setup()
	const expectedResult = 100000
	var count int64
	f := func(i int) {
		checkIfValueWasProcessed(i, t)
		atomic.AddInt64(&count, 1)
	}
	pool := NewPool(4)
	for i := 0; i < expectedResult; i++ {
		copyValToFunc := i
		pool.CallWorker(func() {
			f(copyValToFunc)
		})
	}
	pool.WaitWorkdersAndClose()
	if count != expectedResult {
		t.Fatalf("Expected %d, but receive %d", expectedResult, count)
	}
}

func TestPoolShouldIncrementGettingFromChannel(t *testing.T) {
	setup()
	const expectedResult = 100000
	var count int64
	chValues := make(chan int)
	chFinished := make(chan int)
	go func() {
		for i := 0; i < expectedResult; i++ {
			chValues <- i
		}
		close(chValues)
		chFinished <- 1
	}()
	f := func(i int) {
		checkIfValueWasProcessed(i, t)
		atomic.AddInt64(&count, 1)
	}
	pool := NewPool(4)
	for val := range chValues {
		copyValToFunc := val
		pool.CallWorker(func() {
			f(copyValToFunc)
		})
	}
	pool.WaitWorkdersAndClose()
	<-chFinished
	if count != expectedResult {
		t.Fatalf("Expected %d, but receive %d", expectedResult, count)
	}
}

func TestTwoPoolsShouldIncrementAndDecrement(t *testing.T) {
	setup()
	const expectedResult, incrementUntil = 0, 100000
	var count int64
	chDataBetweenPools := make(chan int)
	poolDecrement := NewPool(4)
	poolIncrement := NewPool(4)

	workerIncrementAndSendNegative := func(i int) {
		checkIfValueWasProcessed(i, t)
		atomic.AddInt64(&count, 1)
		chDataBetweenPools <- -i
	}
	workerReceiveAndDecrement := func() {
		for i := range chDataBetweenPools {
			checkIfValueWasProcessed(i, t)
			atomic.AddInt64(&count, -1)
		}
	}
	poolDecrement.CallWorkersUntilFill(workerReceiveAndDecrement)
	for i := 1; i <= incrementUntil; i++ {
		copyValToFunc := i
		poolIncrement.CallWorker(func() {
			workerIncrementAndSendNegative(copyValToFunc)
		})
	}
	poolIncrement.WaitWorkdersAndClose()
	close(chDataBetweenPools)
	poolDecrement.WaitWorkdersAndClose()
	if count != expectedResult {
		t.Fatalf("Expected %d, but receive %d", expectedResult, count)
	}
}

func TestMustIncrementInBackgroundFunc(t *testing.T) {
	setup()
	const expectedResult = 100000
	var count int64
	chFinished := CallBackgroundWorker(func() {
		for i := 0; i < expectedResult; i++ {
			atomic.AddInt64(&count, 1)
		}
	})
	<-chFinished
	if count != expectedResult {
		t.Fatalf("Expected %d, but receive %d", expectedResult, count)
	}
}

func TestCallMustBeUntilLimit(t *testing.T) {
	setup()
	var size int64 = 4
	var zero int64 = 0
	pool := NewPool(int(size))
	var counter int64 = 0
	funcao := func() {
		time.Sleep(1 * time.Second)
		atomic.AddInt64(&counter, 1)
	}
	if zero != pool.GetAmountWorkersProcessing() {
		t.Fatalf("Expected %d, but received %d", zero, pool.GetAmountWorkersProcessing())
	}
	pool.CallWorker(funcao)
	pool.CallWorkersUntilFill(funcao)
	if size != pool.GetAmountWorkersProcessing() {
		t.Fatalf("Expected %d, but received %d", zero, pool.GetAmountWorkersProcessing())
	}
	pool.WaitWorkdersAndClose()
	if counter != size {
		t.Fatalf("Expected %d, but received %d", size, counter)
	}
}

func checkIfValueWasProcessed(key int, t *testing.T) {
	if _, loaded := mapValues.LoadOrStore(key, 1); loaded && !t.Failed() {
		t.Errorf("Value %d already processed", key)
		t.Fail()
	}
}

func setup() {
	mapValues = sync.Map{}
}
