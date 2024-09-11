package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type AntResponse[T any] struct {
	ErrorMessage   string `json:"errorMessage,omitempty"`
	SuccessMessage string `json:"successMessage,omitempty"`
	Code           int    `json:"code,omitempty"`
	Result         T      `json:"result,omitempty"`
}

func errorResponseJson(statusCode int, message string) error {
	return echo.NewHTTPError(statusCode, AntResponse[string]{
		ErrorMessage: message,
		Code:         statusCode,
	})
}

func successResponseJson[T any](c echo.Context, successMessage string, result T) error {
	return c.JSON(http.StatusOK, AntResponse[T]{
		SuccessMessage: successMessage,
		Code:           http.StatusOK,
		Result:         result,
	})
}
