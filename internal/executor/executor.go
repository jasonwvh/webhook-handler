package executor

import (
	"sync"
)

type Executor struct {
	wg     sync.WaitGroup
	jobCh  chan func()
	stopCh chan struct{}
}

func NewExecutor(workers int) *Executor {
	e := &Executor{
		jobCh:  make(chan func(), 1000),
		stopCh: make(chan struct{}),
	}

	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.worker()
	}

	return e
}

func (e *Executor) Submit(job func()) {
	select {
	case e.jobCh <- job:
		return
	case <-e.stopCh:
		return
	}
}

func (e *Executor) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

func (e *Executor) worker() {
	defer e.wg.Done()

	for {
		select {
		case job := <-e.jobCh:
			job()
		case <-e.stopCh:
			return
		}
	}
}