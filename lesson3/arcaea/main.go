package main

import (
	"arcaea/song"
	"fmt"
	"sort"
)

func update(songs *[]song.Song) {
	var index int
	fmt.Println("请选择你要更新的歌曲(输入歌曲前的数字)：")
	for {
		fmt.Scan(&index)
		if index >= 1 && index <= len(*songs) {
			break
		} else {
			fmt.Println("无效的编号，请重新输入：")
		}
	}
	var score float32
	var point float32
	for {
		fmt.Printf("当前歌曲%s 请输入你的分数(0-10002221)", (*songs)[index-1].Name)
		fmt.Scan(&score)
		if score > 10002221 || score < 0 {
			fmt.Println("输入错误，请重新输入")
		} else if score > 10000000 {
			point = ((*songs)[index-1].Difficulty) + 2
			(*songs)[index-1].PotentialPoint = point
			break
		} else if score >= 9800000 {
			point = ((*songs)[index-1].Difficulty) + 1 + (score-9800000)/200000
			(*songs)[index-1].PotentialPoint = point
			break
		} else if score < 9800000 {
			point = ((*songs)[index-1].Difficulty) + (score-9500000)/300000
			(*songs)[index-1].PotentialPoint = point
			break
		}
	}
	(*songs)[index-1].Score = score
	if (*songs)[index-1].PotentialPoint < 0 {
		(*songs)[index-1].PotentialPoint = 0
	}
	fmt.Printf("当前歌曲 %s 分数为 %.1f 单曲潜力值为 %.4f\n",
		(*songs)[index-1].Name,
		(*songs)[index-1].Score,
		(*songs)[index-1].PotentialPoint)

}

func delete(songs *[]song.Song) {
	var index int
	fmt.Println("请选择你要删除的歌曲(输入歌曲前的数字，输入0退出)：")
	for {
		fmt.Scan(&index)
		if index >= 1 && index <= len(*songs) {
			break
		} else if index == 0 {
			return
		} else {
			fmt.Println("无效的编号，请重新输入：")
		}
	}
	(*songs)[index-1].Score = 0
	(*songs)[index-1].PotentialPoint = 0
}
func average(songs []song.Song) {
	var sum, ave float32
	for _, song := range songs {
		sum += song.PotentialPoint
	}
	ave = sum / 5
	fmt.Printf(" 当前平均ptt(取最好的五个成绩)： %f\n", ave)
}

func best5(songs []song.Song) {
	sort.Slice(songs, func(i, j int) bool {
		return songs[i].PotentialPoint > songs[j].PotentialPoint
	})
	fmt.Printf("%-3s %-20s %-6s %-10s %-10s\n",
		"No", "歌曲名", "定数", "分数", "PTT")
	limit := 5
	if len(songs) > limit {
		limit = len(songs)
	}
	for i := 0; i < 5; i++ {
		v := songs[i]
		fmt.Printf("%-3d %-20s %-6.1f %-10.0f %-10.4f\n",
			i+1,
			v.Name,
			v.Difficulty,
			v.Score,
			v.PotentialPoint,
		)
	}

}
func main() {
	var a int
	songs := song.Getallsongs()
	fmt.Println("欢迎使用arcaea查分器\n当前歌曲列表为")
	for i, v := range song.Getallsongs() {
		fmt.Printf("%d. %s | 定数%.1f\n", i+1, v.Name, v.Difficulty)
	}
	for {
		fmt.Println("请选择你想使用的功能\n1.更新歌曲成绩\n2.删除歌曲成绩\n3.查询当前平均ptt\n4.查询best5分表\n5.退出")
		fmt.Scan(&a)
		switch a {
		case 1:
			update(&songs)
		case 2:
			delete(&songs)
		case 3:
			average(songs)
		case 4:
			best5(songs)
		case 5:
			fmt.Println("程序退出")
			return

		}
	}
}
