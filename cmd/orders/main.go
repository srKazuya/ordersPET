package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/srKazuya/ordersPET/internal/config"
	getter "github.com/srKazuya/ordersPET/internal/service/getter"
	saver "github.com/srKazuya/ordersPET/internal/service/saver"

	"github.com/srKazuya/ordersPET/internal/http-server/handlers/get"
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
	fmt.Println("loaded config OK")
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
		os.Exit(1)
	}

	for _, ad := range cfg.Kafka.Brokers {
		address = append(address, ad)
	}
	p, err := kafka.NewProducer(log, address)
	if err != nil {
		log.Error("failed to init producer", sl.Err(err))
	}

	saver := saver.New(log, storage)
	c, err := kafka.NewConsumer(saver, log, address, cfg.Topic, cfg.ConsumerGroup)
	if err != nil {
		log.Error("failed to conusme", sl.Err(err))
	}
	go func() {
		c.Start(log)
	}()

	getter := getter.New(log, storage)
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(nwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	router.Post("/save", save.New(log, p, cfg.Kafka.Topic))
	router.Get("/orders/{order_uid}", get.New(log, getter))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting HTTP server", slog.String("address", cfg.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", sl.Err(err))
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Info("shutting down server...")

	c.Stop()
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
