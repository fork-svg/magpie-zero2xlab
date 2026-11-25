package main

import (
	"log"
	"os"

	"github.com/fork-svg/magpie-zero2xlab/config"
	"github.com/fork-svg/magpie-zero2xlab/internal/api"
	"github.com/fork-svg/magpie-zero2xlab/internal/auth"
	"github.com/fork-svg/magpie-zero2xlab/internal/db"
	"github.com/fork-svg/magpie-zero2xlab/internal/proxy"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Check required environment variables
	if cfg.Zero2XLabAPIKey == "" {
		log.Println("WARNING: ZERO2XLAB_API_KEY not set, proxy will not work")
	}

	// Initialize JWT
	auth.Init(cfg.JWTSecret)

	// Initialize database
	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize proxy
	proxyHandler := proxy.NewProxy(cfg)

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	apiGroup := router.Group("/api")
	{
		// Public routes
		apiGroup.POST("/register", api.Register)
		apiGroup.POST("/login", api.Login)

		// Protected routes (JWT auth)
		protected := apiGroup.Group("")
		protected.Use(api.AuthMiddleware())
		{
			protected.GET("/user", api.GetCurrentUser)
			protected.GET("/api-keys", api.ListAPIKeys)
			protected.POST("/api-keys", api.CreateAPIKey)
			protected.DELETE("/api-keys/:key", api.DeleteAPIKey)
			protected.GET("/logs", api.GetRequestLogs)
		}

		// Gateway routes (API key auth) - OpenAI compatible
		gateway := apiGroup.Group("/v1")
		gateway.Use(api.APIKeyAuthMiddleware())
		{
			// Chat completions
			gateway.POST("/chat/completions", func(c *gin.Context) {
				apiKeyID := c.GetUint("api_key_id")
				if err := proxyHandler.ForwardRequest(c.Writer, c.Request, apiKeyID); err != nil {
					log.Printf("Proxy error: %v", err)
					c.JSON(502, gin.H{"error": err.Error()})
				}
			})

			// Models list (mock)
			gateway.GET("/models", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"object": "list",
					"data": []gin.H{
						{"id": "oneastronomy-r1", "object": "model", "owned_by": "zero2xlab"},
						{"id": "oneastronomy-chat", "object": "model", "owned_by": "zero2xlab"},
					},
				})
			})
		}
	}

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
