package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloud-barista/cm-ant/api/handler"
	"github.com/cloud-barista/cm-ant/internal/domain"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
)

var SpiderUrl = os.Getenv("SPIDER_URL")
var TumblebugUrl = os.Getenv("TUMBLEBUG_URL")

func initConfig() {
	defaultSpiderUrl := "http://localhost:1024"
	defaultTumblebugUrl := "http://localhost:1323"

	if SpiderUrl == "" {
		SpiderUrl = defaultSpiderUrl
	}

	if TumblebugUrl == "" {
		TumblebugUrl = defaultTumblebugUrl
	}
}

func main() {
	initConfig()

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.SetTrustedProxies([]string{"IPv4", " IPv4 CIDRs", "IPv6 addresses"})

	// background worker settings
	worker := utils.NewWorker(30 * time.Minute)
	go worker.Run()
	defer worker.Shutdown()

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	loadTestRouter := router.Group("/load-test")
	{
		loadTestRouter.POST("/", handler.ExecuteLoadTestHandler())
		loadTestRouter.GET("/", handler.GetLoadTestResultHandler())
	}

	agent := router.Group("/agent")
	{
		agent.POST("/install", handler.CreateAgentOnHostHandler())
		agent.POST("/start", handler.StartAgentOnHostHandler())
		agent.POST("/stop", handler.StopAgentOnHostHandler())
		agent.POST("/remove", handler.RemoveAgentOnHostHandler())
	}

	// local database initialize
	domain.InitializeDatabase()

	log.Fatal(router.Run(":8080"))
}
