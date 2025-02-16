package repository

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type HistoryRepository struct {
	logger *slog.Logger
	db     *sqlx.DB
}

func NewHistoryRepository(logger *slog.Logger, db *sqlx.DB) *HistoryRepository {
	return &HistoryRepository{
		logger: logger,
		db:     db,
	}
}

func (h *HistoryRepository) GetCoinReceivedHistory(username string) ([]model.Receive, error) {
	const op = "repository.history.GetCoinReceivedHistory"

	query := fmt.Sprintf(`SELECT sender, amount FROM %s WHERE receiver = $1 ORDER BY created_at`, transactionsTable)
	var received []model.Receive

	if err := h.db.Select(&received, query, username); err != nil {
		return nil, apierror.NewAPIError(apierror.InternalError,
			errors.Wrapf(err, "%s: (failed get user's received history)", op))
	}

	return received, nil
}

func (h *HistoryRepository) GetCoinSentHistory(username string) ([]model.Send, error) {
	const op = "repository.history.GetCoinSentHistory"

	query := fmt.Sprintf(`SELECT receiver, amount FROM %s WHERE sender = $1 ORDER BY created_at`, transactionsTable)
	var sent []model.Send

	if err := h.db.Select(&sent, query, username); err != nil {
		return nil, apierror.NewAPIError(apierror.InternalError,
			errors.Wrapf(err, "%s: (failed get user's sent history)", op))
	}

	return sent, nil
}
