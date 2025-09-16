package main

import (
	"fmt"

	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/cloud-barista/cm-ant/internal/app"
	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitRouter initializes the routing for CM-ANT API server.

// @title CM-ANT REST API
// @version 0.3.4
// @description CM-ANT REST API swagger document.
// @basePath /ant

type CallerHook struct{}

func (h CallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if pc, file, line, ok := runtime.Caller(3); ok {
		shortFile := file[strings.LastIndex(file, "/")+1:]
		e.Str("file", fmt.Sprintf("%s:%d", shortFile, line))
		funcName := strings.Replace(runtime.FuncForPC(pc).Name(), "github.com/cloud-barista/", "", 1)
		e.Str("func", funcName)
	}
}

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		level := strings.ToUpper(fmt.Sprintf("%s", i))
		switch level {
		case "DEBUG":
			return fmt.Sprintf("\033[36m| %-6s|\033[0m", level) // Cyan
		case "INFO":
			return fmt.Sprintf("\033[32m| %-6s|\033[0m", level) // Green
		case "WARN":
			return fmt.Sprintf("\033[33m| %-6s|\033[0m", level) // Yellow
		case "ERROR":
			return fmt.Sprintf("\033[31m| %-6s|\033[0m", level) // Red
		case "FATAL":
			return fmt.Sprintf("\033[35m| %-6s|\033[0m", level) // Magenta
		default:
			return fmt.Sprintf("| %-6s|", level) // Default color
		}
	}
	output.FormatMessage = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("message: \033[1m%s\033[0m", i)
	}

	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("\033[1m%s\033[0m", i)
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger().Hook(CallerHook{})

	err := utils.Script(utils.JoinRootPathWith("/script/install_required_utils.sh"), []string{})
	if err != nil {
		log.Fatal().Msg("required tool can not install")
	}

	// Initialize the configuration for CM-Ant server
	err = config.InitConfig()
	if err != nil {
		log.Fatal().Msgf("CM-Ant server config error: %v", err)
	}

	// Create a new instance of the CM-Ant server
	s, err := app.NewAntServer()
	if err != nil {
		log.Fatal().Msgf("CM-Ant server creation error: %v", err)
	}

	// Initialize the router for the CM-Ant server
	err = s.InitRouter()
	if err != nil {
		log.Fatal().Msgf("CM-Ant server init router error: %v", err)
	}

	log.Info().Msgf("CM-Ant server initialization completed successfully.")

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Info().Msgf("Starting the CM-Ant server...")
	go func() {
		if err := s.Start(); err != nil {
			log.Fatal().Msgf("CM-Ant start server error: %v", err)
		}
	}()

	log.Info().Msgf("CM-Ant server started successfully. Waiting for termination signal...")

	// Wait for termination signal
	<-stop

	log.Info().Msgf("Shutting down CM-Ant server...")

	// Perform any necessary cleanup actions here, such as closing connections or saving state.
	// Optionally wait for pending operations to complete gracefully.

	log.Info().Msgf("CM-Ant server stopped gracefully.")
	os.Exit(0)
}
