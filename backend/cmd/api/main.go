package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"pack-calculator/internal/cache"
	"pack-calculator/internal/handlers"
	"pack-calculator/internal/middleware"
	"pack-calculator/internal/repository"
	"strconv"
	"time"
)

func main() {
	// Get environment variables with defaults
	port := getEnv("PORT", "8080")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "packcalculator")

	// Initialize database connection with retry logic
	var db *sql.DB
	var err error
	maxRetries := 30

	log.Println("Connecting to database...")
	for i := 0; i < maxRetries; i++ {
		db, err = repository.InitDB(dbHost, dbPort, dbUser, dbPassword, dbName)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	defer db.Close()

	// Configure connection pool for optimal performance
	db.SetMaxOpenConns(50)                  // Maximum number of open connections (increased for concurrency)
	db.SetMaxIdleConns(10)                  // Maximum number of idle connections (reduced to save memory)
	db.SetConnMaxLifetime(1 * time.Minute)  // Connection lifetime (shorter to avoid stale connections)
	db.SetConnMaxIdleTime(30 * time.Second) // Close unused connections faster

	log.Println("Connected to database successfully")
	log.Printf("Connection pool configured: max_open=50, max_idle=10, lifetime=1m, idle_timeout=30s")

	// Initialize repository
	repo := repository.NewRepository(db)

	// Initialize database schema
	log.Println("Initializing database schema...")
	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Seed default pack sizes
	log.Println("Seeding default pack sizes...")
	if err := repo.SeedDefaultPackSizes(); err != nil {
		log.Fatalf("Failed to seed pack sizes: %v", err)
	}

	// Prepare SQL statements for better performance
	log.Println("Preparing SQL statements...")
	if err := repo.PrepareStatements(); err != nil {
		log.Fatalf("Failed to prepare statements: %v", err)
	}
	log.Println("Prepared statements ready")

	// Initialize cache
	cacheSize := 1000 // Default cache size
	if cacheSizeStr := getEnv("CACHE_SIZE", ""); cacheSizeStr != "" {
		if size, err := strconv.Atoi(cacheSizeStr); err == nil {
			cacheSize = size
		}
	}
	memCache := cache.NewMemoryCache(cacheSize)
	log.Printf("Memory cache initialized with max size: %d", cacheSize)

	// Initialize handlers
	handler := handlers.NewHandler(repo, memCache)

	// Initialize middleware
	// Rate limiter: 100 requests per 10 seconds per IP (burst of 20)
	rateLimiter := middleware.NewRateLimiter(100*time.Millisecond, 20)
	rateLimit := middleware.RateLimitMiddleware(rateLimiter)

	// API key authentication (optional, for write operations on pack sizes)
	apiKey := getEnv("API_KEY", "") // Leave empty for no auth
	apiKeyAuth := middleware.NewAPIKeyAuth(apiKey)

	log.Println("Rate limiting enabled: 100 req/10s per IP")
	if apiKey != "" {
		log.Println("API key authentication enabled for pack size modifications")
	}

	// Compression middleware
	compress := middleware.CompressionMiddleware

	// Setup routes with middleware (compression + rate limiting + CORS)
	http.HandleFunc("/health", handlers.EnableCORS(handler.HealthCheck))

	// Calculator endpoint with compression, rate limiting, and CORS
	http.HandleFunc("/api/calculate", handlers.EnableCORS(compress(rateLimit(handler.CalculatePacks))))

	// Pack sizes endpoint with compression, rate limiting, and optional auth
	http.HandleFunc("/api/packs", handlers.EnableCORS(compress(rateLimit(apiKeyAuth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetPackSizes(w, r)
		case http.MethodPost:
			handler.AddPackSize(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Delete pack size with compression, rate limiting, and optional auth
	http.HandleFunc("/api/packs/", handlers.EnableCORS(compress(rateLimit(apiKeyAuth.AuthMiddleware(handler.DeletePackSize)))))

	// Order history with compression and rate limiting
	http.HandleFunc("/api/orders", handlers.EnableCORS(compress(rateLimit(handler.GetOrders))))

	// Start server
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
