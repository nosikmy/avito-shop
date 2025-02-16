package handler

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Auth(input model.AuthInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *MockAuthService) GenerateToken(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ParseToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func TestHandler_validateAuthInput(t *testing.T) {
	tests := []struct {
		name           string
		args           model.AuthInput
		wantErr        *apierror.APIError
		wantErrMessage string
	}{
		{
			name: "success",
			args: model.AuthInput{
				Username: "Ivan",
				Password: "123456",
			},
			wantErr:        nil,
			wantErrMessage: "",
		},
		{
			name: "empty username",
			args: model.AuthInput{
				Username: "",
				Password: "123456",
			},
			wantErr:        &apierror.InvalidAuthInput,
			wantErrMessage: "empty username",
		},
		{
			name: "invalid password",
			args: model.AuthInput{
				Username: "Ivan",
				Password: "1234",
			},
			wantErr:        &apierror.InvalidAuthInput,
			wantErrMessage: "password is too short, it must be at least 6 characters long",
		},
		{
			name: "empty password",
			args: model.AuthInput{
				Username: "Ivan",
				Password: "",
			},
			wantErr:        &apierror.InvalidAuthInput,
			wantErrMessage: "empty password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuthInput(tt.args)
			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, apiErr.Error(), fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage))
				return
			}
			assert.NoError(t, err)
		})
	}

}

func TestHandler_Auth(t *testing.T) {
	type inputArgs struct {
		authOutputError          error
		generateTokenOutputToken string
		generateTokenOutputError error
		url                      string
		body                     string
	}

	tests := []struct {
		name    string
		args    inputArgs
		wantErr *apierror.APIError
	}{
		{
			name: "success",
			args: inputArgs{
				authOutputError:          nil,
				generateTokenOutputToken: "test_token",
				generateTokenOutputError: nil,
				url:                      "localhost:8080/api/auth",
				body:                     `{"username": "testuser", "password": "testpass"}`,
			},
			wantErr: nil,
		},
		{
			name: "body is not json",
			args: inputArgs{
				authOutputError:          nil,
				generateTokenOutputToken: "test_token",
				generateTokenOutputError: nil,
				url:                      "localhost:8080/api/auth",
				body:                     `not json`,
			},
			wantErr: &apierror.BadRequestError,
		},
		{
			name: "err in service.Auth",
			args: inputArgs{
				authOutputError:          apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				generateTokenOutputToken: "",
				generateTokenOutputError: nil,
				url:                      "localhost:8080/api/auth",
				body:                     `{"username": "testuser", "password": "testpass"}`,
			},
			wantErr: &apierror.InternalError,
		},
		{
			name: "err in service.GenerateToken",
			args: inputArgs{
				authOutputError:          nil,
				generateTokenOutputToken: "",
				generateTokenOutputError: apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				url:                      "localhost:8080/api/auth",
				body:                     `{"username": "testuser", "password": "testpass"}`,
			},
			wantErr: &apierror.InternalError,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := new(MockAuthService)
			log := slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			h := NewHandler(log, authService, nil)
			authService.On("Auth", mock.Anything).Return(tt.args.authOutputError)
			authService.On("GenerateToken", mock.Anything).
				Return(tt.args.generateTokenOutputToken, tt.args.generateTokenOutputError)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", tt.args.url, bytes.NewBufferString(tt.args.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Auth(c)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Status, w.Code)
				assert.Contains(t, w.Body.String(), tt.wantErr.Message)
				return
			}
			assert.Equal(t, w.Code, http.StatusOK)
			assert.Contains(t, w.Body.String(), tt.args.generateTokenOutputToken)
		})
	}
}

func TestHandler_UserIdentify(t *testing.T) {
	type inputArgs struct {
		parseTokeOutputUsername string
		parseTokenOutputError   error
		authHeader              string
	}

	tests := []struct {
		name    string
		args    inputArgs
		wantErr *apierror.APIError
	}{
		{
			name: "success",
			args: inputArgs{
				parseTokeOutputUsername: "username",
				parseTokenOutputError:   nil,
				authHeader:              "Bearer valid_token",
			},
			wantErr: nil,
		},
		{
			name: "error in service.ParseHeader",
			args: inputArgs{
				parseTokeOutputUsername: "username",
				parseTokenOutputError:   apierror.NewAPIErrorWithMsg(apierror.InternalError, "mock"),
				authHeader:              "Bearer valid_token",
			},
			wantErr: &apierror.InternalError,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := new(MockAuthService)
			log := slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			h := NewHandler(log, authService, nil)
			authService.On("ParseToken", mock.Anything).
				Return(tt.args.parseTokeOutputUsername, tt.args.parseTokenOutputError)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set(authHeader, tt.args.authHeader)

			h.UserIdentify(c)

			if tt.wantErr != nil {
				assert.Equal(t, w.Code, tt.wantErr.Status)
				assert.Contains(t, w.Body.String(), tt.wantErr.Message)
				return
			}
			username, ok := c.Get("username")
			assert.True(t, ok)
			assert.Contains(t, username, tt.args.parseTokeOutputUsername)
		})
	}

}

func TestHandler_getTokenFromHeader(t *testing.T) {
	tests := []struct {
		name           string
		args           string
		wantToken      string
		wantErr        *apierror.APIError
		wantErrMessage string
	}{
		{
			name:      "success",
			args:      "Bearer test_token",
			wantToken: "test_token",
		},
		{
			name: "empty header",
			args: "",

			wantErr:        &apierror.BadAuthHeaderError,
			wantErrMessage: "empty Authorization header",
		}, {
			name:           "no token",
			args:           "Beagr78rer test_token",
			wantErr:        &apierror.BadAuthHeaderError,
			wantErrMessage: "no token",
		}, {
			name:           "empty token",
			args:           "Bearer ",
			wantErr:        &apierror.BadAuthHeaderError,
			wantErrMessage: "empty token",
		},
	}

	var apiErr apierror.APIError
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := getTokenFromHeader(tt.args)
			if tt.wantErr != nil {
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Empty(t, token)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, apiErr.Error(),
					fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, token, tt.wantToken)
		})
	}
}

func TestHandler_getUsername(t *testing.T) {
	type inputArgs struct {
		empty    bool
		username interface{}
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
				username: "test_username",
			},
		},
		{
			name: "no username in context",
			args: inputArgs{
				empty: true,
			},
			wantErr:        &apierror.UnauthorizedError,
			wantErrMessage: "username field not found in context",
		},
		{
			name: "wrong type",
			args: inputArgs{
				username: 123,
			},
			wantErr:        &apierror.UnauthorizedError,
			wantErrMessage: "username is not a string type",
		},
		{
			name: "empty username",
			args: inputArgs{
				username: "",
			},
			wantErr:        &apierror.UnauthorizedError,
			wantErrMessage: "empty username",
		},
	}

	var apiErr apierror.APIError

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			if !tt.args.empty {
				c.Set("username", tt.args.username)
			}

			username, err := getUsername(c)
			if tt.wantErr != nil {
				ok := errors.As(err, &apiErr)
				assert.True(t, ok)
				assert.Empty(t, username)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, apiErr.Error(), fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, username, tt.args.username)
		})
	}
}
