package repository

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type ShoppingRepository struct {
	logger *slog.Logger
	db     *sqlx.DB
}

func NewShoppingRepository(logger *slog.Logger, db *sqlx.DB) *ShoppingRepository {
	return &ShoppingRepository{
		logger: logger,
		db:     db,
	}
}

func (s *ShoppingRepository) SendCoin(fromUsername, toUsername string, amount int) error {
	const op = "repository.shopping.SendCoin"

	tx, err := s.db.Beginx()
	if err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed begin transaction)", op))
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			s.logger.Error("error while rollback: " + err.Error())
		}
	}()

	querySelectForUpdate := fmt.Sprintf(
		`WITH locked_users AS (
    				SELECT username, balance
					FROM %s
					WHERE username IN ($1, $2) FOR UPDATE
				)
				SELECT username, balance
				FROM locked_users
				WHERE username = $3`, usersTable)

	var user model.User
	if err := tx.Get(&user, querySelectForUpdate, fromUsername, toUsername, fromUsername); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed get user)", op))
	}

	if user.Balance < amount {
		return apierror.NewAPIErrorWithMsg(apierror.NotEnoughMoneyError, op+": (failed get user): not enough money")
	}

	querySend := fmt.Sprintf(`UPDATE %s SET balance = balance - $1 WHERE username = $2`, usersTable)
	if _, err := tx.Exec(querySend, amount, fromUsername); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed send money)", op))
	}

	queryReceive := fmt.Sprintf(`UPDATE %s SET balance = balance + $1 WHERE username = $2`, usersTable)
	if _, err := tx.Exec(queryReceive, amount, toUsername); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed receive money)", op))
	}

	queryAddTransaction := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3)`, transactionsTable)
	if _, err := tx.Exec(queryAddTransaction, fromUsername, toUsername, amount); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed save transaction)", op))
	}

	if err := tx.Commit(); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed commit)", op))
	}

	return nil
}

func (s *ShoppingRepository) Buy(username, item string) error {
	const op = "repository.shopping.Buy"

	tx, err := s.db.Beginx()
	if err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed begin transaction)", op))
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			s.logger.Error("error while rollback: " + err.Error())
		}
	}()

	queryGetItemPrice := fmt.Sprintf(`SELECT price FROM %s WHERE type = $1 FOR SHARE`, itemsTable)
	var itemPrice int
	if err := tx.Get(&itemPrice, queryGetItemPrice, item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apierror.NewAPIError(apierror.InvalidItemError, errors.Wrapf(err, "%s: (failed find item)", op))
		}

		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed get item)", op))
	}

	querySelectForUpdate := fmt.Sprintf(`SELECT username, balance FROM %s WHERE username = $1 FOR UPDATE`, usersTable)
	var user model.User
	if err = tx.Get(&user, querySelectForUpdate, username); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed get user)", op))
	}

	if user.Balance < itemPrice {
		return apierror.NewAPIErrorWithMsg(apierror.NotEnoughMoneyError, op+": (failed get user): not enough money")
	}

	queryBuy := fmt.Sprintf(`UPDATE %s SET balance = balance - $1 WHERE username = $2`, usersTable)
	if _, err = tx.Exec(queryBuy, itemPrice, username); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed buy item)", op))
	}

	queryAddPurchases := fmt.Sprintf(
		`INSERT into %s VALUES ($1, $2, 1) ON CONFLICT (username, item) DO UPDATE SET quantity = %s.quantity + 1`,
		purchasesTable, purchasesTable)
	if _, err = tx.Exec(queryAddPurchases, username, item); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed save purchas)", op))
	}

	if err := tx.Commit(); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed commit)", op))
	}

	return nil
}
