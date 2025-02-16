package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

func (h *Handler) GetInfo(ctx *gin.Context) {
	const op = "handler.shop.GetInfo"

	username, err := getUsername(ctx)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while getting username", op))
		return
	}

	h.logger.Info("getting info", slog.String("username", username))

	info, err := h.shopService.GetInfo(username)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while getting info", op))
		return
	}

	h.logger.Info("Got info", slog.String("username", username))

	ctx.JSON(http.StatusOK, info)
}

func validateSendCoinInput(input model.Send, username string) error {
	switch {
	case input.Amount <= 0:
		return apierror.NewAPIErrorWithMsg(apierror.BadRequestError, "invalid amount of coin")
	case input.ToUser == "":
		return apierror.NewAPIErrorWithMsg(apierror.BadRequestError, "empty receiver")
	case username == input.ToUser:
		return apierror.NewAPIErrorWithMsg(apierror.BadRequestError, "can't send to yourself")
	default:
		return nil
	}
}

func (h *Handler) SendCoin(ctx *gin.Context) {
	const op = "handler.shop.SendCoin"

	username, err := getUsername(ctx)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while getting username", op))
		return
	}

	var input model.Send
	if err := ctx.ShouldBindJSON(&input); err != nil {
		apierror.LogAndRespondError(ctx, h.logger,
			apierror.NewAPIError(apierror.BadRequestError, errors.Wrap(err, op+": error while getting data from request body")))
		return
	}

	if err := validateSendCoinInput(input, username); err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while validating input", op))
		return
	}

	h.logger.Info("Sending coins",
		slog.String("from", username),
		slog.String("to", input.ToUser),
		slog.Int("amount", input.Amount))

	if err = h.shopService.SendCoin(username, input); err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while getting info", op))
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *Handler) Buy(ctx *gin.Context) {
	const op = "handler.shop.Buy"

	username, err := getUsername(ctx)
	if err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while getting username", op))
		return
	}

	item := ctx.Param("item")
	if item == "" {
		apierror.LogAndRespondError(ctx, h.logger, apierror.NewAPIErrorWithMsg(apierror.InvalidItemError, op+": empty item"))
		return
	}

	h.logger.Info("buying item", slog.String("username", username), slog.String("item", item))
	if err = h.shopService.Buy(username, item); err != nil {
		apierror.LogAndRespondError(ctx, h.logger, errors.Wrapf(err, "%s: error while buying item", op))
		return
	}

	h.logger.Info("item was bought", slog.String("username", username), slog.String("item", item))

	ctx.Status(http.StatusOK)
}
