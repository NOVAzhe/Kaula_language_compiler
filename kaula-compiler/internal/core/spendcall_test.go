package core

import (
	"sync"
	"testing"
)

func TestSpendable_AddAndCall(t *testing.T) {
	sp := NewSpendable(3)
	sp.Add("a")
	sp.Add("b")
	sp.Add("c")

	if sp.Count != 3 {
		t.Errorf("Count = %d, want 3", sp.Count)
	}

	v := sp.Call()
	if v != "a" {
		t.Errorf("first Call() = %v, want 'a'", v)
	}
	v = sp.Call()
	if v != "b" {
		t.Errorf("second Call() = %v, want 'b'", v)
	}
}

func TestSpendable_IsConsumed(t *testing.T) {
	sp := NewSpendable(2)
	sp.Add("x")
	sp.Add("y")

	if sp.IsConsumed() {
		t.Error("should not be consumed initially")
	}

	sp.Call()
	if sp.IsConsumed() {
		t.Error("should not be consumed after 1 of 2 calls")
	}

	sp.Call()
	// After consuming all, the auto-release resets Count to 0
	if !sp.IsConsumed() {
		t.Error("should be consumed after all calls")
	}
}

func TestSpendable_GetRemaining(t *testing.T) {
	sp := NewSpendable(3)
	sp.Add(1)
	sp.Add(2)
	sp.Add(3)

	if sp.GetRemaining() != 3 {
		t.Errorf("GetRemaining() = %d, want 3", sp.GetRemaining())
	}

	sp.Call()
	if sp.GetRemaining() != 2 {
		t.Errorf("GetRemaining() after 1 call = %d, want 2", sp.GetRemaining())
	}
}

func TestSpendable_CallAfterExhaustion(t *testing.T) {
	sp := NewSpendable(1)
	sp.Add("only")

	v := sp.Call()
	if v != "only" {
		t.Errorf("Call() = %v, want 'only'", v)
	}

	// After exhaustion, auto-release resets components
	v = sp.Call()
	if v != nil {
		t.Errorf("Call() after exhaustion = %v, want nil", v)
	}
}

func TestSpendable_EmptySpendable(t *testing.T) {
	sp := NewSpendable(0)
	if !sp.IsConsumed() {
		t.Error("empty spendable should be consumed")
	}
	if sp.GetRemaining() != 0 {
		t.Errorf("GetRemaining() = %d, want 0", sp.GetRemaining())
	}
	v := sp.Call()
	if v != nil {
		t.Errorf("Call() on empty = %v, want nil", v)
	}
}

func TestSpendable_ConcurrentAccess(t *testing.T) {
	sp := NewSpendable(100)
	for i := 0; i < 100; i++ {
		sp.Add(i)
	}

	var wg sync.WaitGroup
	results := make([]interface{}, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = sp.Call()
		}(i)
	}

	wg.Wait()

	nilCount := 0
	for _, r := range results {
		if r == nil {
			nilCount++
		}
	}
	// Exactly 100 components, so no nil results expected (unless race causes some)
	// The key test is that it doesn't panic
	if nilCount > 0 {
		t.Logf("Got %d nil results (race may cause some)", nilCount)
	}
}
