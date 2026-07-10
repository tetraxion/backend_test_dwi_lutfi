package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"task-tracker-backend/db"
	"task-tracker-backend/handler"
	"task-tracker-backend/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	// Pilih storage backend berdasarkan env USE_POSTGRES
	var repo handler.TaskRepo
	if os.Getenv("USE_POSTGRES") != "" {
		pool, err := db.Connect(ctx)
		if err != nil {
			log.Fatalf("database connection failed: %v", err)
		}
		defer pool.Close()

		if err := db.Migrate(ctx, pool); err != nil {
			log.Fatalf("migration failed: %v", err)
		}

		repo = repository.NewPostgresTaskRepository(pool)
		log.Println("🐘  Using PostgreSQL storage")
	} else {
		repo = repository.NewTaskRepository()
		log.Println("💾  Using in-memory storage (set USE_POSTGRES=true to use PostgreSQL)")
	}

	// Setup handler
	taskHandler := handler.NewTaskHandler(repo)

	// Router Gin
	router := gin.Default()

	// Konfigurasi CORS
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Route group /api/v1
	v1 := router.Group("/api/v1")
	taskHandler.RegisterRoutes(v1)

	// Endpoint health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	fmt.Printf("🚀  Task Tracker API listening on http://0.0.0.0:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
