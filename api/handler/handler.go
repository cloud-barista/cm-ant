package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/utils"
	"github.com/gin-gonic/gin"
)

type LoadTestProperties struct {
	Threads  int64  `form:"threads" binding:"omitempty"`
	RampTime int64  `form:"rampTime" binding:"omitempty"`
	Loop     int64  `form:"loop" binding:"omitempty"`
	Hostname string `form:"hostname" bind:"omitempty"`
	Port     int64  `form:"port" bind:"omitempty"`
	Path     string `form:"path" bind:"omitempty"`
}

func NewLoadTestProperties() LoadTestProperties {
	return LoadTestProperties{
		Threads:  1,
		RampTime: 1,
		Loop:     1,
		Hostname: "localhost",
		Port:     8080,
		Path:     "/milkyway/cpus",
	}
}

func LoadTestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		loadTestProperties := NewLoadTestProperties()

		if err := c.ShouldBindQuery(&loadTestProperties); err != nil {
			// handle query bind error
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "request param is bad",
			})
			return
		}

		currentTime := time.Now()
		formattedTime := currentTime.Format("20060102150405")

		filePath := fmt.Sprintf("%s_%s", formattedTime, "config.properties")
		propertiesData := utils.StructToMap(loadTestProperties)

		err := utils.WritePropertiesFile(filePath, propertiesData)
		if err != nil {
			fmt.Println("Error writing properties file:", err)
			return
		}

	}
}
