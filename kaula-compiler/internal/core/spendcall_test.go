package core

import (
	"sync"
	"testing"
)

func TestSpendable_BasicLifecycle(t *testing.T) {
	sp := NewSpendable(3)

	if !sp.IsConsumed() {
		t.Error("Empty spendable should be consumed")
	}
	if sp.GetRemaining() != 0 {
		t.Errorf("Empty spendable should have 0 remaining, got %d", sp.GetRemaining())
	}

	sp.Add("component1")
	sp.Add("component2")

	if sp.IsConsumed() {
		t.Error("Spendable with components should not be consumed")
	}
	if sp.GetRemaining() != 2 {
		t.Errorf("Should have 2 remaining, got %d", sp.GetRemaining())
	}

	// First call
	c1 := sp.Call()
	if c1 != "component1" {
		t.Errorf("First call should return component1, got %v", c1)
	}
	if sp.GetRemaining() != 1 {
		t.Errorf("Should have 1 remaining, got %d", sp.GetRemaining())
	}

	// Second call
	c2 := sp.Call()
	if c2 != "component2" {
		t.Errorf("Second call should return component2, got %v", c2)
	}
	if !sp.IsConsumed() {
		t.Error("Should be consumed after all components used")
	}
	if sp.GetRemaining() != 0 {
		t.Errorf("Should have 0 remaining, got %d", sp.GetRemaining())
	}

	// Third call after consumed - should return nil
	c3 := sp.Call()
	if c3 != nil {
		t.Errorf("Call after consumed should return nil, got %v", c3)
	}
}

func TestSpendable_ConcurrentAccess(t *testing.T) {
	sp := NewSpendable(100)
	for i := 0; i < 100; i++ {
		sp.Add(i)
	}

	var wg sync.WaitGroup
	results := make([]interface{}, 100)

	// Launch 100 goroutines to consume components concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = sp.Call()
		}(i)
	}

	wg.Wait()

	// Verify all components were consumed
	if !sp.IsConsumed() {
		t.Error("Should be consumed after all concurrent calls")
	}
	if sp.GetRemaining() != 0 {
		t.Errorf("Should have 0 remaining, got %d", sp.GetRemaining())
	}

	// Count non-nil results
	nonNilCount := 0
	for _, r := range results {
		if r != nil {
			nonNilCount++
		}
	}

	// All 100 calls should have returned a component
	if nonNilCount != 100 {
		t.Errorf("Expected 100 non-nil results, got %d", nonNilCount)
	}
}

func TestSpendable_AutoRelease(t *testing.T) {
	sp := NewSpendable(2)
	sp.Add("a")
	sp.Add("b")

	// Consume all components
	sp.Call()
	sp.Call()

	// After consumption, internal slice should be nil
	sp.Mutex.RLock()
	if sp.Components != nil {
		t.Error("Components should be nil after auto-release")
	}
	if sp.Count != 0 {
		t.Errorf("Count should be 0 after auto-release, got %d", sp.Count)
	}
	if sp.CallCounter != 0 {
		t.Errorf("CallCounter should be 0 after auto-release, got %d", sp.CallCounter)
	}
	sp.Mutex.RUnlock()
}

func TestSpendable_EmptyCalls(t *testing.T) {
	sp := NewSpendable(0)

	// Call on empty spendable
	result := sp.Call()
	if result != nil {
		t.Errorf("Call on empty spendable should return nil, got %v", result)
	}

	// Should be consumed
	if !sp.IsConsumed() {
		t.Error("Empty spendable should be consumed")
	}
}
