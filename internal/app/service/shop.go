package service

import (
	"fmt"
	"log/slog"

	"github.com/nosikmy/avito-shop/internal/app/model"
)

type InfoRepository interface {
	GetCoinsAmount(username string) (int, error)
	GetInventory(username string) ([]model.Item, error)
}

type HistoryRepository interface {
	GetCoinReceivedHistory(username string) ([]model.Receive, error)
	GetCoinSentHistory(username string) ([]model.Send, error)
}

type ShoppingRepository interface {
	SendCoin(fromUsername, toUsername string, amount int) error
	Buy(username, item string) error
}

type ShopService struct {
	logger             *slog.Logger
	infoRepository     InfoRepository
	historyRepository  HistoryRepository
	shoppingRepository ShoppingRepository
}

func NewShopService(logger *slog.Logger, i InfoRepository, h HistoryRepository, s ShoppingRepository) *ShopService {
	return &ShopService{
		logger:             logger,
		infoRepository:     i,
		historyRepository:  h,
		shoppingRepository: s,
	}
}

func (s *ShopService) GetInfo(username string) (model.InfoOutput, error) {
	const op = "service.shop.GetInfo"

	coins, err := s.infoRepository.GetCoinsAmount(username)
	if err != nil {
		return model.InfoOutput{}, fmt.Errorf("%s: %w", op, err)
	}

	inventory, err := s.infoRepository.GetInventory(username)
	if err != nil {
		return model.InfoOutput{}, fmt.Errorf("%s: %w", op, err)
	}

	received, err := s.historyRepository.GetCoinReceivedHistory(username)
	if err != nil {
		return model.InfoOutput{}, fmt.Errorf("%s: %w", op, err)
	}

	sent, err := s.historyRepository.GetCoinSentHistory(username)
	if err != nil {
		return model.InfoOutput{}, fmt.Errorf("%s: %w", op, err)
	}

	info := model.InfoOutput{
		Coins:     coins,
		Inventory: inventory,
		CoinHistory: model.CoinHistory{
			Received: received,
			Sent:     sent,
		},
	}

	return info, nil
}

func (s *ShopService) SendCoin(username string, send model.Send) error {
	const op = "service.shop.SendCoin"

	if err := s.shoppingRepository.SendCoin(username, send.ToUser, send.Amount); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *ShopService) Buy(username, item string) error {
	const op = "service.shop.Buy"

	if err := s.shoppingRepository.Buy(username, item); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
