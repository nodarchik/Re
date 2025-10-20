package tests

import (
	"fmt"
	"pack-calculator/internal/cache"
	"pack-calculator/internal/calculator"
	"pack-calculator/internal/models"
	"sync"
	"testing"
	"time"
)

// MockRepository for testing without database
type MockRepository struct {
	packSizes []int
}

func (m *MockRepository) GetPackSizesAsSlice() ([]int, error) {
	return m.packSizes, nil
}

func (m *MockRepository) GetAllPackSizes() ([]models.PackSize, error) {
	packs := make([]models.PackSize, len(m.packSizes))
	for i, size := range m.packSizes {
		packs[i] = models.PackSize{ID: i + 1, Size: size}
	}
	return packs, nil
}

func (m *MockRepository) AddPackSize(size int) error {
	m.packSizes = append(m.packSizes, size)
	return nil
}

func (m *MockRepository) DeletePackSize(size int) error {
	return nil
}

func (m *MockRepository) PackSizeExists(size int) (bool, error) {
	return false, nil
}

func (m *MockRepository) SaveOrder(order *models.Order) error {
	return nil
}

func (m *MockRepository) GetAllOrders(limit int) ([]models.Order, error) {
	return []models.Order{}, nil
}

func NewMockRepository(packSizes []int) *MockRepository {
	return &MockRepository{packSizes: packSizes}
}

// TestStressConcurrentRequests tests handling of many concurrent requests
func TestStressConcurrentRequests(t *testing.T) {
	// Setup - test directly with calculator, not handlers
	// This avoids HTTP overhead and tests pure algorithm performance
	packSizes := []int{250, 500, 1000, 2000, 5000}

	numRequests := 1000
	concurrency := 50

	t.Logf("Running stress test: %d calculations with %d concurrent workers", numRequests, concurrency)

	var wg sync.WaitGroup
	errors := make(chan error, numRequests)
	successCount := make(chan int, numRequests)

	startTime := time.Now()

	// Create worker pool
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(requestNum int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Vary the amounts to test different scenarios
			amount := 251 + (requestNum % 10000)

			calc := calculator.NewCalculator(packSizes)
			_, _, err := calc.Calculate(amount)

			if err == nil {
				successCount <- 1
			} else {
				errors <- fmt.Errorf("calculation %d failed: %v", requestNum, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	close(successCount)

	duration := time.Since(startTime)

	// Count results
	errorCount := 0
	for range errors {
		errorCount++
	}

	success := 0
	for range successCount {
		success++
	}

	t.Logf("Completed in %v", duration)
	t.Logf("Success: %d, Errors: %d", success, errorCount)
	t.Logf("Throughput: %.2f calculations/second", float64(numRequests)/duration.Seconds())

	if errorCount > 0 {
		t.Errorf("Had %d errors out of %d requests", errorCount, numRequests)
	}

	// Expect good throughput
	minThroughput := 100.0 // calculations per second
	actualThroughput := float64(numRequests) / duration.Seconds()
	if actualThroughput < minThroughput {
		t.Errorf("Throughput too low: %.2f calc/s (expected > %.2f)", actualThroughput, minThroughput)
	}
}

// TestEdgeCasesComprehensive tests all popular edge cases
func TestEdgeCasesComprehensive(t *testing.T) {
	tests := []struct {
		name          string
		packSizes     []int
		amount        int
		expectError   bool
		validatePacks func(packs map[int]int, total int) bool
	}{
		{
			name:        "Exact match - single pack",
			packSizes:   []int{250, 500, 1000},
			amount:      500,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total == 500 && len(packs) == 1 && packs[500] == 1
			},
		},
		{
			name:        "One item - smallest pack",
			packSizes:   []int{250, 500, 1000},
			amount:      1,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total == 250 && packs[250] == 1
			},
		},
		{
			name:        "Just over pack size - use next pack",
			packSizes:   []int{250, 500, 1000},
			amount:      251,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				// Should be 1×500, not 2×250
				return total == 500 && len(packs) == 1
			},
		},
		{
			name:        "Prime number pack sizes",
			packSizes:   []int{23, 31, 53},
			amount:      100,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total >= 100
			},
		},
		{
			name:        "Large prime numbers",
			packSizes:   []int{97, 103, 107},
			amount:      1000,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total >= 1000
			},
		},
		{
			name:        "Single pack size",
			packSizes:   []int{100},
			amount:      350,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total == 400 && packs[100] == 4
			},
		},
		{
			name:        "Many small packs vs one large",
			packSizes:   []int{1, 1000},
			amount:      999,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				// Should prefer 1×1000 over 999×1
				return total == 1000 && packs[1000] == 1
			},
		},
		{
			name:        "Fibonacci-like sizes",
			packSizes:   []int{1, 2, 3, 5, 8, 13, 21},
			amount:      50,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total >= 50
			},
		},
		{
			name:        "Powers of 2",
			packSizes:   []int{1, 2, 4, 8, 16, 32, 64, 128},
			amount:      100,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total >= 100
			},
		},
		{
			name:        "Maximum allowed amount",
			packSizes:   []int{1000, 5000, 10000},
			amount:      10000000,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total >= 10000000
			},
		},
		{
			name:        "Coprime pack sizes",
			packSizes:   []int{7, 11, 13},
			amount:      100,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total >= 100
			},
		},
		{
			name:        "Large gaps between sizes",
			packSizes:   []int{1, 1000000},
			amount:      500000,
			expectError: false,
			validatePacks: func(packs map[int]int, total int) bool {
				return total == 500000 || total == 1000000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := calculator.NewCalculator(tt.packSizes)
			packs, total, err := calc.Calculate(tt.amount)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify total meets requirement
			if total < tt.amount {
				t.Errorf("Total %d is less than requested amount %d", total, tt.amount)
			}

			// Verify calculation is correct
			calculatedTotal := 0
			for packSize, count := range packs {
				calculatedTotal += packSize * count
			}

			if calculatedTotal != total {
				t.Errorf("Pack calculation mismatch: sum=%d, reported=%d", calculatedTotal, total)
			}

			// Custom validation
			if tt.validatePacks != nil && !tt.validatePacks(packs, total) {
				t.Errorf("Custom validation failed for packs=%v, total=%d", packs, total)
			}

			t.Logf("Amount: %d, Total: %d, Packs: %v", tt.amount, total, packs)
		})
	}
}

