package handler

import (
	"log"

	"github.com/labstack/echo/v4"
)

func InstallAgent() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("install agent")
		return nil
	}
}

func UninstallAgent() echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("uninstall agent")
		return nil
	}
}
