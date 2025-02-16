package apierror

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type APIError struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
	wrapped error
}

func (m APIError) Cause() error {
	return m.wrapped
}

func (m APIError) Unwrap() error {
	return m.wrapped
}

func (m APIError) Error() string {
	if m.wrapped != nil {
		return fmt.Sprintf("%s: %s", m.Message, m.wrapped.Error())
	}
	return m.Message
}

var (
	InternalError = APIError{
		Status:  http.StatusInternalServerError,
		Message: "internal error",
	}
	UnauthorizedError = APIError{
		Status:  http.StatusUnauthorized,
		Message: "unauthorized",
	}
	WrongPasswordError = APIError{
		Status:  http.StatusUnauthorized,
		Message: "wrong password",
	}
	BadAuthHeaderError = APIError{
		Status:  http.StatusUnauthorized,
		Message: "empty or invalid authorization header",
	}
	BadTokenError = APIError{
		Status:  http.StatusUnauthorized,
		Message: "empty or invalid token",
	}
	BadRequestError = APIError{
		Status:  http.StatusBadRequest,
		Message: "bad request error",
	}
	InvalidItemError = APIError{
		Status:  http.StatusBadRequest,
		Message: "no such item exists",
	}
	InvalidAuthInput = APIError{
		Status:  http.StatusBadRequest,
		Message: "invalid username or password",
	}
	NotEnoughMoneyError = APIError{
		Status:  http.StatusBadRequest,
		Message: "not enough money",
	}
)

func NewAPIError(apiErr APIError, err error) error {
	apiErr.wrapped = err
	return apiErr
}

func NewAPIErrorWithMsg(apiErr APIError, msg string) error {
	apiErr.wrapped = errors.New(msg)
	return apiErr
}

func GetAPIError(err error) APIError {
	if apiErr := new(APIError); errors.As(err, apiErr) {
		return *apiErr
	}
	return InternalError
}

func LogAndRespondError(ctx *gin.Context, l *slog.Logger, err error) {
	l.Error(err.Error())
	apiError := GetAPIError(err)
	ctx.AbortWithStatusJSON(apiError.Status, APIError{
		Message: apiError.Message,
	})
}
