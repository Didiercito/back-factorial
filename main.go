package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"examen-back/config"
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

	port := getEnv("PORT", "8080")
	log.Printf("Server running on port %s", port)
	
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