package repository

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type InfoRepository struct {
	logger *slog.Logger
	db     *sqlx.DB
}

func NewInfoRepository(logger *slog.Logger, db *sqlx.DB) *InfoRepository {
	return &InfoRepository{
		logger: logger,
		db:     db,
	}
}

func (i *InfoRepository) GetCoinsAmount(username string) (int, error) {
	const op = "repository.info.GetCoinsAmount"

	query := fmt.Sprintf(`SELECT balance FROM %s WHERE username = $1`, usersTable)
	var balance int

	if err := i.db.Get(&balance, query, username); err != nil {
		return 0, apierror.NewAPIError(apierror.InternalError,
			errors.Wrapf(err, "%s: (failed get user's balance)", op))
	}

	return balance, nil
}

func (i *InfoRepository) GetInventory(username string) ([]model.Item, error) {
	const op = "repository.info.GetInventory"

	query := fmt.Sprintf(`SELECT item, quantity FROM %s WHERE username = $1`, purchasesTable)
	var inventory []model.Item

	if err := i.db.Select(&inventory, query, username); err != nil {
		return nil, apierror.NewAPIError(apierror.InternalError,
			errors.Wrapf(err, "%s: (failed get user's inventory)", op))
	}

	return inventory, nil
}
