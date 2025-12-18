package main

import (
	"lesson8/api"
	"lesson8/dao"
)

func main() {
	dao.InitDB()
	dao.InitRedis()
	api.InitRouter()
}
