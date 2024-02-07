package main

import (
	"time"

	"github.com/cloud-barista/cm-ant/api/handler"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// default middleware setting
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// background worker settings
	worker := utils.NewWorker(60 * time.Second)
	go worker.Run()
	defer worker.Shutdown()

	router.GET("/load-test", handler.LoadTestHandler())
	router.Run()

}
