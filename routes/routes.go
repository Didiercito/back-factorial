package routes

import (
	"examen-back/handlers"
	"examen-back/middleware"
	
	"github.com/gin-gonic/gin"
)

func SetupAnalysisRoutes(router *gin.Engine) {
	analysisHandler := handlers.NewAnalysisHandler()
	
	router.Use(middleware.CORSMiddleware())
	
	api := router.Group("/api/v1")
	{
		api.POST("/analyze", analysisHandler.AnalyzeCode)
		
		api.GET("/health", analysisHandler.GetHealth)
		api.GET("/info", analysisHandler.GetAnalysisInfo)
	}
	
	router.POST("/analyze", analysisHandler.AnalyzeCode)
}

