package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/internal/infra/db"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// antRepositories holds the various Repositorie used by the application.
type antRepositories struct {
	loadRepo *load.LoadRepository
}

// antServices holds the various Service used by the application.
type antServices struct {
	loadService *load.LoadService
}

// AntServer represents the server instance for the CM-Ant application.
// It contains an Echo instance for handling HTTP requests
// and multiple services for validating the core functionality of cloud migration.
type AntServer struct {
	e        *echo.Echo
	services *antServices
}

// NewAntServer initializes and returns a new instance of AntServer.
func NewAntServer() (*AntServer, error) {

	e := echo.New()

	conn, err := initializeDBConn()
	if err != nil {
		return nil, fmt.Errorf("CM-Ant database initialize error: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	tumblebugClient := tumblebug.NewTumblebugClient(client)
	repos := initializeRepositories(conn)
	services := initializeServices(repos, tumblebugClient)

	return &AntServer{
		e:        e,
		services: services,
	}, nil
}

// initializeDBConn establishes a connection to the database and returns it.
// It returns an error if the connection fails.
func initializeDBConn() (*gorm.DB, error) {
	conn, err := db.NewDBConnection()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// initializeRepositories initializes the repositories with the given database connection.
func initializeRepositories(conn *gorm.DB) *antRepositories {
	loadRepo := load.NewLoadRepository(conn)

	return &antRepositories{
		loadRepo: loadRepo,
	}
}

// initializeServices initializes the services with the given repositories and various client.
func initializeServices(repos *antRepositories, tbClient *tumblebug.TumblebugClient) *antServices {
	loadServ := load.NewLoadService(repos.loadRepo, tbClient)

	return &antServices{
		loadService: loadServ,
	}
}

// Start launches the Echo HTTP server on the port specified in the application
// configuration. It returns an error if the server fails to start.
func (a *AntServer) Start() error {
	return a.e.Start(fmt.Sprintf(":%s", config.AppConfig.Server.Port))
}
