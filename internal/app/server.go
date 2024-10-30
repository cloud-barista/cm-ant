package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/internal/infra/db"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/spider"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// JsonResult is a dummy struct for swagger annotations.
type JsonResult struct {
}

// antRepositories holds the various Repositorie used by the application.
type antRepositories struct {
	loadRepo *load.LoadRepository
	costRepo *cost.CostRepository
}

// antServices holds the various Service used by the application.
type antServices struct {
	loadService *load.LoadService
	costService *cost.CostService
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
	e.HideBanner = true

	conn, err := initializeDBConn()
	if err != nil {
		return nil, fmt.Errorf("CM-Ant database initialize error: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	tumblebugClient := tumblebug.NewTumblebugClient(client)
	spiderClient := spider.NewSpiderClient(client)
	repos := initializeRepositories(conn)
	services := initializeServices(repos, tumblebugClient, spiderClient)

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
	costRepo := cost.NewCostRepository(conn)

	return &antRepositories{
		loadRepo: loadRepo,
		costRepo: costRepo,
	}
}

// initializeServices initializes the services with the given repositories and various client.
func initializeServices(repos *antRepositories, tbClient *tumblebug.TumblebugClient, sClient *spider.SpiderClient) *antServices {
	loadServ := load.NewLoadService(repos.loadRepo, tbClient)

	cc := cost.NewAwsCostExplorerSpiderCostCollector(sClient, tbClient)
	pc := cost.NewSpiderPriceCollector(sClient)
	costServ := cost.NewCostService(repos.costRepo, pc, cc)

	return &antServices{
		loadService: loadServ,
		costService: costServ,
	}
}

// Start launches the Echo HTTP server on the port specified in the application
// configuration. It returns an error if the server fails to start.
func (a *AntServer) Start() error {
	return a.e.Start(fmt.Sprintf(":%s", config.AppConfig.Server.Port))
}
