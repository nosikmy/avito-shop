package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

const (
	authHeader    = "Authorization"
	usernameField = "username"
)

func validateAuthInput(input model.AuthInput) error {
	switch {
	case input.Username == "":
		return apierror.NewAPIErrorWithMsg(apierror.InvalidAuthInput, "empty username")
	case input.Password == "":
		return apierror.NewAPIErrorWithMsg(apierror.InvalidAuthInput, "empty password")
	case len(input.Password) < 6:
		return apierror.NewAPIErrorWithMsg(
			apierror.InvalidAuthInput, "password is too short, it must be at least 6 characters long")
	default:
		return nil
	}
}

func (h *Handler) Auth(ctx *gin.Context) {
	const op = "handler.auth.Auth"
	var input model.AuthInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		apierror.LogAndRespondError(ctx, h.logger,
			apierror.NewAPIErrorWithMsg(apierror.BadRequestError, op+": "+"error while getting data from request body"))
		return
	}

	if err := validateAuthInput(input); err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: validation failed", op))
		return
	}

	h.logger.Info("authentication user", slog.String("username", input.Username))

	if err := h.authService.Auth(input); err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while authentication user", op))
		return
	}

	h.logger.Info("user authenticated", slog.String("username", input.Username))

	token, err := h.authService.GenerateToken(input.Username)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while generating token", op))
		return
	}

	h.logger.Info("token generated", slog.String("username", input.Username))

	ctx.JSON(http.StatusOK, model.AuthOutput{
		Token: token,
	})
}

func getTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", apierror.NewAPIErrorWithMsg(apierror.BadAuthHeaderError, "empty Authorization header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", apierror.NewAPIErrorWithMsg(apierror.BadAuthHeaderError, "no token")
	}

	if len(headerParts[1]) == 0 {
		return "", apierror.NewAPIErrorWithMsg(apierror.BadAuthHeaderError, "empty token")
	}

	return headerParts[1], nil
}

func (h *Handler) UserIdentify(ctx *gin.Context) {
	const op = "handler.auth.UserIdentify"

	header := ctx.GetHeader(authHeader)
	token, err := getTokenFromHeader(header)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while getting token", op))
		return
	}

	username, err := h.authService.ParseToken(token)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while parse token", op))
		return
	}

	ctx.Set(usernameField, username)
	ctx.Next()
}

func getUsername(ctx *gin.Context) (string, error) {
	data, ok := ctx.Get(usernameField)
	if !ok {
		return "", apierror.NewAPIErrorWithMsg(apierror.UnauthorizedError, "username field not found in context")
	}

	username, ok := data.(string)
	if !ok {
		return "", apierror.NewAPIErrorWithMsg(apierror.UnauthorizedError, "username is not a string type")
	}

	if username == "" {
		return "", apierror.NewAPIErrorWithMsg(apierror.UnauthorizedError, "empty username")
	}

	return username, nil
}
