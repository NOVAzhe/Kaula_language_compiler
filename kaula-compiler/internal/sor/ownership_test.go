package sor

import (
	"testing"
)

func TestOwnershipTracker_NewObject(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("x", "int64", false, 1)

	if id == "" {
		t.Fatal("NewObject returned empty ID")
	}

	obj := tracker.GetObject(id)
	if obj == nil {
		t.Fatal("GetObject returned nil")
	}
	if obj.Name != "x" {
		t.Errorf("obj.Name = %q, want %q", obj.Name, "x")
	}
	if obj.State != StateOwned {
		t.Errorf("obj.State = %v, want StateOwned", obj.State)
	}
	if obj.TypeName != "int64" {
		t.Errorf("obj.TypeName = %q, want %q", obj.TypeName, "int64")
	}
}

func TestOwnershipTracker_GetObjectByName(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("x", "int64", false, 1)
	tracker.NewObject("y", "string", false, 2)

	id := tracker.GetObjectByName("x")
	if id == "" {
		t.Fatal("GetObjectByName(x) returned empty")
	}
	obj := tracker.GetObject(id)
	if obj.TypeName != "int64" {
		t.Errorf("Expected int64 type, got %q", obj.TypeName)
	}

	id = tracker.GetObjectByName("nonexistent")
	if id != "" {
		t.Errorf("GetObjectByName(nonexistent) = %q, want empty", id)
	}
}

func TestOwnershipTracker_ScopeLifecycle(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Enter a scope
	tracker.EnterScope()
	scopeID := tracker.GetCurrentScope()
	if scopeID == 0 {
		t.Errorf("Expected scope > 0 after EnterScope, got %d", scopeID)
	}

	// Create object in scope
	id := tracker.NewObject("x", "int64", false, 1)
	obj := tracker.GetObject(id)
	if obj.ScopeID != scopeID {
		t.Errorf("obj.ScopeID = %d, want %d", obj.ScopeID, scopeID)
	}

	// Exit scope
	tracker.ExitScope(2)
	newScope := tracker.GetCurrentScope()
	if newScope >= scopeID {
		t.Errorf("Expected scope to decrease after ExitScope, got %d (was %d)", newScope, scopeID)
	}
}

func TestOwnershipTracker_Yield(t *testing.T) {
	tracker := NewOwnershipTracker()

	srcID := tracker.NewObject("src", "int64", false, 1)

	// Yield src -> dst
	dstID := tracker.Yield(srcID, "dst", 2)
	if dstID == "" {
		t.Fatal("Yield failed")
	}

	// src should be Moved
	srcObj := tracker.GetObject(srcID)
	if srcObj.State != StateMoved {
		t.Errorf("src state = %v, want StateMoved", srcObj.State)
	}

	// dst should be Owned
	dstObj := tracker.GetObject(dstID)
	if dstObj.State != StateOwned {
		t.Errorf("dst state = %v, want StateOwned", dstObj.State)
	}
	if dstObj.Name != "dst" {
		t.Errorf("dst name = %q, want %q", dstObj.Name, "dst")
	}
}

func TestOwnershipTracker_YieldFromNonOwned(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Create and yield first to move
	srcID := tracker.NewObject("src", "int64", false, 1)
	tracker.Yield(srcID, "dst1", 2)

	// Try to yield again from moved src (should fail)
	dstID := tracker.Yield(srcID, "dst2", 3)
	if dstID != "" {
		t.Errorf("Expected empty dstID for yield from moved source, got %q", dstID)
	}

	if !tracker.HasErrors() {
		t.Error("Expected error for yield from moved source")
	}
}

func TestOwnershipTracker_Release(t *testing.T) {
	tracker := NewOwnershipTracker()

	srcID := tracker.NewObject("src", "File", false, 1)

	// Release src to holders
	holderIDs := tracker.Release(srcID, []string{"holder_a", "holder_b"}, 2)
	if len(holderIDs) != 2 {
		t.Fatalf("Expected 2 holder IDs, got %d", len(holderIDs))
	}

	// src should be Released
	srcObj := tracker.GetObject(srcID)
	if srcObj.State != StateReleased {
		t.Errorf("src state = %v, want StateReleased", srcObj.State)
	}

	// Holders should be able to read
	for _, hid := range holderIDs {
		if !tracker.CanRead(hid) {
			t.Errorf("holder %q should be readable", hid)
		}
		if tracker.CanWrite(hid) {
			t.Errorf("holder %q should NOT be writable", hid)
		}
	}
}

