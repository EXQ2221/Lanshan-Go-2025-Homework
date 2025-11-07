package main

import "fmt"

var x, y, z int

func cclt(x int, y int, z int) func() int {
	return func() int {
		switch z {
		case 1:
			return x + y
		case 2:
			return x - y
		case 3:
			return x * y
		case 4:
			return x / y
		default:
			return 0

		}

	}
}
func main() {
	fmt.Println("请输入第一个数：")
	fmt.Scan(&x)
	fmt.Println("请输入你想运算的符号(数字)\n1. +\n2. -\n3. *\n4. /")
	fmt.Scan(&z)
	fmt.Println("请输入第二个数")
	fmt.Scan(&y)
	reslut := cclt(x, y, z)
	fmt.Println("result=", reslut())
}
