package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/handler"
	"github.com/nosikmy/avito-shop/internal/app/model"
	"github.com/nosikmy/avito-shop/internal/app/repository"
	"github.com/nosikmy/avito-shop/internal/app/server"
	"github.com/nosikmy/avito-shop/internal/app/service"
)

func main() {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	if err := godotenv.Load(); err != nil {
		log.Error("error loading .env file: " + err.Error())
		return
	}

	cfgDB := repository.Config{
		Host:     os.Getenv(model.EnvDatabaseHost),
		Port:     os.Getenv(model.EnvDatabasePort),
		Username: os.Getenv(model.EnvDatabaseUser),
		Password: os.Getenv(model.EnvDatabasePassword),
		DBName:   os.Getenv(model.EnvDatabaseName),
	}

	db, err := repository.NewPostgresDB(cfgDB)
	if err != nil {
		log.Error("error occurred while init DB: " + err.Error())
		return
	}

	authRepository := repository.NewAuthRepository(log, db)
	historyRepository := repository.NewHistoryRepository(log, db)
	infoRepository := repository.NewInfoRepository(log, db)
	shoppingRepository := repository.NewShoppingRepository(log, db)

	authService := service.NewAuthService(log, authRepository)
	shopService := service.NewShopService(log, infoRepository, historyRepository, shoppingRepository)

	handlers := handler.NewHandler(log, authService, shopService)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srvPort := os.Getenv(model.EnvServerPort)
	srv, err := server.NewServer(srvPort, handlers.InitRoutes())
	if err != nil {
		log.Error("error creating new server: " + err.Error())
		return
	}

	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("error while running server: " + err.Error())
		}
		log.Info("server shuts down")
		cancel()
	}()

	log.Info("server is running om port: " + srvPort)

	<-ctx.Done()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("can't terminate server: %s" + err.Error())
	}

	if err := db.Close(); err != nil {
		log.Error("can't close DB connection: %s" + err.Error())
	}
}
