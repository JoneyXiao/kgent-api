package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kgent-api/api/config"
	"kgent-api/api/controllers"
	"kgent-api/api/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Kubernetes configuration and clients
	k8sconfig := config.NewK8sConfig().InitRestConfig(
		config.WithQps(100),
		config.WithBurst(200),
		config.WithTimeout(30),
	)
	if err := k8sconfig.Error(); err != nil {
		log.Fatalf("Failed to initialize Kubernetes config: %v", err)
	}

	restMapper := k8sconfig.InitRestMapper()
	dynamicClient := k8sconfig.InitDynamicClient()
	informer := k8sconfig.InitInformer()
	clientSet := k8sconfig.InitClientSet()

	// Initialize services and controllers
	resourceCtl := controllers.NewResourceCtl(
		services.NewResourceService(&restMapper, dynamicClient, informer),
	)
	podLogCtl := controllers.NewPodLogEventCtl(
		services.NewPodLogEventService(clientSet),
	)

	// Setup Gin with middleware
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API versioning with v1 group
	v1 := r.Group("/api/v1")
	{
		// Resource endpoints
		v1.GET("/resources/:resource", resourceCtl.List())
		v1.DELETE("/resources/:resource", resourceCtl.Delete())
		v1.POST("/resources/:resource", resourceCtl.Create())
		v1.GET("/resources/gvr", resourceCtl.GetGVR())

		// Pod logs and events
		v1.GET("/pods/logs", podLogCtl.GetLog())
		v1.GET("/pods/events", podLogCtl.GetEvent())
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Set shutdown timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
