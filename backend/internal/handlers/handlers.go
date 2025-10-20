package handlers

import (
	"fmt"
	"net/http"
	"pack-calculator/internal/cache"
	"pack-calculator/internal/calculator"
	"pack-calculator/internal/models"
	"pack-calculator/internal/repository"
	"strconv"
	"strings"
	"time"

	json "github.com/goccy/go-json"
)

// Handler manages HTTP requests
type Handler struct {
	repo  *repository.Repository
	cache cache.Cache
}

// NewHandler creates a new handler instance
func NewHandler(repo *repository.Repository, cacheImpl cache.Cache) *Handler {
	if cacheImpl == nil {
		cacheImpl = &cache.NoOpCache{} // Default to no cache
	}
	return &Handler{
		repo:  repo,
		cache: cacheImpl,
	}
}

// EnableCORS middleware to allow cross-origin requests
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// CalculatePacks handles POST /api/calculate
func (h *Handler) CalculatePacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req models.PackCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate amount
	if req.Amount < 1 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Amount must be at least 1"})
		return
	}

	// Set reasonable upper limit to prevent memory exhaustion
	const maxAmount = 10000000 // 10 million items max
	if req.Amount > maxAmount {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Amount too large. Maximum allowed: %d items", maxAmount),
		})
		return
	}

	// Get pack sizes from database
	packSizes, err := h.repo.GetPackSizesAsSlice()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get pack sizes"})
		return
	}

	if len(packSizes) == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "No pack sizes configured"})
		return
	}

	// Check cache first
	cacheKey := cache.GenerateCacheKey(req.Amount, packSizes)
	if cachedPacks, cachedTotal, found := h.cache.Get(cacheKey); found {
		// Calculate total packs from cached data
		totalPacks := 0
		for _, count := range cachedPacks {
			totalPacks += count
		}

		result := models.PackCalculationResult{
			Amount:     req.Amount,
			TotalItems: cachedTotal,
			TotalPacks: totalPacks,
			Packs:      cachedPacks,
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	// Calculate optimal packs
	calc := calculator.NewCalculator(packSizes)
	packs, totalItems, totalPacks, err := calc.CalculateWithDetails(req.Amount)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Cache the result (TTL: 1 hour)
	h.cache.Set(cacheKey, packs, totalItems, 1*time.Hour)

	// Create result
	result := models.PackCalculationResult{
		Amount:     req.Amount,
		TotalItems: totalItems,
		TotalPacks: totalPacks,
		Packs:      packs,
	}

	// Save order to database
	order := &models.Order{
		Amount:     req.Amount,
		TotalItems: totalItems,
		TotalPacks: totalPacks,
		Packs:      packs,
	}

	if err := h.repo.SaveOrder(order); err != nil {
		// Log error but don't fail the request
		// The calculation is still valid even if we can't save it
	}

	respondJSON(w, http.StatusOK, result)
}

// GetPackSizes handles GET /api/packs
func (h *Handler) GetPackSizes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	packSizes, err := h.repo.GetAllPackSizes()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get pack sizes"})
		return
	}

	respondJSON(w, http.StatusOK, packSizes)
}

// AddPackSize handles POST /api/packs
func (h *Handler) AddPackSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Size int `json:"size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Size < 1 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Size must be at least 1"})
		return
	}

	// Check if pack size already exists
	exists, err := h.repo.PackSizeExists(req.Size)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to check pack size"})
		return
	}

	if exists {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "Pack size already exists"})
		return
	}

	if err := h.repo.AddPackSize(req.Size); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to add pack size"})
		return
	}

	// Clear cache when pack sizes change
	h.cache.Clear()

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Pack size added successfully"})
}

// DeletePackSize handles DELETE /api/packs/{size}
func (h *Handler) DeletePackSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract size from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	sizeStr := parts[len(parts)-1]
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid size"})
		return
	}

	if err := h.repo.DeletePackSize(size); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	// Clear cache when pack sizes change
	h.cache.Clear()

	respondJSON(w, http.StatusOK, map[string]string{"message": "Pack size deleted successfully"})
}

// GetOrders handles GET /api/orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get limit from query param, default to 100
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	orders, err := h.repo.GetAllOrders(limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get orders"})
		return
	}

	respondJSON(w, http.StatusOK, orders)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	stats := h.cache.Stats()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"cache": map[string]interface{}{
			"hits":      stats.Hits,
			"misses":    stats.Misses,
			"hit_ratio": stats.HitRatio,
			"size":      stats.Size,
		},
	})
}

// respondJSON writes a buffered JSON response for better performance
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Let Go handle the response encoding and Content-Length automatically
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Error encoding response: %v\n", err)
		return
	}
}
