package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cloud-barista/cm-ant/pkg/antcontainer"
	"github.com/gin-gonic/gin"
)

func SlaveCreateHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		containerManager := antcontainer.CManager
		err := containerManager.BuildJMeterDockerImage()
		if err != nil {
			log.Printf("Error while creating container; %v\n", err)
			c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "internal server error",
				"message": fmt.Sprintf("Error while creating container; %v", err),
			})

			return
		}
	}
}
