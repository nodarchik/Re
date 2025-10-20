package models

import "time"

// PackSize represents a pack size configuration
type PackSize struct {
	ID        int       `json:"id" db:"id"`
	Size      int       `json:"size" db:"size"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PackCalculationRequest represents the input for pack calculation
type PackCalculationRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

// PackCalculationResult represents the result of pack calculation
type PackCalculationResult struct {
	Amount     int         `json:"amount"`
	TotalItems int         `json:"total_items"`
	TotalPacks int         `json:"total_packs"`
	Packs      map[int]int `json:"packs"` // map[packSize]quantity
}

// Order represents a saved order calculation
type Order struct {
	ID         int         `json:"id" db:"id"`
	Amount     int         `json:"amount" db:"amount"`
	TotalItems int         `json:"total_items" db:"total_items"`
	TotalPacks int         `json:"total_packs" db:"total_packs"`
	PacksJSON  string      `json:"-" db:"packs_json"` // JSON string for DB storage
	Packs      map[int]int `json:"packs" db:"-"`      // Parsed packs
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}
