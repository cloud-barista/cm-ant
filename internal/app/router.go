package app

import (
	_ "github.com/cloud-barista/cm-ant/api"

	"github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
)

func (server *AntServer) InitRouter() error {
	setMiddleware(server.e)

	antRouter := server.e.Group("/ant")

	{
		antRouter.GET("/readyz", server.readyz)
		antRouter.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	apiRouter := antRouter.Group("/api")
	versionRouter := apiRouter.Group("/v1")

	{

		loadRouter := versionRouter.Group("/load")

		{
			loadRouter.GET("/generators", server.getAllLoadGeneratorInstallInfo)
			loadRouter.POST("/generators", server.installLoadGenerator)
			loadRouter.DELETE("/generators/:loadGeneratorInstallInfoId", server.uninstallLoadGenerator)

			// load test metrics agent
			loadRouter.POST("/monitoring/agent/install", server.installMonitoringAgent, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))
			loadRouter.GET("/monitoring/agent", server.getAllMonitoringAgentInfos)
			loadRouter.POST("/monitoring/agent/uninstall", server.uninstallMonitoringAgent, middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))

			loadTestRouter := loadRouter.Group("/tests")

			{
				// load test execution
				loadTestRouter.POST("/run", server.runLoadTest)
				loadTestRouter.POST("/stop", server.stopLoadTest)

				// load test state
				loadTestRouter.GET("/state", server.getAllLoadTestExecutionState)
				loadTestRouter.GET("/state/:loadTestKey", server.getLoadTestExecutionState)

				// load test history
				loadTestRouter.GET("/infos", server.getAllLoadTestExecutionInfos)
				loadTestRouter.GET("/infos/:loadTestKey", server.getLoadTestExecutionInfo)

				// load test result
				loadTestRouter.GET("/result", server.getLoadTestResult)
				loadTestRouter.GET("/result/metrics", server.getLoadTestMetrics)
			}
		}
	}

	{
		priceRouter := versionRouter.Group("/price")
		{
			priceRouter.POST("/info", server.updatePriceInfos)
			priceRouter.GET("/info", server.getPriceInfos)
		}

		costRouter := versionRouter.Group("/cost")
		{
			costRouter.POST("/info", server.updateCostInfos)
			costRouter.GET("/info", server.getCostInfos)
		}
	}

	return nil
}

// setStatic sets up static file serving for CSS, JS, and templates.
// func setStatic(e *echo.Echo) {
// 	e.Static("/web/templates", utils.RootPath()+"/web/templates")
// 	e.Static("/css", utils.RootPath()+"/web/css")
// 	e.Static("/js", utils.RootPath()+"/web/js")
// }
