package calculator

import (
	"errors"
	"math"
	"sort"
)

// Calculator handles pack size calculations using dynamic programming
type Calculator struct {
	packSizes []int
}

// NewCalculator creates a new calculator with given pack sizes
func NewCalculator(packSizes []int) *Calculator {
	// Sort pack sizes for consistent processing
	sorted := make([]int, len(packSizes))
	copy(sorted, packSizes)
	sort.Ints(sorted)
	return &Calculator{packSizes: sorted}
}

// Calculate finds the optimal pack combination for a given amount
// Rule 1: Only whole packs (no breaking)
// Rule 2: Minimize total items sent (takes precedence)
// Rule 3: Among solutions with same item count, minimize number of packs
func (c *Calculator) Calculate(amount int) (map[int]int, int, error) {
	if amount <= 0 {
		return nil, 0, errors.New("amount must be positive")
	}
	if len(c.packSizes) == 0 {
		return nil, 0, errors.New("no pack sizes available")
	}

	// Find the maximum target we need to check
	// We need to find the smallest combination that meets or exceeds 'amount'
	// The worst case is using all smallest packs, but we limit search space
	maxTarget := amount + c.packSizes[len(c.packSizes)-1]

	// dp[i] stores the minimum number of packs to achieve exactly i items
	// Initialize with max value (impossible state)
	dp := make([]int, maxTarget+1)
	for i := range dp {
		dp[i] = math.MaxInt32
	}
	dp[0] = 0 // Base case: 0 items needs 0 packs

	// parent[i] stores which pack size was used to reach state i
	parent := make([]int, maxTarget+1)

	// Dynamic programming: build up solutions for all amounts up to maxTarget
	for i := 0; i <= maxTarget; i++ {
		if dp[i] == math.MaxInt32 {
			continue // Can't reach this state
		}

		// Try adding each pack size
		for _, packSize := range c.packSizes {
			next := i + packSize
			if next <= maxTarget {
				// Update if this gives fewer packs for the same total
				if dp[next] > dp[i]+1 {
					dp[next] = dp[i] + 1
					parent[next] = packSize
				}
			}
		}
	}

	// Find the minimum total items >= amount with a valid solution
	bestTotal := -1
	for i := amount; i <= maxTarget; i++ {
		if dp[i] != math.MaxInt32 {
			bestTotal = i
			break
		}
	}

	if bestTotal == -1 {
		return nil, 0, errors.New("no valid pack combination found")
	}

	// Backtrack to find which packs were used
	packs := make(map[int]int)
	current := bestTotal
	for current > 0 {
		packUsed := parent[current]
		packs[packUsed]++
		current -= packUsed
	}

	return packs, bestTotal, nil
}

// CalculateWithDetails returns detailed results including total packs
func (c *Calculator) CalculateWithDetails(amount int) (map[int]int, int, int, error) {
	packs, totalItems, err := c.Calculate(amount)
	if err != nil {
		return nil, 0, 0, err
	}

	// Count total number of packs
	totalPacks := 0
	for _, count := range packs {
		totalPacks += count
	}

	return packs, totalItems, totalPacks, nil
}
