package main

import (
	"fmt"


	"github.com/srKazuya/ordersPET/internal/config"
	"github.com/srKazuya/ordersPET/internal/storage/postgres"
)

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
	pgConfig := postgres.Config{
		DSN: fmt.Sprintf("host=%s user=%s port=%s password=%s dbname=%s sslmode=%s",
			cfg.DataBase.Host,
			cfg.DataBase.User,
			cfg.DataBase.Port,
			cfg.DataBase.Password,
			cfg.DataBase.Dbname,
			cfg.DataBase.Sslmode,
		),
	}

	storage, err := postgres.New(pgConfig)
	if err !=nil{
		//log errror
	}
	_ = storage

	//route

	//server
}
