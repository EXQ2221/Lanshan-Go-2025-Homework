package main

import "fmt"

var x, y int

func Freq(num []int) map[int]int {
	frequency := make(map[int]int)
	for _, v := range num {
		frequency[v]++
	}
	return frequency
}

func main() {
	num := []int{1, 2, 3, 4, 5, 2, 3, 4, 5, 3, 4, 1, 3}
	result := Freq(num)
	for k, v := range result {
		fmt.Printf("数字%d出现了%d次\n", k, v)
	}

}