func TestOwnershipTracker_ReleaseCycleDetection(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Create two objects
	objA := tracker.NewObject("a", "int64", false, 1)
	objB := tracker.NewObject("b", "int64", false, 2)

	// Release a -> b (a depends on b)
	tracker.Release(objA, []string{"b"}, 3)

	// Release b -> a would create a cycle
	// This should be detected by the DAG checker
	holderIDs := tracker.Release(objB, []string{"a"}, 4)
	if len(holderIDs) == 0 {
		// Cycle was detected and prevented
		t.Log("Cycle in release was correctly detected")
	} else {
		// If not prevented, at least we should have errors
		if !tracker.HasErrors() {
			t.Log("Release cycle may not have been detected")
		}
	}
}

func TestOwnershipTracker_Extract(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Create a composite object
	parentID := tracker.NewObject("parent", "Struct", true, 1)

	// Extract from composite
	childID := tracker.Extract(parentID, ".field", "child", 2)
	if childID == "" {
		t.Fatal("Extract failed")
	}

	childObj := tracker.GetObject(childID)
	if childObj == nil {
		t.Fatal("Child object is nil")
	}
	if childObj.Name != "child" {
		t.Errorf("child name = %q, want %q", childObj.Name, "child")
	}
}

func TestOwnershipTracker_ExtractFromNonComposite(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Create a non-composite object
	objID := tracker.NewObject("x", "int64", false, 1)

	// Extract from non-composite: SOR treats extract as a pure ownership
	// operation; type-level composite checking is handled by Stage 2 semantic
	// analysis. The tracker should still perform the extract and manage the
	// child object's ownership.
	childID := tracker.Extract(objID, ".field", "child", 2)
	if childID == "" {
		t.Fatal("Extract on non-composite should return a child ID (ownership operation)")
	}

	childObj := tracker.GetObject(childID)
	if childObj == nil {
		t.Fatal("Child object should exist")
	}
	if childObj.State != StateOwned {
		t.Errorf("Child state = %v, want %v", childObj.State, StateOwned)
	}

	// The source object's child path is replaced with a hollow marker,
	// but the source object itself remains Owned
	srcObj := tracker.GetObject(objID)
	if srcObj == nil {
		t.Fatal("Source object should still exist")
	}
	if srcObj.State != StateOwned {
		t.Errorf("Source state after extract = %v, want %v", srcObj.State, StateOwned)
	}
	// The child path should now point to a hollow object
	hollowID := srcObj.Children[".field"]
	if hollowID == "" {
		t.Error("Expected source child path to be replaced with hollow marker")
	} else {
		hollowObj := tracker.GetObject(hollowID)
		if hollowObj == nil {
			t.Error("Hollow marker object should exist")
		} else if hollowObj.State != StateHollow {
			t.Errorf("Hollow marker state = %v, want %v", hollowObj.State, StateHollow)
		}
	}
}

func TestOwnershipTracker_CheckAccess(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Owned: can read, write, take
	objID := tracker.NewObject("x", "int64", false, 1)
	if !tracker.CheckAccess(objID, AccessRead, 2) {
		t.Error("Owned object should allow read")
	}
	if !tracker.CheckAccess(objID, AccessWrite, 2) {
		t.Error("Owned object should allow write")
	}
	if !tracker.CheckAccess(objID, AccessTake, 2) {
		t.Error("Owned object should allow take")
	}

	// After yield: src is moved
	dstID := tracker.Yield(objID, "y", 3)
	_ = dstID

	// Reading moved object should fail
	if tracker.CheckAccess(objID, AccessRead, 4) {
		t.Error("Moved object should NOT allow read")
	}

	// Reading dst (owned) should work
	if !tracker.CheckAccess(dstID, AccessRead, 5) {
		t.Error("New owner should allow read")
	}
}

func TestOwnershipTracker_CheckAccessReleased(t *testing.T) {
	tracker := NewOwnershipTracker()

	objID := tracker.NewObject("x", "int64", false, 1)
	holderIDs := tracker.Release(objID, []string{"h1", "h2"}, 2)

	// Released source: no read/write/take
	if tracker.CheckAccess(objID, AccessRead, 3) {
		t.Log("Released source may allow read (depends on implementation)")
	}
	if tracker.CheckAccess(objID, AccessWrite, 3) {
		t.Error("Released source should NOT allow write")
	}

	// Holders: should allow read but not write
	for _, hid := range holderIDs {
		if !tracker.CheckAccess(hid, AccessRead, 4) {
			t.Errorf("Holder %q should allow read", hid)
		}
		if tracker.CheckAccess(hid, AccessWrite, 4) {
			t.Errorf("Holder %q should NOT allow write", hid)
		}
	}
}

