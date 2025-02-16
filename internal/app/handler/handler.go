package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/nosikmy/avito-shop/internal/app/model"
)

type AuthService interface {
	Auth(input model.AuthInput) error
	GenerateToken(username string) (string, error)
	ParseToken(token string) (string, error)
}
type ShopService interface {
	GetInfo(username string) (model.InfoOutput, error)
	SendCoin(username string, send model.Send) error
	Buy(username, item string) error
}

type Handler struct {
	logger      *slog.Logger
	authService AuthService
	shopService ShopService
}

func NewHandler(logger *slog.Logger, a AuthService, s ShopService) *Handler {
	return &Handler{
		logger:      logger,
		authService: a,
		shopService: s,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	apiRouter := router.Group("/api")
	{
		apiRouter.GET("/info", h.UserIdentify, h.GetInfo)
		apiRouter.POST("/sendCoin", h.UserIdentify, h.SendCoin)
		apiRouter.GET("/buy/:item", h.UserIdentify, h.Buy)
		apiRouter.POST("/auth", h.Auth)
	}

	return router
}
