package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
)

type AuthRepository struct {
	logger *slog.Logger
	db     *sqlx.DB
}

func NewAuthRepository(logger *slog.Logger, db *sqlx.DB) *AuthRepository {
	return &AuthRepository{
		logger: logger,
		db:     db,
	}
}

func (a *AuthRepository) Auth(username, passwordHash string) error {
	const op = "repository.auth.Auth"

	query := fmt.Sprintf(`SELECT username, password_hash FROM %s WHERE username = $1`, usersTable)
	var user model.AuthDB

	if err := a.db.Get(&user, query, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return a.createNewUser(username, passwordHash)
		}
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed get user)", op))
	}

	if passwordHash != user.PasswordHash {
		return apierror.NewAPIErrorWithMsg(apierror.WrongPasswordError, op+": (failed sign in user): wrong password")
	}

	return nil
}

func (a *AuthRepository) createNewUser(username, passwordHash string) error {
	const op = "repository.auth.createNewUser"

	moneyForStart, err := strconv.Atoi(os.Getenv(model.EnvMoneyForStart))
	if err != nil {
		return apierror.NewAPIError(apierror.InternalError,
			errors.Wrapf(err, "%s: (%s must be numeric)", op, model.EnvMoneyForStart))
	}

	querySignUp := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3)`, usersTable)
	if _, err := a.db.Exec(querySignUp, username, passwordHash, moneyForStart); err != nil {
		return apierror.NewAPIError(apierror.InternalError, errors.Wrapf(err, "%s: (failed sign up user)", op))
	}

	return nil
}