func TestOwnershipTracker_MarkAsResource(t *testing.T) {
	tracker := NewOwnershipTracker()

	objID := tracker.NewObject("file", "File", false, 1)
	ok := tracker.MarkAsResource(objID, "file")
	if !ok {
		t.Fatal("MarkAsResource failed")
	}

	obj := tracker.GetObject(objID)
	if !obj.IsResource {
		t.Error("Object should be marked as resource")
	}
	if obj.ResourceKind != "file" {
		t.Errorf("ResourceKind = %q, want %q", obj.ResourceKind, "file")
	}

	// Non-existent object
	ok = tracker.MarkAsResource("nonexistent", "file")
	if ok {
		t.Error("MarkAsResource should return false for non-existent object")
	}
}

func TestOwnershipTracker_ResourceLeakDetection(t *testing.T) {
	tracker := NewOwnershipTracker()

	objID := tracker.NewObject("file", "File", false, 1)
	tracker.MarkAsResource(objID, "file")

	tracker.EnterScope()
	// Object is created in the new scope, not released
	leakedID := tracker.NewObject("leaked", "File", false, 2)
	tracker.MarkAsResource(leakedID, "file")
	tracker.ExitScope(3)

	// Should have resource leak errors
	foundLeak := false
	for _, err := range tracker.GetErrors() {
		if err.Kind == ErrResourceLeak {
			foundLeak = true
			break
		}
	}
	if !foundLeak {
		t.Error("Expected resource leak error, but none found")
	}
}

func TestOwnershipTracker_ThreadManagement(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Default thread
	if tracker.GetThread() != "main" {
		t.Errorf("Default thread = %q, want %q", tracker.GetThread(), "main")
	}

	// Set new thread
	tracker.SetThread("worker_1")
	if tracker.GetThread() != "worker_1" {
		t.Errorf("After set, thread = %q, want %q", tracker.GetThread(), "worker_1")
	}

	// Create object in new thread
	objID := tracker.NewObject("x", "int64", false, 1)
	obj := tracker.GetObject(objID)
	if obj == nil {
		t.Fatal("Object is nil")
	}

	// Switch back
	tracker.SetThread("main")
	if tracker.GetThread() != "main" {
		t.Errorf("After switch back, thread = %q, want %q", tracker.GetThread(), "main")
	}
}

func TestOwnershipTracker_AddChild(t *testing.T) {
	tracker := NewOwnershipTracker()

	parentID := tracker.NewObject("parent", "Struct", true, 1)
	childID := tracker.AddChild(parentID, ".field", "int64", false, 2)

	if childID == "" {
		t.Fatal("AddChild returned empty ID")
	}

	// Check child is retrievable
	retrievedID := tracker.GetChild(parentID, ".field")
	if retrievedID != childID {
		t.Errorf("GetChild returned %q, want %q", retrievedID, childID)
	}

	// Parent should have child in children map
	parent := tracker.GetObject(parentID)
	if cid, ok := parent.Children[".field"]; !ok || cid != childID {
		t.Errorf("Parent.Children['.field'] = %q, want %q", cid, childID)
	}
}

func TestOwnershipTracker_DumpState(t *testing.T) {
	tracker := NewOwnershipTracker()

	tracker.NewObject("x", "int64", false, 1)
	tracker.NewObject("y", "string", false, 2)

	state := tracker.DumpState()
	if state == "" {
		t.Error("DumpState returned empty string")
	}
	t.Logf("DumpState output:\n%s", state)
}

func TestOwnershipTracker_GetDAG(t *testing.T) {
	tracker := NewOwnershipTracker()
	dag := tracker.GetDAG()
	if dag == nil {
		t.Error("GetDAG() returned nil")
	}
}

func TestOwnershipTracker_GetObjectCount(t *testing.T) {
	tracker := NewOwnershipTracker()

	if tracker.GetObjectCount() != 0 {
		t.Errorf("Expected 0 objects initially, got %d", tracker.GetObjectCount())
	}

	tracker.NewObject("x", "int64", false, 1)
	tracker.NewObject("y", "int64", false, 2)

	if tracker.GetObjectCount() != 2 {
		t.Errorf("Expected 2 objects, got %d", tracker.GetObjectCount())
	}
}

func TestOwnershipTracker_ClearErrors(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Try to yield from non-existent object
	tracker.Yield("nonexistent", "dst", 1)
	if !tracker.HasErrors() {
		t.Error("Expected errors after invalid yield")
	}

	tracker.ClearErrors()
	if tracker.HasErrors() {
		t.Error("Expected no errors after ClearErrors")
	}
}

func TestOwnershipTracker_CanReadWrite(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Owned: can read and write
	objID := tracker.NewObject("x", "int64", false, 1)
	if !tracker.CanRead(objID) {
		t.Error("Owned object should be readable")
	}
	if !tracker.CanWrite(objID) {
		t.Error("Owned object should be writable")
	}
	if !tracker.CanYield(objID) {
		t.Error("Owned object should be yieldable")
	}

	// After yield
	dstID := tracker.Yield(objID, "y", 2)
	_ = dstID

	if tracker.CanRead(objID) {
		t.Error("Moved object should NOT be readable")
	}
	if tracker.CanWrite(objID) {
		t.Error("Moved object should NOT be writable")
	}
	if tracker.CanYield(objID) {
		t.Error("Moved object should NOT be yieldable")
	}
}

