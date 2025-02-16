package service

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) Auth(username, passwordHash string) error {
	args := m.Called(username, passwordHash)
	return args.Error(0)
}

func TestNewAuthService(t *testing.T) {
	type inputArgs struct {
		logger         *slog.Logger
		authRepository AuthRepository
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
			s := NewAuthService(tt.args.logger, tt.args.authRepository)
			assert.Equal(t, &AuthService{
				logger:         tt.args.logger,
				authRepository: tt.args.authRepository}, s)
		})
	}
}

func TestAuthService_GenerateTokenParseToken(t *testing.T) {
	type inputArgs struct {
		username      string
		envTokenTTL   string
		envSigningKey string
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
				username:      "username",
				envTokenTTL:   "4",
				envSigningKey: "jgrh4r5ehg",
			},
		},
		{
			name: "invalid signing key",
			args: inputArgs{
				username:      "username",
				envTokenTTL:   "4",
				envSigningKey: "",
			},
			wantErr:        &apierror.InternalError,
			wantErrMessage: "service.auth.GenerateToken: (failed get sign key): sign key is empty",
		},
		{
			name: "invalid token TTL",
			args: inputArgs{
				username:      "username",
				envTokenTTL:   "stroka",
				envSigningKey: "gr5h567ui896l7kj",
			},
			wantErr: &apierror.InternalError,
			wantErrMessage: fmt.Sprintf(
				"service.auth.GenerateToken: (env %s must be numeric): strconv.Atoi: parsing \"stroka\": invalid syntax",
				model.EnvTokenTTLHours),
		},
	}
	var log *slog.Logger
	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	s := NewAuthService(log, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(model.EnvTokenTTLHours, tt.args.envTokenTTL)
			os.Setenv(model.EnvSigningKey, tt.args.envSigningKey)
			token, errGenerate := s.GenerateToken(tt.args.username)

			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(errGenerate, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, apiErr.Error(),
					fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage))
				return
			}
			username, errParse := s.ParseToken(token)
			assert.NoError(t, errGenerate)
			assert.NoError(t, errParse)
			assert.Equal(t, tt.args.username, username)
		})
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	godotenv.Load()
	type inputArgs struct {
		username      string
		envTokenTTL   string
		envSigningKey string
	}
	tests := []struct {
		name           string
		args           inputArgs
		wantErr        *apierror.APIError
		wantErrMessage string
	}{
		{
			name: "changed signing key",
			args: inputArgs{
				username:      "username",
				envTokenTTL:   "4",
				envSigningKey: "ffege4g4gh4wa",
			},
			wantErr:        &apierror.BadTokenError,
			wantErrMessage: "service.auth.ParseToken: (failed parse token): token contains an invalid number of segments",
		},
	}
	var log *slog.Logger
	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	s := NewAuthService(log, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			godotenv.Load()
			token, errGenerate := s.GenerateToken(tt.args.username)
			os.Setenv(model.EnvTokenTTLHours, tt.args.envTokenTTL)
			os.Setenv(model.EnvSigningKey, tt.args.envSigningKey)
			username, errParse := s.ParseToken(token)
			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(errParse, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErr.Status, apiErr.Status)
				assert.Equal(t, fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage), apiErr.Error())
				return
			}

			assert.NoError(t, errGenerate)
			assert.NoError(t, errParse)
			assert.Equal(t, tt.args.username, username)
		})
	}
}

func TestAuthService_generatePassword(t *testing.T) {
	type inputArgs struct {
		password string
		envSalt  string
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
				password: "username",
				envSalt:  "jngvurj",
			},
		},
		{
			name: "empty salt",
			args: inputArgs{
				password: "username",
				envSalt:  "",
			},
			wantErr:        &apierror.InternalError,
			wantErrMessage: "salt is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(model.EnvPasswordSalt, tt.args.envSalt)
			passwordHash1, err1 := generatePasswordHash(tt.args.password)
			if tt.wantErr != nil {
				var apiErr apierror.APIError
				ok := errors.As(err1, &apiErr)
				assert.True(t, ok)
				assert.Equal(t, apiErr.Status, tt.wantErr.Status)
				assert.Equal(t, apiErr.Error(),
					fmt.Sprintf("%s: %s", tt.wantErr.Message, tt.wantErrMessage))
				return
			}
			passwordHash2, err2 := generatePasswordHash(tt.args.password)
			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.Equal(t, passwordHash2, passwordHash1)
		})
	}
}

func TestAuth(t *testing.T) {
	type inputArgs struct {
		username        string
		password        string
		envSalt         string
		authOutputError error
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
				password: "15646156",
				envSalt:  "vsso9evijo",
			},
		},
		{
			name: "empty salt",
			args: inputArgs{
				username: "username",
				password: "15646156",
				envSalt:  "",
			},
			wantErr:        &apierror.InternalError,
			wantErrMessage: "salt is empty",
		},
	}
	var log *slog.Logger
	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	authRepository := new(MockAuthRepository)
	s := NewAuthService(log, authRepository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(model.EnvPasswordSalt, tt.args.envSalt)
			authRepository.On("Auth", mock.Anything, mock.Anything).Return(tt.args.authOutputError)
			err := s.Auth(model.AuthInput{Username: tt.args.username, Password: tt.args.password})

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