// TestMemoryStress tests memory usage under load
func TestMemoryStress(t *testing.T) {
	t.Log("Testing memory stress with large cache")

	memCache := cache.NewMemoryCache(10000) // Large cache

	packSizes := []int{250, 500, 1000, 2000, 5000}

	// Fill cache with many different calculations
	for i := 1; i <= 10000; i++ {
		calc := calculator.NewCalculator(packSizes)
		packs, total, err := calc.Calculate(i)

		if err != nil {
			t.Fatalf("Calculation failed for amount %d: %v", i, err)
		}

		// Add to cache
		key := cache.GenerateCacheKey(i, packSizes)
		memCache.Set(key, packs, total, 1*time.Hour)
	}

	stats := memCache.Stats()
	t.Logf("Cache filled with %d items", stats.Size)

	// Verify cache size limit is respected
	if stats.Size > 10000 {
		t.Errorf("Cache exceeded max size: %d > 10000", stats.Size)
	}

	// Test cache hits
	hits := 0
	for i := 1; i <= 1000; i++ {
		key := cache.GenerateCacheKey(i, packSizes)
		if _, _, found := memCache.Get(key); found {
			hits++
		}
	}

	t.Logf("Cache hit rate for first 1000 items: %d%%", hits/10)
}

// TestAlgorithmPerformance tests performance across different input sizes
func TestAlgorithmPerformance(t *testing.T) {
	packSizes := []int{250, 500, 1000, 2000, 5000}
	calc := calculator.NewCalculator(packSizes)

	testSizes := []int{
		10,
		100,
		1000,
		10000,
		100000,
		1000000,
		5000000,
	}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Amount_%d", size), func(t *testing.T) {
			start := time.Now()
			packs, total, err := calc.Calculate(size)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Calculation failed: %v", err)
			}

			if total < size {
				t.Errorf("Total %d < amount %d", total, size)
			}

			t.Logf("Amount: %d, Total: %d, Packs: %v, Time: %v",
				size, total, packs, duration)

			// Performance expectations
			maxDuration := time.Second * 3
			if duration > maxDuration {
				t.Errorf("Calculation too slow: %v (expected < %v)", duration, maxDuration)
			}
		})
	}
}