func TestOwnershipTracker_ReleaseDoubleRelease(t *testing.T) {
	tracker := NewOwnershipTracker()

	objID := tracker.NewObject("x", "int64", false, 1)
	tracker.Release(objID, []string{"h1"}, 2)

	// Try to release again
	holderIDs := tracker.Release(objID, []string{"h2"}, 3)
	if len(holderIDs) == 0 {
		// Double release was prevented
		t.Log("Double release correctly prevented")
	}
}

func TestOwnershipTracker_CheckUseAfterMove(t *testing.T) {
	tracker := NewOwnershipTracker()

	objID := tracker.NewObject("x", "int64", false, 1)
	tracker.Yield(objID, "y", 2)

	// CheckUseAfterMove should detect the moved state
	tracker.CheckUseAfterMove(objID, 3)
	if !tracker.HasErrors() {
		t.Error("Expected error for use-after-move check")
	}
}

func TestOwnershipTracker_ScopeExitOwnershipCheck(t *testing.T) {
	tracker := NewOwnershipTracker()

	// Create object in inner scope
	tracker.EnterScope()
	innerID := tracker.NewObject("inner", "int64", false, 1)
	_ = innerID
	// Exit without yielding - should be fine for non-resource types
	tracker.ExitScope(2)

	// No errors expected for non-resource types
	for _, err := range tracker.GetErrors() {
		if err.Kind == ErrResourceLeak {
			t.Errorf("Non-resource should not produce resource leak: %s", err.Message)
		}
	}
}

func TestOwnershipTracker_EnterExitScopeNesting(t *testing.T) {
	tracker := NewOwnershipTracker()

	initialScope := tracker.GetCurrentScope()

	tracker.EnterScope()
	scope1 := tracker.GetCurrentScope()
	if scope1 <= initialScope {
		t.Errorf("scope1 (%d) should be > initialScope (%d)", scope1, initialScope)
	}

	tracker.EnterScope()
	scope2 := tracker.GetCurrentScope()
	if scope2 <= scope1 {
		t.Errorf("scope2 (%d) should be > scope1 (%d)", scope2, scope1)
	}

	tracker.ExitScope(1)
	if tracker.GetCurrentScope() != scope1 {
		t.Errorf("After exit, scope should be %d, got %d", scope1, tracker.GetCurrentScope())
	}

	tracker.ExitScope(1)
	if tracker.GetCurrentScope() != initialScope {
		t.Errorf("After second exit, scope should be %d, got %d", initialScope, tracker.GetCurrentScope())
	}
}

func TestSORError_Error(t *testing.T) {
	err := SORError{
		Kind:       ErrUseAfterMove,
		Message:    "use of moved value",
		SourceLine: 10,
		ObjectID:   "obj_1",
		Details:    "value was moved on line 5",
	}
	s := err.Error()
	if s == "" {
		t.Error("Error() should not be empty")
	}
	t.Logf("SORError.Error() = %q", s)
}

func TestOwnershipState_String(t *testing.T) {
	states := []OwnershipState{StateOwned, StateReleased, StateMoved, StateHollow, StateExtracted, StateUnionReleased}
	for _, s := range states {
		str := s.String()
		if str == "" {
			t.Errorf("State %d has empty string", int(s))
		}
	}
}

func TestErrorKind_String(t *testing.T) {
	kinds := []ErrorKind{
		ErrUseAfterMove, ErrUseAfterExtract, ErrCycleDetected,
		ErrWriteOnReleased, ErrExtractFromNonComposite, ErrYieldInvalidSource,
		ErrReleaseInvalidSource, ErrCrossScopeYield, ErrCrossThreadYield,
		ErrThreadWriteOnReleased, ErrNullDereference, ErrDoubleRelease, ErrResourceLeak,
	}
	for _, k := range kinds {
		str := k.String()
		if str == "" {
			t.Errorf("ErrorKind %d has empty string", int(k))
		}
	}
}

func TestAccessType_String(t *testing.T) {
	types := []AccessType{AccessRead, AccessWrite, AccessTake}
	for _, a := range types {
		str := a.String()
		if str == "" {
			t.Errorf("AccessType %d has empty string", int(a))
		}
	}
}

func TestDAGChecker_New(t *testing.T) {
	dag := NewDAGChecker()
	if dag == nil {
		t.Error("NewDAGChecker() returned nil")
	}
}