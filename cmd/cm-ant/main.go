package main

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
)

var once sync.Once

func main() {

	once.Do(func() {
		err := configuration.Initialize()
		if err != nil {
			log.Println(err)
			log.Fatal("error while reading config file.")
		}
		router := InitRouter()
		log.Fatal(router.Start(fmt.Sprintf(":%s", configuration.Get().Server.Port)))
	})
}

func InitRouter() *echo.Echo {
	e := echo.New()

	e.Static("/web/templates", configuration.RootPath()+"/web/template")
	e.Use(middleware.Logger(), middleware.Recover(), middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "request timeout",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			log.Println(c.Path())
		},
		Timeout: 120 * time.Second,
	}))

	antRouter := e.Group("/ant")

	{

		antRouter.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "CM-Ant API server is running",
			})
		})

		connectionRouter := antRouter.Group("/env")

		{
			connectionRouter.GET("", handler.GetAllRemoteConnection())
		}

		loadRouter := antRouter.Group("/load")

		{
			loadRouter.POST("/install", handler.InstallLoadGeneratorHandler())
			loadRouter.POST("/start", handler.RunLoadTestHandler())
			loadRouter.POST("/stop", handler.StopLoadTestHandler())
			loadRouter.GET("/result", handler.GetLoadTestResultHandler())
			loadRouter.GET("/config/:configId", handler.GetLoadConfigHandler())
			loadRouter.GET("/state", handler.GetAllLoadExecutionStateHandler())
		}
	}
	return e
}
