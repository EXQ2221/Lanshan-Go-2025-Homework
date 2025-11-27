package pool

import (
	"search/seekfile"
	"sync"
)

type SearchTask struct {
	Path    string
	Keyword string
}
type Pool struct {
	TaskChan   chan SearchTask
	ResultChan chan seekfile.SearchResult
	Wg         *sync.WaitGroup
}

func Workers(people int, corridor int) *Pool {
	p := &Pool{
		TaskChan:   make(chan SearchTask, corridor),
		Wg:         &sync.WaitGroup{},
		ResultChan: make(chan seekfile.SearchResult, 10),
	}

	for range people {
		go func() {
			for t := range p.TaskChan {
				result := seekfile.Searchfile(t.Path, t.Keyword)
				p.ResultChan <- result
				p.Wg.Done()
			}
		}()
	}
	return p
}

func (p *Pool) Submit(t SearchTask) {
	p.Wg.Add(1)
	p.TaskChan <- t

}
func (p *Pool) CloseTaskChan() {
	close(p.TaskChan)
}
func (p *Pool) Wait() {
	p.Wg.Wait()
	close(p.ResultChan)
}
