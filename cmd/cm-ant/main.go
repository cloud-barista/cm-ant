package main

import (
	"fmt"
	"github.com/cloud-barista/cm-ant/pkg/load/api/handler"
	"github.com/gin-gonic/gin"
	"log"
	"sync"

	"github.com/cloud-barista/cm-ant/pkg/configuration"
)

var once sync.Once

func main() {

	once.Do(func() {
		err := configuration.Initialize()
		if err != nil {
			log.Println(err)
			log.Fatal("error while reading config file.")
		}
		router := InitRouter()
		log.Fatal(router.Run(fmt.Sprintf(":%s", configuration.Get().Server.Port)))
	})
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob(configuration.JoinRootPathWith("/web/templates/*"))

	antRouter := router.Group("/ant")

	{
		loadRouter := antRouter.Group("/load")

		{
			loadRouter.POST("/install", handler.InstallLoadGeneratorHandler())
			loadRouter.POST("/start", handler.RunLoadTestHandler())
			loadRouter.POST("/stop", handler.StopLoadTestHandler())
			loadRouter.GET("/:testId/result", handler.GetLoadTestResultHandler())
		}
	}

	return router
}
