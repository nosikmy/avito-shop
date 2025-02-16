package repository

import (
	"log/slog"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
)

func TestNewHistoryRepository(t *testing.T) {
	type inputArgs struct {
		logger *slog.Logger
		db     *sqlx.DB
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
			s := NewHistoryRepository(tt.args.logger, tt.args.db)
			assert.Equal(t, &HistoryRepository{
				logger: tt.args.logger,
				db:     tt.args.db}, s)
		})
	}
}
