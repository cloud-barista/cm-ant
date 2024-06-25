package app

import (
	"log"
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/load/services"
	"github.com/cloud-barista/cm-ant/pkg/render"
	"github.com/cloud-barista/cm-ant/pkg/utils"

	_ "github.com/cloud-barista/cm-ant/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	zerolog "github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

// @title CM-ANT API
// @version 0.1
// @description
// @basePath /ant
// InitRouter initializes the routing for CM-ANT API server.
func (server *AntServer) InitRouter() error {
	setStatic(server.e)
	setMiddleware(server.e)

	tmpl := render.NewTemplate()
	server.e.Renderer = tmpl

	antRouter := server.e.Group("/ant")

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
	}

	apiRouter := antRouter.Group("/api")
	versionRouter := apiRouter.Group("/v1")

	versionRouter.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "CM-Ant API server is running",
		})
	})

	{

		connectionRouter := versionRouter.Group("/env")

		{
			connectionRouter.GET("", server.getAllLoadEnvironments)
		}

		loadRouter := versionRouter.Group("/load")

		{
			// load tester
			loadRouter.POST("/tester", server.installLoadTester)
			loadRouter.DELETE("/tester/:envId", server.uninstallLoadTester)

			// load test execution
			loadRouter.POST("/start", server.runLoadTest)
			loadRouter.POST("/stop", server.stopLoadTest)

			// load test result
			loadRouter.GET("/result", server.getLoadTestResult)
			loadRouter.GET("/result/metrics", server.getLoadTestMetrics)

			// load test history
			loadRouter.GET("/config", server.getAllLoadConfig)
			loadRouter.GET("/config/:loadTestKey", server.getLoadConfig)

			// load test state
			loadRouter.GET("/state", server.getAllLoadExecutionState)
			loadRouter.GET("/state/:loadTestKey", server.getLoadExecutionState)

			// load test metrics agent
			loadRouter.POST("/monitoring/agent/install", server.installMonitoringAgent, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))
			loadRouter.GET("/monitoring/agent", server.getAllMonitoringAgentInfos)
			loadRouter.POST("/monitoring/agent/uninstall", server.uninstallMonitoringAgent, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))
		}

	}

	{
		priceRouter := versionRouter.Group("/price")
		{
			priceRouter.POST("", server.getPriceInfo)
		}
	}

	return nil
}

// setStatic sets up static file serving for CSS, JS, and templates.
func setStatic(e *echo.Echo) {
	e.Static("/web/templates", utils.RootPath()+"/web/templates")
	e.Static("/css", utils.RootPath()+"/web/css")
	e.Static("/js", utils.RootPath()+"/web/js")
}

// setMiddleware configures middleware for the Echo server.
func setMiddleware(e *echo.Echo) {
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
		middleware.TimeoutWithConfig(
			middleware.TimeoutConfig{
				Skipper:      middleware.DefaultSkipper,
				ErrorMessage: "request timeout",
				OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
					log.Println(c.Path())
				},
				Timeout: 300 * time.Second,
			},
		),
		middleware.Recover(),
		middleware.RequestID(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)),
		middleware.CORS(),
	)
}
