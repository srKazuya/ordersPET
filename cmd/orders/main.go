package main

import "github.com/srKazuya/ordersPET/internal/config"

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	_ = cfg

	//log

	//stor

	//route

	//server
}
