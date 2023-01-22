package grpools

import (
	"sync"
	"sync/atomic"
)

// NewPool returns a Pool
func NewPool(amountGoroutines int) *Pool {
	p := Pool{}
	p.initPool(amountGoroutines)
	return &p
}

type Pool struct {
	size              int
	chWorkers         chan<- func()
	chFinished        <-chan int
	workersProcessing int64
}

func (p *Pool) CallWorker(worker func()) {
	p.chWorkers <- worker
	atomic.AddInt64(&p.workersProcessing, 1)
}

func (p *Pool) CallWorkersUntilFill(worker func()) {
	for i := 0; i < p.size; i++ {
		if atomic.LoadInt64(&p.workersProcessing) >= int64(p.size) {
			break
		}
		p.chWorkers <- worker
		atomic.AddInt64(&p.workersProcessing, 1)
	}
}

func CallBackgroundWorker(worker func()) chan int {
	chFinished := make(chan int)
	go func() {
		worker()
		close(chFinished)
	}()
	return chFinished
}

func (p *Pool) WaitWorkdersAndClose() {
	close(p.chWorkers)
	<-p.chFinished
}

func (p *Pool) GetAmountWorkersProcessing() int64 {
	return atomic.LoadInt64(&p.workersProcessing)
}

func (p *Pool) initPool(size int) {
	chWorkers := make(chan func())
	chFinished := make(chan int)
	p.size = size
	p.workersProcessing = 0
	p.chWorkers = chWorkers
	p.chFinished = chFinished
	var wg sync.WaitGroup
	wg.Add(p.size)
	for i := 0; i < p.size; i++ {
		go func() {
			for work := range chWorkers {
				work()
				atomic.AddInt64(&p.workersProcessing, -1)
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		chFinished <- 1
	}()
}
