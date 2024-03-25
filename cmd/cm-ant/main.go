package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
	"github.com/cloud-barista/cm-ant/pkg/load/api/handler"
	"github.com/cloud-barista/cm-ant/pkg/load/domain"
	"github.com/gin-gonic/gin"
)

var once sync.Once

func main() {

	once.Do(func() {
		err := configuration.InitConfig("")
		if err != nil {
			log.Println(err)
			log.Fatal("error while reading config file.")
		}
	})

	// TODO: DB Configuration - confirm kind of db
	// temp db init
	domain.InitializeDatabase()
	router := SetRouter()

	//// background worker settings
	//worker := utils.NewWorker(30 * time.Minute)
	//go worker.Run()
	//defer worker.Shutdown()

	log.Fatal(router.Run(fmt.Sprintf(":%s", configuration.Get().Server.Port)))
}

func SetRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob(configuration.JoinRootPathWith("/web/templates/*"))

	antRouter := router.Group("/ant")

	{
		loadRouter := antRouter.Group("/load")

		{
			loadRouter.POST("/install", handler.InstallLoadGeneratorHandler())
			loadRouter.POST("/start", handler.RunLoadTestHandler())
			loadRouter.POST("/stop", handler.StopLoadTestHandler())
		}
	}

	return router
}
