package main

import (
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/api/handler"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
)

func main() {
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
	router.POST("/load-test", handler.LoadTestHandler())
	router.POST("/slave", handler.SlaveCreateHandler())
	router.Run()

}
