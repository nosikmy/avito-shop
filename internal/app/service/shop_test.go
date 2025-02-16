package service

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetCoinsAmount(username string) (int, error) {
	args := m.Called(username)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetInventory(username string) ([]model.Item, error) {
	args := m.Called(username)
	inventory, _ := args.Get(0).([]model.Item)
	return inventory, args.Error(1)
}

func (m *MockRepository) SendCoin(fromUsername, toUsername string, amount int) error {
	args := m.Called(fromUsername, toUsername, amount)
	return args.Error(0)
}

func (m *MockRepository) Buy(username, item string) error {
	args := m.Called(username, item)
	return args.Error(0)
}

func (m *MockRepository) GetCoinReceivedHistory(username string) ([]model.Receive, error) {
	args := m.Called(username)
	received, _ := args.Get(0).([]model.Receive)
	return received, args.Error(1)
}

func (m *MockRepository) GetCoinSentHistory(username string) ([]model.Send, error) {
	args := m.Called(username)
	sent, _ := args.Get(0).([]model.Send)
	return sent, args.Error(1)
}

func TestNewShopService(t *testing.T) {
	type inputArgs struct {
		logger             *slog.Logger
		infoRepository     InfoRepository
		historyRepository  HistoryRepository
		shoppingRepository ShoppingRepository
	}
	tests := []struct {
		name    string
		args    inputArgs
		wantErr *apierror.APIError
	}{
		{
			name: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewShopService(tt.args.logger, tt.args.infoRepository, tt.args.historyRepository, tt.args.shoppingRepository)
			assert.Equal(t, &ShopService{
				logger:             tt.args.logger,
				infoRepository:     tt.args.infoRepository,
				historyRepository:  tt.args.historyRepository,
				shoppingRepository: tt.args.shoppingRepository}, s)
		})
	}
}

func TestShopService_GetInfo(t *testing.T) {
	type inputArgs struct {
		username                         string
		getCoinsAmountOutputAmount       int
		getCoinsAmountOutputError        error
		getInventoryOutputInventory      []model.Item
		getInventoryOutputError          error
		getReceivedHistoryOutputReceived []model.Receive
		getReceivedHistoryOutputError    error
		getSentHistoryOutputSent         []model.Send
		getSentHistoryOutputError        error
	}

	tests := []struct {
		name             string
		args             inputArgs
		wantErr          *apierror.APIError
		wantedErrMessage string
	}{
		{
			name: "success",
			args: inputArgs{
				username:                   "username",
				getCoinsAmountOutputAmount: 1000,
			},
		},
		{
			name: "err get coin amount",
			args: inputArgs{
				username:                   "username",
				getCoinsAmountOutputAmount: 0,
				getCoinsAmountOutputError:  apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
			},
			wantErr: &apierror.InternalError,
		},
		{
			name: "err get inventory",
			args: inputArgs{
				username:                   "username",
				getCoinsAmountOutputAmount: 0,
				getInventoryOutputError:    apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
			},
			wantErr: &apierror.InternalError,
		},
		{
			name: "err get received history",
			args: inputArgs{
				username:                      "username",
				getCoinsAmountOutputAmount:    0,
				getReceivedHistoryOutputError: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
			},
			wantErr: &apierror.InternalError,
		},
		{
			name: "err get sent history",
			args: inputArgs{
				username:                   "username",
				getCoinsAmountOutputAmount: 0,
				getSentHistoryOutputError:  apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
			},
			wantErr: &apierror.InternalError,
		},
	}
	var log *slog.Logger
	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shopRepository := new(MockRepository)
			s := NewShopService(log, shopRepository, shopRepository, shopRepository)
			shopRepository.On("GetCoinsAmount", tt.args.username).
				Return(tt.args.getCoinsAmountOutputAmount, tt.args.getCoinsAmountOutputError)
			shopRepository.On("GetInventory", tt.args.username).
				Return(tt.args.getInventoryOutputInventory, tt.args.getInventoryOutputError)
			shopRepository.On("GetCoinReceivedHistory", tt.args.username).
				Return(tt.args.getReceivedHistoryOutputReceived, tt.args.getReceivedHistoryOutputError)
			shopRepository.On("GetCoinSentHistory", tt.args.username).
				Return(tt.args.getSentHistoryOutputSent, tt.args.getSentHistoryOutputError)
			info, err := s.GetInfo(tt.args.username)
			t.Log(tt.name, fmt.Sprintf("%T", err), err, info)
			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErr.Status, apiErr.Status)
				assert.Equal(t, tt.wantErr.Message, apiErr.Message)
				return
			}
			assert.NoError(t, err)

		})
	}
}

func TestShopService_SendCoin(t *testing.T) {
	type inputArgs struct {
		sendCoinOutputErr error
		username          string
		send              model.Send
	}
	tests := []struct {
		name             string
		args             inputArgs
		wantErr          *apierror.APIError
		wantedErrMessage string
	}{
		{
			name: "success",
			args: inputArgs{
				username: "username",
				send: model.Send{
					ToUser: "toUser",
					Amount: 344,
				},
			},
		},
		{
			name: "err in repository.SendCoin",
			args: inputArgs{
				sendCoinOutputErr: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				username:          "username",
				send: model.Send{
					ToUser: "toUser",
					Amount: 344,
				},
			},
			wantErr: &apierror.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log *slog.Logger
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			shopRepository := new(MockRepository)
			s := NewShopService(log, shopRepository, shopRepository, shopRepository)
			shopRepository.On("SendCoin", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.args.sendCoinOutputErr)
			err := s.SendCoin(tt.args.username, tt.args.send)

			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, tt.wantErr.Message, apiErr.Message)
				return
			}
			assert.NoError(t, err)

		})
	}
}

func TestShopService_Buy(t *testing.T) {
	type inputArgs struct {
		buyOutputErr error
		username     string
		item         string
	}
	tests := []struct {
		name             string
		args             inputArgs
		wantErr          *apierror.APIError
		wantedErrMessage string
	}{
		{
			name: "success",
			args: inputArgs{
				username: "username",
				item:     "cup",
			},
		},
		{
			name: "error in repository.Buy",
			args: inputArgs{
				buyOutputErr: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				username:     "username",
				item:         "cup",
			},
			wantErr: &apierror.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log *slog.Logger
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			shopRepository := new(MockRepository)
			s := NewShopService(log, shopRepository, shopRepository, shopRepository)
			shopRepository.On("Buy", mock.Anything, mock.Anything).
				Return(tt.args.buyOutputErr)
			err := s.Buy(tt.args.username, tt.args.item)

			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, tt.wantErr.Message, apiErr.Message)
				return
			}
			assert.NoError(t, err)

		})
	}
}
