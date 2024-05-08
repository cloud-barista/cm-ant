package main

import (
	"fmt"
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
	zerolog "github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var once sync.Once

// @title CM-ANT API
// @version 0.1
// @description

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

	e.Use(
		middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				LogError:         true,
				LogRequestID:     true,
				LogRemoteIP:      true,
				LogHost:          true,
				LogMethod:        true,
				LogURI:           true,
				LogUserAgent:     false,
				LogStatus:        true,
				LogLatency:       true,
				LogContentLength: true,
				LogResponseSize:  true,
				LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
					if v.Error == nil {
						zerolog.Info().
							Str("id", v.RequestID).
							Str("client_ip", v.RemoteIP).
							Str("host", v.Host).
							Str("method", v.Method).
							Str("URI", v.URI).
							Int("status", v.Status).
							Int64("latency", v.Latency.Nanoseconds()).
							Str("latency_human", v.Latency.String()).
							Str("bytes_in", v.ContentLength).
							Int64("bytes_out", v.ResponseSize).
							Msg("request")
					} else {
						zerolog.Error().
							Err(v.Error).
							Str("id", v.RequestID).
							Str("client_ip", v.RemoteIP).
							Str("host", v.Host).
							Str("method", v.Method).
							Str("URI", v.URI).
							Int("status", v.Status).
							Int64("latency", v.Latency.Nanoseconds()).
							Str("latency_human", v.Latency.String()).
							Str("bytes_in", v.ContentLength).
							Int64("bytes_out", v.ResponseSize).
							Msg("request error")
					}
					return nil
				},
			},
		),
		middleware.Recover(),
		middleware.RequestID(),
	)
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "request timeout",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			log.Println(c.Path())
		},
		Timeout: 300 * time.Second,
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

		apiRouter := antRouter.Group("/api")

		versionRouter := apiRouter.Group("/v1")

		versionRouter.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "CM-Ant API server is running",
			})
		})

		connectionRouter := versionRouter.Group("/env")

		{
			connectionRouter.GET("", handler.GetAllLoadEnvironments())
		}

		loadRouter := versionRouter.Group("/load")

		{
			// load tester
			loadRouter.POST("/tester", handler.InstallLoadTesterHandler())
			loadRouter.DELETE("/tester", handler.UninstallLoadTesterHandler())

			// load test execution
			loadRouter.POST("/start", handler.RunLoadTestHandler())
			loadRouter.POST("/stop", handler.StopLoadTestHandler())

			// load test result
			loadRouter.GET("/result", handler.GetLoadTestResultHandler())
			loadRouter.GET("/result/metrics", handler.GetLoadTestMetricsHandler())

			// load test history
			loadRouter.GET("/config", handler.GetAllLoadConfigHandler())
			loadRouter.GET("/config/:loadTestKey", handler.GetLoadConfigHandler())

			// load test state
			loadRouter.GET("/state", handler.GetAllLoadExecutionStateHandler())
			loadRouter.GET("/state/:loadTestKey", handler.GetLoadExecutionStateHandler())

			// load test metrics agent
			loadRouter.POST("/agent", handler.InstallAgent())
			loadRouter.DELETE("/agent/:agentId", handler.UninstallAgent())

			// TBD
			loadRouter.POST("/mock/migrate", handler.MockMigration())

		}

		version2Router := apiRouter.Group("/v2")
		lr := version2Router.Group("/load")
		{
			// load test metrics agent
			lr.POST("/agent", handler.InstallAgentV2())
			lr.GET("/agent", handler.GetAllAgentInstallInfo())
			lr.DELETE("/agent/:agentInstallInfoId", handler.UninstallAgentV2())
		}

	}
	return e
}
