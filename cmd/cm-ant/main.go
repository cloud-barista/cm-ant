package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloud-barista/cm-ant/internal/app"
	"github.com/cloud-barista/cm-ant/pkg/config"
)

func main() {
	// Initialize the configuration for CM-Ant server
	err := config.InitConfig()

	if err != nil {
		log.Fatal("CM-Ant server config error : ", err)
	}

	// Create a new instance of the CM-Ant server
	s, err := app.NewAntServer()
	if err != nil {
		log.Fatal("CM-Ant server creation error : ", err)
	}

	// Initialize the router for the CM-Ant server
	err = s.InitRouter()
	if err != nil {
		log.Fatal("CM-Ant server init router error : ", err)
	}

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the CM-Ant server in a separate goroutine
	go func() {
		if err := s.Start(); err != nil {
			log.Fatalf("CM-Ant start server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-stop

	// Attempt to gracefully shutdown the CM-Ant server
	log.Println("Shutting down CM-Ant server...")

	// Perform any necessary cleanup actions here, such as closing connections or saving state.

	// Optionally wait for pending operations to complete gracefully.

	// Exit the program
	log.Println("CM-Ant server stopped gracefully")
	os.Exit(0)
}
