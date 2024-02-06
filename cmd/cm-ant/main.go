package main

import (
	"github.com/cloud-barista/cm-ant/api/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// default middleware setting
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/load-test", handler.LoadTestHandler())
	router.Run()
}
