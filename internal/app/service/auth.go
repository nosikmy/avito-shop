package service

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type AuthRepository interface {
	Auth(username, passwordHash string) error
}

type AuthService struct {
	logger         *slog.Logger
	authRepository AuthRepository
}

func NewAuthService(logger *slog.Logger, a AuthRepository) *AuthService {
	return &AuthService{
		logger:         logger,
		authRepository: a,
	}
}

func generatePasswordHash(password string) (string, error) {
	hash := sha3.New256()
	hash.Write([]byte(password))

	salt := os.Getenv(model.EnvPasswordSalt)
	if salt == "" {
		return "", apierror.NewAPIErrorWithMsg(apierror.InternalError, "salt is empty")
	}

	return fmt.Sprintf("%x", hash.Sum([]byte(salt))), nil
}

func (a *AuthService) Auth(input model.AuthInput) error {
	const op = "service.auth.Auth"

	passwordHash, err := generatePasswordHash(input.Password)
	if err != nil {
		return fmt.Errorf("%s: (failed generate password hash): %w", op, err)
	}

	if err := a.authRepository.Auth(input.Username, passwordHash); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *AuthService) GenerateToken(username string) (string, error) {
	const op = "service.auth.GenerateToken"

	tokenTTL, err := strconv.Atoi(os.Getenv(model.EnvTokenTTLHours))
	if err != nil {
		return "", apierror.NewAPIError(apierror.InternalError,
			errors.Wrapf(err, "%s: (env %s must be numeric)", op, model.EnvTokenTTLHours))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Duration(tokenTTL) * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		Id:        username,
	})

	signingKey := os.Getenv(model.EnvSigningKey)
	if signingKey == "" {
		return "", apierror.NewAPIErrorWithMsg(apierror.InternalError, op+": (failed get sign key): sign key is empty")
	}

	signedToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", apierror.NewAPIError(apierror.BadTokenError, errors.Wrapf(err, "%s: (failed sign token)", op))
	}

	return signedToken, nil
}

func (a *AuthService) ParseToken(token string) (string, error) {
	const op = "service.auth.ParseToken"

	t, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", apierror.NewAPIErrorWithMsg(apierror.BadTokenError, op+": (failed parse token): invalid signing method")
		}

		signingKey := os.Getenv(model.EnvSigningKey)
		if signingKey == "" {
			return "", apierror.NewAPIErrorWithMsg(apierror.InternalError, op+": (failed get sign key): sign key is empty")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return "", apierror.NewAPIError(apierror.BadTokenError, errors.Wrapf(err, "%s: (failed parse token)", op))
	}

	claims, ok := t.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", apierror.NewAPIErrorWithMsg(apierror.BadTokenError, op+": (failed parse token): invalid access token claims type")
	}

	return claims.Id, nil
}
