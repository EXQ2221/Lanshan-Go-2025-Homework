package main

import (
	"fmt"
	"lesson6/api"
	"lesson6/dao"
)

func main() {
	err := dao.LoadDB()
	if err != nil {
		fmt.Println("fail to load data")
	}

	err = dao.LoadRefreshToken()
	if err != nil {
		fmt.Println("fail to load refresh token")
	}
	r := api.InitRouter()
	r.Run(":8080")
}
