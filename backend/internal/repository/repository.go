package repository

import (
	"database/sql"
	"fmt"
	"pack-calculator/internal/models"
	"time"

	json "github.com/goccy/go-json"
	_ "github.com/lib/pq"
)

// Repository handles database operations
type Repository struct {
	db                 *sql.DB
	getPackSizesStmt   *sql.Stmt
	addPackSizeStmt    *sql.Stmt
	deletePackSizeStmt *sql.Stmt
	saveOrderStmt      *sql.Stmt
	getOrdersStmt      *sql.Stmt
}

// NewRepository creates a new repository instance with prepared statements
func NewRepository(db *sql.DB) *Repository {
	repo := &Repository{db: db}

	// Prepare statements (will be initialized after schema is created)
	return repo
}

// PrepareStatements prepares SQL statements for better performance
func (r *Repository) PrepareStatements() error {
	var err error

	// Prepare get pack sizes statement
	r.getPackSizesStmt, err = r.db.Prepare(`SELECT id, size, created_at FROM pack_sizes ORDER BY size ASC`)
	if err != nil {
		return fmt.Errorf("failed to prepare get pack sizes statement: %w", err)
	}

	// Prepare add pack size statement
	r.addPackSizeStmt, err = r.db.Prepare(`INSERT INTO pack_sizes (size, created_at) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("failed to prepare add pack size statement: %w", err)
	}

	// Prepare delete pack size statement
	r.deletePackSizeStmt, err = r.db.Prepare(`DELETE FROM pack_sizes WHERE size = $1`)
	if err != nil {
		return fmt.Errorf("failed to prepare delete pack size statement: %w", err)
	}

	// Prepare save order statement
	r.saveOrderStmt, err = r.db.Prepare(`INSERT INTO orders (amount, total_items, total_packs, packs_json, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`)
	if err != nil {
		return fmt.Errorf("failed to prepare save order statement: %w", err)
	}

	// Prepare get orders statement
	r.getOrdersStmt, err = r.db.Prepare(`SELECT id, amount, total_items, total_packs, packs_json, created_at FROM orders ORDER BY created_at DESC LIMIT $1`)
	if err != nil {
		return fmt.Errorf("failed to prepare get orders statement: %w", err)
	}

	return nil
}

// InitDB initializes the database connection
func InitDB(host, port, user, password, dbname string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// InitSchema creates the necessary database tables
func (r *Repository) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS pack_sizes (
			id SERIAL PRIMARY KEY,
			size INTEGER NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			amount INTEGER NOT NULL,
			total_items INTEGER NOT NULL,
			total_packs INTEGER NOT NULL,
			packs_json TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pack_sizes_size ON pack_sizes(size)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// PackSize operations

// GetAllPackSizes retrieves all pack sizes from the database
func (r *Repository) GetAllPackSizes() ([]models.PackSize, error) {
	// Use prepared statement if available, otherwise use direct query
	var rows *sql.Rows
	var err error

	if r.getPackSizesStmt != nil {
		rows, err = r.getPackSizesStmt.Query()
	} else {
		rows, err = r.db.Query(`SELECT id, size, created_at FROM pack_sizes ORDER BY size ASC`)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query pack sizes: %w", err)
	}
	defer rows.Close()

	var packSizes []models.PackSize
	for rows.Next() {
		var ps models.PackSize
		if err := rows.Scan(&ps.ID, &ps.Size, &ps.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan pack size: %w", err)
		}
		packSizes = append(packSizes, ps)
	}

	return packSizes, nil
}

// GetPackSizesAsSlice returns pack sizes as a slice of integers
func (r *Repository) GetPackSizesAsSlice() ([]int, error) {
	packSizes, err := r.GetAllPackSizes()
	if err != nil {
		return nil, err
	}

	sizes := make([]int, len(packSizes))
	for i, ps := range packSizes {
		sizes[i] = ps.Size
	}

	return sizes, nil
}

// AddPackSize adds a new pack size to the database
func (r *Repository) AddPackSize(size int) error {
	var err error
	if r.addPackSizeStmt != nil {
		_, err = r.addPackSizeStmt.Exec(size, time.Now())
	} else {
		_, err = r.db.Exec(`INSERT INTO pack_sizes (size, created_at) VALUES ($1, $2)`, size, time.Now())
	}
	if err != nil {
		return fmt.Errorf("failed to add pack size: %w", err)
	}
	return nil
}

// DeletePackSize removes a pack size from the database
func (r *Repository) DeletePackSize(size int) error {
	query := `DELETE FROM pack_sizes WHERE size = $1`
	result, err := r.db.Exec(query, size)
	if err != nil {
		return fmt.Errorf("failed to delete pack size: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("pack size %d not found", size)
	}

	return nil
}

// PackSizeExists checks if a pack size exists
func (r *Repository) PackSizeExists(size int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pack_sizes WHERE size = $1)`
	var exists bool
	err := r.db.QueryRow(query, size).Scan(&exists)
	return exists, err
}

// Order operations

// SaveOrder saves an order calculation to the database
func (r *Repository) SaveOrder(order *models.Order) error {
	// Convert packs map to JSON
	packsJSON, err := json.Marshal(order.Packs)
	if err != nil {
		return fmt.Errorf("failed to marshal packs: %w", err)
	}

	query := `INSERT INTO orders (amount, total_items, total_packs, packs_json, created_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err = r.db.QueryRow(query,
		order.Amount,
		order.TotalItems,
		order.TotalPacks,
		string(packsJSON),
		time.Now(),
	).Scan(&order.ID)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

// GetAllOrders retrieves all orders from the database
func (r *Repository) GetAllOrders(limit int) ([]models.Order, error) {
	query := `SELECT id, amount, total_items, total_packs, packs_json, created_at 
			  FROM orders ORDER BY created_at DESC LIMIT $1`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.Amount,
			&order.TotalItems,
			&order.TotalPacks,
			&order.PacksJSON,
			&order.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Parse the JSON packs
		if err := json.Unmarshal([]byte(order.PacksJSON), &order.Packs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal packs: %w", err)
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// SeedDefaultPackSizes adds default pack sizes if the table is empty
func (r *Repository) SeedDefaultPackSizes() error {
	// Check if pack sizes already exist
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM pack_sizes`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count pack sizes: %w", err)
	}

	// If pack sizes exist, don't seed
	if count > 0 {
		return nil
	}

	// Default pack sizes from the problem statement
	defaultSizes := []int{250, 500, 1000, 2000, 5000}

	for _, size := range defaultSizes {
		if err := r.AddPackSize(size); err != nil {
			return fmt.Errorf("failed to seed pack size %d: %w", size, err)
		}
	}

	return nil
}