// TestCacheEfficiency tests cache hit ratio improvement
func TestCacheEfficiency(t *testing.T) {
	memCache := cache.NewMemoryCache(1000)
	packSizes := []int{250, 500, 1000, 2000, 5000}

	// Simulate repeated calculations (realistic usage)
	commonAmounts := []int{100, 250, 500, 750, 1000, 1500, 2000}

	// First pass - all misses
	for i := 0; i < 100; i++ {
		amount := commonAmounts[i%len(commonAmounts)]
		key := cache.GenerateCacheKey(amount, packSizes)

		// Check cache (miss expected first time)
		if _, _, found := memCache.Get(key); !found {
			// Calculate and cache
			calc := calculator.NewCalculator(packSizes)
			packs, total, _ := calc.Calculate(amount)
			memCache.Set(key, packs, total, 1*time.Hour)
		}
	}

	stats := memCache.Stats()
	t.Logf("After first pass - Hits: %d, Misses: %d, Ratio: %.2f%%",
		stats.Hits, stats.Misses, stats.HitRatio*100)

	// Second pass - expect high hit ratio
	for i := 0; i < 100; i++ {
		amount := commonAmounts[i%len(commonAmounts)]
		key := cache.GenerateCacheKey(amount, packSizes)
		memCache.Get(key)
	}

	stats = memCache.Stats()
	t.Logf("After second pass - Hits: %d, Misses: %d, Ratio: %.2f%%",
		stats.Hits, stats.Misses, stats.HitRatio*100)

	// Expect >80% hit ratio after warmup
	if stats.HitRatio < 0.8 {
		t.Errorf("Cache hit ratio too low: %.2f%% (expected >80%%)", stats.HitRatio*100)
	}
}

// TestBoundaryConditions tests edge values
func TestBoundaryConditions(t *testing.T) {
	packSizes := []int{250, 500, 1000, 2000, 5000}
	calc := calculator.NewCalculator(packSizes)

	tests := []struct {
		name   string
		amount int
		valid  bool
	}{
		{"Minimum valid", 1, true},
		{"Zero", 0, false},
		{"Negative", -1, false},
		{"Max int32", 2147483647, true}, // Will be slow but should work
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := calc.Calculate(tt.amount)

			if tt.valid && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}

			if !tt.valid && err == nil {
				t.Errorf("Expected error but got success")
			}
		})
	}
}

// BenchmarkCalculation benchmarks the calculation performance
func BenchmarkCalculation(b *testing.B) {
	packSizes := []int{250, 500, 1000, 2000, 5000}
	calc := calculator.NewCalculator(packSizes)

	amounts := []int{100, 1000, 10000, 100000}

	for _, amount := range amounts {
		b.Run(fmt.Sprintf("Amount_%d", amount), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				calc.Calculate(amount)
			}
		})
	}
}

// BenchmarkConcurrent benchmarks concurrent calculations
func BenchmarkConcurrent(b *testing.B) {
	packSizes := []int{250, 500, 1000, 2000, 5000}

	b.RunParallel(func(pb *testing.PB) {
		calc := calculator.NewCalculator(packSizes)
		amount := 1000

		for pb.Next() {
			calc.Calculate(amount)
		}
	})
}
