package pool

import "sync"

type Task struct {
	Run func()
}
type Pool struct {
	ch chan Task
	Wg *sync.WaitGroup
}

var lock = sync.Mutex{}

func Workers(people int, corridor int) *Pool {
	p := &Pool{
		ch: make(chan Task, corridor),
		Wg: &sync.WaitGroup{},
	}

	for range people {
		go func() {
			for t := range p.ch {
				t.Run()
			}
		}()
	}
	return p
}

func (p *Pool) Submit(t Task) {
	p.Wg.Add(1)
	p.ch <- t

}
func (p *Pool) Wait() {
	p.Wg.Wait()
}
func (p *Pool) Close() {
	close(p.ch)
}
