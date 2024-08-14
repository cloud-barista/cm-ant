package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloud-barista/cm-ant/internal/app"
	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/utils"
)

// InitRouter initializes the routing for CM-ANT API server.

// @title CM-ANT REST API
// @version 0.2.2
// @description CM-ANT REST API swagger document.
// @basePath /ant
func main() {

	err := utils.Script(utils.JoinRootPathWith("/script/install_rsync.sh"), []string{})
	if err != nil {
		log.Fatal("required tool can not install")
	}
	utils.LogInfo("Starting CM-Ant server initialization...")

	// Initialize the configuration for CM-Ant server
	err = config.InitConfig()
	if err != nil {
		log.Fatalf("[ERROR] CM-Ant server config error: %v", err)
	}

	// Create a new instance of the CM-Ant server
	s, err := app.NewAntServer()
	if err != nil {
		log.Fatalf("[ERROR] CM-Ant server creation error: %v", err)
	}

	// Initialize the router for the CM-Ant server
	err = s.InitRouter()
	if err != nil {
		log.Fatalf("[ERROR] CM-Ant server init router error: %v", err)
	}

	utils.LogInfo("CM-Ant server initialization completed successfully.")

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	utils.LogInfo("Starting the CM-Ant server...")
	go func() {
		if err := s.Start(); err != nil {
			log.Fatalf("[ERROR] CM-Ant start server error: %v", err)
		}
	}()

	utils.LogInfo("CM-Ant server started successfully. Waiting for termination signal...")

	// Wait for termination signal
	<-stop

	utils.LogInfo("Shutting down CM-Ant server...")

	// Perform any necessary cleanup actions here, such as closing connections or saving state.
	// Optionally wait for pending operations to complete gracefully.

	utils.LogInfo("CM-Ant server stopped gracefully.")
	os.Exit(0)
}
