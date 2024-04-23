package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/load/api/handler"
	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/cloud-barista/cm-ant/docs"
	"github.com/cloud-barista/cm-ant/pkg/configuration"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var once sync.Once

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// @title CM-ANT API
// @version 0.1
// @description
// @BasePath /ant

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

	e.Static("/web/templates", configuration.RootPath()+"/web/templates")
	e.Static("/css", configuration.RootPath()+"/web/css")
	e.Static("/js", configuration.RootPath()+"/web/js")

	e.Use(middleware.Logger(), middleware.Recover(), middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "request timeout",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			log.Println(c.Path())
		},
		Timeout: 120 * time.Second,
	}))

	// config template
	tmpl := configuration.NewTemplate()

	e.Renderer = tmpl

	antRouter := e.Group("/ant")

	{

		antRouter.GET("/swagger/*", echoSwagger.WrapHandler)
		antRouter.GET("", func(c echo.Context) error {
			return c.Render(http.StatusOK, "home.page.tmpl", nil)
		})

		antRouter.GET("/results", func(c echo.Context) error {
			result, err := services.GetAllLoadExecutionConfig()

			if err != nil {
				log.Printf("error while get load test execution config; %+v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{
					"message": "something went wrong.try again.",
				})

			}

			return c.Render(http.StatusOK, "results.page.tmpl", result)
		})

		antRouter.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "CM-Ant API server is running",
			})
		})

		connectionRouter := antRouter.Group("/env")

		{
			connectionRouter.GET("", handler.GetAllLoadEnvironments())
		}

		loadRouter := antRouter.Group("/load")

		{
			loadRouter.POST("/install", handler.InstallLoadGeneratorHandler())
			loadRouter.POST("/start", handler.RunLoadTestHandler())
			loadRouter.POST("/stop", handler.StopLoadTestHandler())
			loadRouter.GET("/result", handler.GetLoadTestResultHandler())
			loadRouter.GET("/config", handler.GetAllLoadConfigHandler())
			loadRouter.GET("/config/:loadTestKey", handler.GetLoadConfigHandler())
			loadRouter.GET("/state", handler.GetAllLoadExecutionStateHandler())
		}
	}
	return e
}
