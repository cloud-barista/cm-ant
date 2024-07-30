package app

import (
	"time"

	"github.com/cloud-barista/cm-ant/internal/render"
	"github.com/cloud-barista/cm-ant/internal/utils"

	_ "github.com/cloud-barista/cm-ant/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	zerolog "github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (server *AntServer) InitRouter() error {
	setStatic(server.e)
	setMiddleware(server.e)

	tmpl := render.NewTemplate()
	server.e.Renderer = tmpl

	antRouter := server.e.Group("/ant")

	{
		antRouter.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	apiRouter := antRouter.Group("/api")
	versionRouter := apiRouter.Group("/v1")

	versionRouter.GET("/readyz", server.readyz)

	{

		loadRouter := versionRouter.Group("/load")

		{
			loadRouter.GET("/generators", server.getAllLoadGeneratorInstallInfo)
			loadRouter.POST("/generators", server.installLoadGenerator)
			loadRouter.DELETE("/generators/:loadGeneratorInstallInfoId", server.uninstallLoadGenerator)

			// load test metrics agent
			loadRouter.POST("/monitoring/agent/install", server.installMonitoringAgent, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))
			loadRouter.GET("/monitoring/agent", server.getAllMonitoringAgentInfos)
			loadRouter.POST("/monitoring/agent/uninstall", server.uninstallMonitoringAgent, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))

			loadTestRouter := loadRouter.Group("/tests")

			{
				// load test execution
				loadTestRouter.POST("/run", server.runLoadTest)
				loadTestRouter.POST("/stop", server.stopLoadTest)

				// load test state
				loadTestRouter.GET("/state", server.getAllLoadTestExecutionState)
				loadTestRouter.GET("/state/:loadTestKey", server.getLoadTestExecutionState)

				// load test history
				loadTestRouter.GET("/infos", server.getAllLoadTestExecutionInfos)
				loadTestRouter.GET("/infos/:loadTestKey", server.getLoadTestExecutionInfo)

				// load test result
				loadTestRouter.GET("/result", server.getLoadTestResult)
				loadTestRouter.GET("/result/metrics", server.getLoadTestMetrics)
			}
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
					utils.LogInfo(c.Path())
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
