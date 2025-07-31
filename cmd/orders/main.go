package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/srKazuya/ordersPET/internal/config"

	"github.com/srKazuya/ordersPET/internal/http-server/handlers/save"
	nwLogger "github.com/srKazuya/ordersPET/internal/http-server/middleware/nwLogger"
	kafka "github.com/srKazuya/ordersPET/internal/kafka"
	"github.com/srKazuya/ordersPET/internal/lib/logger/sl"
	"github.com/srKazuya/ordersPET/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var address []string

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	log.Info("init server", slog.String("address", cfg.Address))
	log.Debug("log debug mode enabl;ed")

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
	if err != nil {
		log.Error("failed to init stroage", sl.Err(err))
	}
	_ = storage

	for _, ad := range cfg.Kafka.Brokers {
		address = append(address, ad)
	}
	p, err := kafka.NewProducer(log, address)
	if err != nil {
		log.Error("failed to init producer", sl.Err(err))
	}

	// for i := 0; i < 100; i++ {
	// 	msg := fmt.Sprintf("ASdasda %d", i)
	// 	if err = p.Produce(msg, cfg.Kafka.Topic); err != nil {
	// 		log.Error("asdas")
	// 	}
	// }
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(nwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/save", save.New(log, p, cfg.Kafka.Topic))

	//route

	//server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log

}
