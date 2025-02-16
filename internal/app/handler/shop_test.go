package handler

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type MockShopService struct {
	mock.Mock
}

func (m *MockShopService) GetInfo(username string) (model.InfoOutput, error) {
	args := m.Called(username)
	info, _ := args.Get(0).(model.InfoOutput)
	return info, args.Error(1)
}

func (m *MockShopService) SendCoin(username string, send model.Send) error {
	args := m.Called(username, send)
	return args.Error(0)
}

func (m *MockShopService) Buy(username, item string) error {
	args := m.Called(username, item)
	return args.Error(0)
}

func TestHandler_GetInfo(t *testing.T) {
	type inputArgs struct {
		getInfoOutputInfo  model.InfoOutput
		getInfoOutputError error
		username           any
	}

	tests := []struct {
		name    string
		args    inputArgs
		wantErr *apierror.APIError
	}{
		{
			name: "success",
			args: inputArgs{
				getInfoOutputInfo: model.InfoOutput{
					Coins:     1000,
					Inventory: nil,
					CoinHistory: model.CoinHistory{
						Received: nil,
						Sent:     nil,
					},
				},
				getInfoOutputError: nil,
				username:           "username",
			},
		},
		{
			name: "err in service.GetInfo",
			args: inputArgs{
				getInfoOutputInfo:  model.InfoOutput{},
				getInfoOutputError: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				username:           "username",
			},
			wantErr: &apierror.InternalError,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shopService := new(MockShopService)
			log := slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			h := NewHandler(log, nil, shopService)
			shopService.On("GetInfo", mock.Anything, mock.Anything).Return(tt.args.getInfoOutputInfo, tt.args.getInfoOutputError)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(usernameField, tt.args.username)
			c.Request = httptest.NewRequest("POST", "localhost:8080/api/info", bytes.NewBufferString(""))
			h.GetInfo(c)

			if tt.wantErr != nil {
				assert.Equal(t, w.Code, tt.wantErr.Status)
				assert.Contains(t, w.Body.String(), tt.wantErr.Message)
				return
			}

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestHandler_SendCoin(t *testing.T) {
	type inputArgs struct {
		sendCoinOutputError error
		username            any
		body                string
	}

	tests := []struct {
		name    string
		args    inputArgs
		wantErr *apierror.APIError
	}{
		{
			name: "success",
			args: inputArgs{
				sendCoinOutputError: nil,
				username:            "username",
				body:                `{"toUser": "to_test_user", "amount": 15}`,
			},
		},
		{
			name: "invalid request body",
			args: inputArgs{
				sendCoinOutputError: nil,
				username:            "username",
				body:                "invalid json",
			},
			wantErr: &apierror.BadRequestError,
		},
		{
			name: "invalid request body",
			args: inputArgs{
				sendCoinOutputError: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				username:            "username",
				body:                `{"toUser": "to_test_user", "amount": 15}`,
			},
			wantErr: &apierror.InternalError,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shopService := new(MockShopService)
			log := slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			h := NewHandler(log, nil, shopService)
			shopService.On("SendCoin", mock.Anything, mock.Anything).Return(tt.args.sendCoinOutputError)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(usernameField, tt.args.username)
			c.Request = httptest.NewRequest("POST", "localhost:8080/api/buy/cup", bytes.NewBufferString(tt.args.body))

			h.SendCoin(c)

			if tt.wantErr != nil {
				assert.Equal(t, w.Code, tt.wantErr.Status)
				assert.Contains(t, w.Body.String(), tt.wantErr.Message)
				return
			}

			assert.Equal(t, w.Code, http.StatusOK)
		})
	}

}

func TestHandler_Buy(t *testing.T) {
	type inputArgs struct {
		buyOutputError error
		username       any
		param          string
	}
	tests := []struct {
		name    string
		args    inputArgs
		wantErr *apierror.APIError
	}{
		{
			name: "success",
			args: inputArgs{
				buyOutputError: nil,
				username:       "username",
				param:          "cup",
			},
		},
		{
			name: "invalid request body",
			args: inputArgs{
				buyOutputError: nil,
				username:       "username",
				param:          "",
			},
			wantErr: &apierror.InvalidItemError,
		},
		{
			name: "error in service.Buy",
			args: inputArgs{
				buyOutputError: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				username:       "username",
				param:          "cup",
			},
			wantErr: &apierror.InternalError,
		},
	}
	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shopService := new(MockShopService)
			log := slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			h := NewHandler(log, nil, shopService)
			shopService.On("Buy", mock.Anything, mock.Anything).Return(tt.args.buyOutputError)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(usernameField, tt.args.username)
			c.Request = httptest.NewRequest("POST", "localhost:8080/api/buy/cup", bytes.NewBufferString(""))
			c.Params = []gin.Param{{Key: "item", Value: tt.args.param}}
			h.Buy(c)

			if tt.wantErr != nil {
				assert.Equal(t, w.Code, tt.wantErr.Status)
				assert.Contains(t, w.Body.String(), tt.wantErr.Message)
				return
			}

			assert.Equal(t, w.Code, http.StatusOK)
		})
	}
}

func TestHandler_validateSendCoinInput(t *testing.T) {
	type inputArgs struct {
		username string
		input    model.Send
	}
	tests := []struct {
		name           string
		args           inputArgs
		wantErr        *apierror.APIError
		wantErrMessage string
	}{
		{
			name: "success",
			args: inputArgs{
				username: "username",
				input: model.Send{
					ToUser: "toUser",
					Amount: 100,
				},
			},
		},
		{
			name: "amount <= 0",
			args: inputArgs{
				username: "username",
				input: model.Send{
					ToUser: "toUser",
					Amount: -100,
				},
			},
			wantErr:        &apierror.BadRequestError,
			wantErrMessage: "invalid amount of coin",
		},
		{
			name: "empty receiver",
			args: inputArgs{
				username: "username",
				input: model.Send{
					ToUser: "",
					Amount: 100,
				},
			},
			wantErr:        &apierror.BadRequestError,
			wantErrMessage: "empty receiver",
		},
		{
			name: "send yourself",
			args: inputArgs{
				username: "username",
				input: model.Send{
					ToUser: "username",
					Amount: 100,
				},
			},
			wantErr:        &apierror.BadRequestError,
			wantErrMessage: "can't send to yourself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSendCoinInput(tt.args.input, tt.args.username)
			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, apiErr.Error(),
					fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage))
				return
			}
			assert.NoError(t, err)
		})
	}

}
