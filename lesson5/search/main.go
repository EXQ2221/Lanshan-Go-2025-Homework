package main

import (
	"fmt"
	"os"
	"search/pool"
	"search/walkdir"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Println("当前工作目录 =", wd)
	var root, keyword string
	fmt.Print("请输入目录: ")
	fmt.Scan(&root)
	fmt.Print("请输入关键字: ")
	fmt.Scan(&keyword)

	paths, err := walkdir.WalkDir(root)
	if err != nil {
		fmt.Println("扫描目录失败:", err)
		return
	}

	p := pool.Workers(10, 20)

	go func() {
		for result := range p.ResultChan {
			if len(result.Line) == 0 {
				continue
			}

			fmt.Println("文件:", result.Path)
			for _, line := range result.Line {
				fmt.Println(" 匹配内容:  ", line)
			}
			fmt.Println()
		}
	}()

	for _, path := range paths {
		p.Submit(pool.SearchTask{
			Path:    path,
			Keyword: keyword,
		})
	}

	p.CloseTaskChan()

	p.Wait()

}
