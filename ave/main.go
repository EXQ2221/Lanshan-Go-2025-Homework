package main

import "fmt"

func average(sum, count int) int {
	return sum / count
}
func main() {
	var count = 0
	var i int
	var sum = 0
	fmt.Println("请输入一个整数（输入0结束）：")
	for {
		fmt.Scan(&i)
		if i == 0 {
			break
		} else {
			sum = sum + i
			count++
		}
	}
	if count == 0 {
		fmt.Println("只输入了0一个数字")
		return
	}
	if ave := average(sum, count); ave > 100 {
		fmt.Printf("考%d分何意味:", ave)
	} else if ave > 60 {
		fmt.Println("成绩合格，你的得分为:", ave)
	} else {
		fmt.Println("成绩不合格，你的得分为:", ave)
	}

}
