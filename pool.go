package pools

import (
	"sync"
)

// NewPool returns a Pool
func NewPool(qtdeGoroutines int) *Pool {
	chWorkers, chFinished := goroutinesPool(qtdeGoroutines)
	return &Pool{size: qtdeGoroutines, chWorkers: chWorkers, chFinished: chFinished}
}

type Pool struct {
	size       int
	chWorkers  chan<- func()
	chFinished <-chan int
}

func (p *Pool) CallWorker(worker func()) {
	p.chWorkers <- worker
}

func (p *Pool) CallWorkersUntilFill(worker func()) {
	for i := 0; i < p.size; i++ {
		p.chWorkers <- worker
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

func goroutinesPool(qtdeGoroutines int) (chan<- func(), <-chan int) {
	chWorkers := make(chan func())
	chFinished := make(chan int)
	var wg sync.WaitGroup
	wg.Add(qtdeGoroutines)
	for i := 0; i < qtdeGoroutines; i++ {
		go func() {
			for work := range chWorkers {
				work()
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		chFinished <- 1
	}()
	return chWorkers, chFinished
}
