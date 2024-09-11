package app

import (
	"strings"
	"time"

	_ "github.com/cloud-barista/cm-ant/api"

	"github.com/labstack/echo/v4/middleware"

	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/labstack/echo/v4"
	zerolog "github.com/rs/zerolog/log"
)

// setMiddleware configures middleware for the Echo server.
func setMiddleware(e *echo.Echo) {
	logSkipPattern := [][]string{
		{"/ant/swagger/*"},
	}
	e.Use(
		middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				Skipper: func(c echo.Context) bool {
					path := c.Request().URL.Path
					query := c.Request().URL.RawQuery
					for _, patterns := range logSkipPattern {
						isAllMatched := true
						for _, pattern := range patterns {
							if !strings.Contains(path+query, pattern) {
								isAllMatched = false
								break
							}
						}
						if isAllMatched {
							return true
						}
					}
					return false
				},
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
							// Str("host", v.Host).
							Str("method", v.Method).
							Str("URI", v.URI).
							Int("status", v.Status).
							// Int64("latency", v.Latency.Nanoseconds()).
							Str("latency_human", v.Latency.String()).
							Str("bytes_in", v.ContentLength).
							Int64("bytes_out", v.ResponseSize).
							Msg("request")
					} else {
						zerolog.Error().
							Err(v.Error).
							Str("id", v.RequestID).
							Str("client_ip", v.RemoteIP).
							// Str("host", v.Host).
							Str("method", v.Method).
							Str("URI", v.URI).
							Int("status", v.Status).
							// Int64("latency", v.Latency.Nanoseconds()).
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

func RequestIdAndDetailsIssuer(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Make X-Request-Id visible to all handlers
		c.Response().Header().Set("Access-Control-Expose-Headers", echo.HeaderXRequestID)

		// Get or generate Request ID
		reqID := c.Request().Header.Get(echo.HeaderXRequestID)
		if reqID == "" {
			reqID = utils.CreateUniqIdBaseOnUnixTime()
		}

		// Set Request on the context
		c.Set("RequestID", reqID)

		return next(c)
	}
}
