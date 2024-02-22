package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-ant/internal/domain"
	"github.com/cloud-barista/cm-ant/pkg/managers"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

func CreateAgentOnHostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentInfo := domain.NewAgentInfo()
		if err := c.ShouldBindBodyWith(&agentInfo, binding.JSON); err != nil {
			log.Printf("error while binding request body; %+v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "error while binding agent info",
			})
			return
		}

		agentIdUuid := uuid.New()
		agentId := agentIdUuid.String()

		agentInfo.AgentId = agentId

		agentManager := managers.NewAgentManager()
		err := agentManager.Install(agentInfo)

		if err != nil {
			log.Println("error while install agent on host; ", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "internal server error",
				"message": "error while install agent on host",
			})
			return
		}

		agentDB := domain.DBMap["agent"]
		err = agentDB.Insert(agentId, agentInfo)

		if err != nil {
			log.Println("error while insert into db; ", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "internal server error",
				"message": "error while install agent on host",
			})
			return
		}

		c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"agentId": agentId,
		})

	}
}

func StartAgentOnHostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		agentId := c.Query("agentId")

		if len(strings.TrimSpace(agentId)) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "agentId must be passed",
			})
			return
		}

		agentDB := domain.DBMap["agent"]
		item, err := agentDB.FindById(agentId)

		if err != nil {
			log.Println("error while insert into db; ", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "internal server error",
				"message": "error while find agent info",
			})
			return
		}

		if item != nil {
			agentInfo := item.(domain.AgentInfo)

			agentManager := managers.NewAgentManager()
			err = agentManager.Start(agentInfo)

			if err != nil {
				log.Println("error while start agent on host; ", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status":  "internal server error",
					"message": "error while start agent on host",
				})
				return
			}

			c.JSON(http.StatusOK, map[string]string{
				"status": "ok",
				"result": agentId,
			})
			return
		}
		c.JSON(http.StatusNotFound, map[string]string{
			"status":  "not found",
			"message": "can not find item",
		})

	}
}

func StopAgentOnHostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		agentId := c.Query("agentId")

		if len(strings.TrimSpace(agentId)) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "agentId must be passed",
			})
			return
		}

		agentDB := domain.DBMap["agent"]
		item, err := agentDB.FindById(agentId)

		if err != nil {
			log.Println("error while insert into db; ", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "internal server error",
				"message": "error while find agent info",
			})
			return
		}

		if item != nil {
			agentInfo := item.(domain.AgentInfo)

			agentManager := managers.NewAgentManager()
			err = agentManager.Stop(agentInfo)

			if err != nil {
				log.Println("error while stop agent on host; ", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status":  "internal server error",
					"message": "error while stop agent on host",
				})
				return
			}

			c.JSON(http.StatusOK, map[string]string{
				"status": "ok",
				"result": agentId,
			})
			return
		}
		c.JSON(http.StatusNotFound, map[string]string{
			"status":  "not found",
			"message": "can not find item",
		})

	}
}

func RemoveAgentOnHostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		agentId := c.Query("agentId")

		if len(strings.TrimSpace(agentId)) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "bad request",
				"message": "agentId must be passed",
			})
			return
		}

		agentDB := domain.DBMap["agent"]
		item, err := agentDB.FindById(agentId)

		if err != nil {
			log.Println("error while insert into db; ", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "internal server error",
				"message": "error while find agent info",
			})
			return
		}

		if item != nil {
			agentInfo := item.(domain.AgentInfo)

			agentManager := managers.NewAgentManager()
			err = agentManager.Remove(agentInfo)

			if err != nil {
				log.Println("error while remove agent on host; ", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status":  "internal server error",
					"message": "error while remove agent on host",
				})
				return
			}

			agentDB.DeleteById(agentId)

			c.JSON(http.StatusOK, map[string]string{
				"status": "ok",
			})
			return
		}
		c.JSON(http.StatusNotFound, map[string]string{
			"status":  "not found",
			"message": "can not find item",
		})

	}
}
