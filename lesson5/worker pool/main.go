package main

import (
	"fmt"
	"projectsekai_calc/pool"
	"sync"
)

func main() {
	p := pool.Workers(10, 10)
	var a int
	var mu sync.Mutex
	for range 1000 {
		p.Submit(pool.Task{
			Run: func() {
				defer p.Wg.Done()
				mu.Lock()
				a++
				mu.Unlock()
			},
		})
	}
	p.Wait()
	p.Close()

	fmt.Println(a)
}
