package calculator

import (
	"testing"
)

func TestCalculator_BasicCases(t *testing.T) {
	tests := []struct {
		name          string
		packSizes     []int
		amount        int
		expectedPacks map[int]int
		expectedTotal int
	}{
		{
			name:          "Single item - exact match",
			packSizes:     []int{250, 500, 1000, 2000, 5000},
			amount:        250,
			expectedPacks: map[int]int{250: 1},
			expectedTotal: 250,
		},
		{
			name:          "One item - smallest pack",
			packSizes:     []int{250, 500, 1000, 2000, 5000},
			amount:        1,
			expectedPacks: map[int]int{250: 1},
			expectedTotal: 250,
		},
		{
			name:          "251 items - use 500 not 2x250",
			packSizes:     []int{250, 500, 1000, 2000, 5000},
			amount:        251,
			expectedPacks: map[int]int{500: 1},
			expectedTotal: 500,
		},
		{
			name:          "501 items - use 1x500 + 1x250",
			packSizes:     []int{250, 500, 1000, 2000, 5000},
			amount:        501,
			expectedPacks: map[int]int{500: 1, 250: 1},
			expectedTotal: 750,
		},
		{
			name:          "12001 items - complex combination",
			packSizes:     []int{250, 500, 1000, 2000, 5000},
			amount:        12001,
			expectedPacks: map[int]int{5000: 2, 2000: 1, 250: 1},
			expectedTotal: 12250,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator(tt.packSizes)
			packs, total, err := calc.Calculate(tt.amount)

			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}

			if total != tt.expectedTotal {
				t.Errorf("Total items = %v, want %v", total, tt.expectedTotal)
			}

			if !mapsEqual(packs, tt.expectedPacks) {
				t.Errorf("Packs = %v, want %v", packs, tt.expectedPacks)
			}
		})
	}
}

func TestCalculator_EdgeCase(t *testing.T) {
	// The critical edge case from requirements
	packSizes := []int{23, 31, 53}
	amount := 500000

	calc := NewCalculator(packSizes)
	packs, total, totalPacks, err := calc.CalculateWithDetails(amount)

	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Expected result: {23: 2, 31: 7, 53: 9429}
	expected := map[int]int{23: 2, 31: 7, 53: 9429}

	if !mapsEqual(packs, expected) {
		t.Errorf("Packs = %v, want %v", packs, expected)
	}

	// Verify the total is correct
	calculatedTotal := 0
	for packSize, count := range packs {
		calculatedTotal += packSize * count
	}

	if calculatedTotal != total {
		t.Errorf("Total mismatch: calculated %v from packs, got %v", calculatedTotal, total)
	}

	// Verify it meets the amount requirement
	if total < amount {
		t.Errorf("Total %v is less than amount %v", total, amount)
	}

	// Verify total packs
	expectedTotalPacks := 2 + 7 + 9429
	if totalPacks != expectedTotalPacks {
		t.Errorf("Total packs = %v, want %v", totalPacks, expectedTotalPacks)
	}

	t.Logf("Edge case passed: amount=%d, total=%d, packs=%v, totalPacks=%d",
		amount, total, packs, totalPacks)
}

func TestCalculator_MinimizeItems(t *testing.T) {
	// Test that we minimize items first (rule 2 takes precedence over rule 3)
	packSizes := []int{250, 500}
	amount := 251

	calc := NewCalculator(packSizes)
	packs, total, err := calc.Calculate(amount)

	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Should use 1x500 (500 items total, 1 pack)
	// Not 2x250 (500 items total, 2 packs) - same items but more packs
	// Not 1x250 (250 items total, 1 pack) - fewer items but doesn't meet requirement
	expected := map[int]int{500: 1}

	if !mapsEqual(packs, expected) {
		t.Errorf("Packs = %v, want %v (should minimize items first)", packs, expected)
	}

	if total != 500 {
		t.Errorf("Total = %v, want 500", total)
	}
}

func TestCalculator_MinimizePacks(t *testing.T) {
	// When same total items, minimize packs (rule 3)
	packSizes := []int{250, 500}
	amount := 500

	calc := NewCalculator(packSizes)
	packs, total, err := calc.Calculate(amount)

	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Should use 1x500 not 2x250 (both give 500 items, but 1 pack is fewer)
	expected := map[int]int{500: 1}

	if !mapsEqual(packs, expected) {
		t.Errorf("Packs = %v, want %v (should minimize packs when items are equal)", packs, expected)
	}

	if total != 500 {
		t.Errorf("Total = %v, want 500", total)
	}
}

func TestCalculator_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		packSizes []int
		amount    int
		wantError bool
	}{
		{
			name:      "Zero amount",
			packSizes: []int{250, 500},
			amount:    0,
			wantError: true,
		},
		{
			name:      "Negative amount",
			packSizes: []int{250, 500},
			amount:    -100,
			wantError: true,
		},
		{
			name:      "Empty pack sizes",
			packSizes: []int{},
			amount:    100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator(tt.packSizes)
			_, _, err := calc.Calculate(tt.amount)

			if (err != nil) != tt.wantError {
				t.Errorf("Calculate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCalculator_LargeNumbers(t *testing.T) {
	packSizes := []int{1000, 5000, 10000}
	amount := 1000000

	calc := NewCalculator(packSizes)
	packs, total, err := calc.Calculate(amount)

	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Verify it meets the requirement
	if total < amount {
		t.Errorf("Total %v is less than amount %v", total, amount)
	}

	// Verify the calculation is correct
	calculatedTotal := 0
	for packSize, count := range packs {
		calculatedTotal += packSize * count
	}

	if calculatedTotal != total {
		t.Errorf("Total mismatch: calculated %v from packs, got %v", calculatedTotal, total)
	}

	t.Logf("Large number test passed: amount=%d, total=%d, packs=%v", amount, total, packs)
}

// Helper function to compare maps
func mapsEqual(a, b map[int]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
