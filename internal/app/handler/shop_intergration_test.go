//go:build integration
// +build integration

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/nosikmy/avito-shop/internal/app/apierror"
	"github.com/nosikmy/avito-shop/internal/app/model"
	"github.com/nosikmy/avito-shop/internal/app/repository"
	"github.com/nosikmy/avito-shop/internal/app/service"
	"github.com/nosikmy/avito-shop/internal/e2e"
)

var (
	db     *sqlx.DB
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
)

func TestMain(m *testing.M) {
	projectPath, err := e2e.GetProjectPath("avito-shop")
	if err != nil {
		log.Fatalln("failed get project path: %w", err)
	}
	if err := godotenv.Load(projectPath + "/.env"); err != nil {
		log.Fatalln("failed load .env file: ", err)
	}

	containerDB, closer, err := e2e.CreatePostgresDB()
	if err != nil {
		log.Fatal(err)
	}
	db = containerDB
	defer func() {
		if err := closer(); err != nil {
			log.Fatalln("failed close db container: ", err)
		}
	}()

	m.Run()
}

func initHandler() *Handler {
	authService := service.NewAuthService(logger, repository.NewAuthRepository(logger, db))
	shopService := service.NewShopService(logger, repository.NewInfoRepository(logger, db),
		repository.NewHistoryRepository(logger, db), repository.NewShoppingRepository(logger, db))

	return NewHandler(logger, authService, shopService)
}

func createUserDB(username, passwordHash string, balance int) error {
	query := "INSERT INTO users VALUES ($1, $2, $3)"
	if _, err := db.Exec(query, username, passwordHash, balance); err != nil {
		return err
	}
	return nil
}

func getUsersBalance(username string) (int, error) {
	query := "SELECT balance FROM users WHERE username=$1"
	var balance int
	if err := db.Get(&balance, query, username); err != nil {
		return 0, err
	}
	return balance, nil
}

var (
	merchPrices = map[string]int{
		"t-shirt":    80,
		"cup":        20,
		"book":       50,
		"pen":        10,
		"powerbank":  200,
		"hoody":      300,
		"umbrella":   200,
		"socks":      10,
		"wallet":     50,
		"pink-hoody": 500,
	}
)

func TestIntegrationHandler_Buy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := initHandler()
	username, password, balance := "user", "password", 1000

	if err := createUserDB(username, password, balance); err != nil {
		t.Errorf("failed create user: %s", err)
	}

	tests := []struct {
		name        string
		contextUser string
		itemParam   string
		wantErr     *apierror.APIError
	}{
		{
			name:        "simple test",
			contextUser: "user",
			itemParam:   "t-shirt",
			wantErr:     nil,
		},
		{
			name:        "second buy test",
			contextUser: "user",
			itemParam:   "powerbank",
			wantErr:     nil,
		},
		{
			name:        "unauthorized test",
			contextUser: "",
			itemParam:   "powerbank",
			wantErr:     &apierror.UnauthorizedError,
		},
		{
			name:        "no param test",
			contextUser: "user",
			itemParam:   "",
			wantErr:     &apierror.InvalidItemError,
		},
		{
			name:        "unknown item test",
			contextUser: "user",
			itemParam:   "something",
			wantErr:     &apierror.InvalidItemError,
		},
		{
			name:        "unknown user test",
			contextUser: "whoever",
			itemParam:   "cup",
			wantErr:     &apierror.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			testContext, _ := gin.CreateTestContext(w)
			testContext.Set("username", tt.contextUser)
			testContext.AddParam("item", tt.itemParam)

			h.Buy(testContext)

			if tt.wantErr != nil {
				var resp apierror.APIError
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed unmarshal body: %s", err)
				}

				assert.Equal(t, tt.wantErr.Message, resp.Message)
				assert.Equal(t, tt.wantErr.Status, w.Code)
				return
			}

			assert.Equal(t, http.StatusOK, w.Code)
			userBalance, err := getUsersBalance(username)
			if err != nil {
				t.Errorf("failed get user's balance: %s", err)
			}

			merchPrice := merchPrices[tt.itemParam]

			t.Logf("current balance: %d, merch: %s, price: %d, last balance: %d",
				userBalance, tt.itemParam, merchPrice, balance)

			assert.Equal(t, userBalance+merchPrice, balance)
			balance -= merchPrice
		})
	}
}

func TestIntegrationHandler_SendCoin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := initHandler()

	var (
		username1, password1, balance1 = "user1", "password", 1000
		username2, password2, balance2 = "user2", "password", 1000
	)

	if err := createUserDB(username1, password1, balance1); err != nil {
		t.Errorf("failed create user: %s", err)
	}
	if err := createUserDB(username2, password2, balance2); err != nil {
		t.Errorf("failed create user: %s", err)
	}

	tests := []struct {
		name              string
		body              string
		contextSenderUser string
		wantErr           *apierror.APIError
	}{
		{
			name:              "simple send coin test",
			body:              `{"toUser":"user2","amount":100}`,
			contextSenderUser: "user1",
			wantErr:           nil,
		},
		{
			name:              "second simple send coin test",
			body:              `{"toUser":"user1","amount":50}`,
			contextSenderUser: "user2",
			wantErr:           nil,
		},
		{
			name:              "not enough balance test",
			body:              `{"toUser":"user2","amount":50000}`,
			contextSenderUser: "user1",
			wantErr:           &apierror.NotEnoughMoneyError,
		},
		{
			name:              "empty body test",
			body:              ``,
			contextSenderUser: "user1",
			wantErr:           &apierror.BadRequestError,
		},
		{
			name:              "empty body test",
			body:              ``,
			contextSenderUser: "user1",
			wantErr:           &apierror.BadRequestError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			testContext, _ := gin.CreateTestContext(w)
			testContext.Set("username", tt.contextSenderUser)
			testContext.Request = &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(tt.body)),
			}

			h.SendCoin(testContext)

			if tt.wantErr != nil {
				var resp apierror.APIError
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed unmarshal body: %s", err)
				}

				assert.Equal(t, tt.wantErr.Message, resp.Message)
				assert.Equal(t, tt.wantErr.Status, w.Code)
				return
			}
			assert.Equal(t, http.StatusOK, w.Code)

			senderBalance, err := getUsersBalance(tt.contextSenderUser)
			if err != nil {
				t.Errorf("failed get user's balance: %s", err)
			}

			var body model.Send
			if err := json.Unmarshal([]byte(tt.body), &body); err != nil {
				t.Errorf("failed unmarshal body: %s", err)
			}

			receiverBalance, err := getUsersBalance(body.ToUser)
			if err != nil {
				t.Errorf("failed get user's balance: %s", err)
			}

			t.Logf("sender:[%s:%d], receiver:[%s:%d]", tt.contextSenderUser, senderBalance, body.ToUser, receiverBalance)

			switch {
			case tt.contextSenderUser == username1 && body.ToUser == username2:
				// balance1 senderBalance
				assert.Equal(t, balance1-body.Amount, senderBalance)
				// balance2 receiverBalance
				assert.Equal(t, balance2+body.Amount, receiverBalance)
				balance1 -= body.Amount
				balance2 += body.Amount
			case tt.contextSenderUser == username2 && body.ToUser == username1:
				// balance2 senderBalance
				assert.Equal(t, balance2-body.Amount, senderBalance)
				// balance1 receiverBalance
				assert.Equal(t, balance1+body.Amount, receiverBalance)
				balance2 -= body.Amount
				balance1 += body.Amount
			}
		})
	}
}
