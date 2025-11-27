package main

import "fmt"

func main() {
	var i int64
	var sum int64 = 1
	fmt.Scanln(&i)
	for ; i > 0; i-- {
		sum = sum * i
	}
	fmt.Println(sum)
}
