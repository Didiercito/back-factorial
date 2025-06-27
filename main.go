package main

import (
	"log"
	"os"
	"examen-back/config"
	"examen-back/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := config.InitDB(); err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer config.CloseDB()

	gin.SetMode(getEnv("GIN_MODE", "debug"))
	r := gin.Default()

	routes.SetupAnalysisRoutes(r)

	port := getEnv("PORT", "8080")
	log.Printf("Server running on port %s", port)
	log.Printf("Endpoints disponibles:")
	log.Printf("  POST /analyze - Análisis de código")
	log.Printf("  GET  /api/v1/health - Estado del servicio")
	log.Printf("  GET  /api/v1/info - Información del analizador")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}