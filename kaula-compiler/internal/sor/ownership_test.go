package sor

import (
	"testing"
)

// ============================================================================
// 所有权追踪器基础功能测试
// ============================================================================

func TestNewOwnershipTracker(t *testing.T) {
	tracker := NewOwnershipTracker()
	if tracker == nil {
		t.Fatal("NewOwnershipTracker() returned nil")
	}
	if tracker.GetCurrentScope() != 0 {
		t.Errorf("初始作用域 = %d, 期望 0", tracker.GetCurrentScope())
	}
	if tracker.GetObjectCount() != 0 {
		t.Errorf("初始对象数 = %d, 期望 0", tracker.GetObjectCount())
	}
	if tracker.GetThread() != "main" {
		t.Errorf("初始线程 = %q, 期望 main", tracker.GetThread())
	}
}

// ============================================================================
// 对象管理测试
// ============================================================================

func TestNewObject(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("x", "int64", false, 1)
	if id == "" {
		t.Fatal("NewObject() returned empty id")
	}

	obj := tracker.GetObject(id)
	if obj == nil {
		t.Fatalf("GetObject(%q) returned nil", id)
	}
	if obj.Name != "x" {
		t.Errorf("Name = %q, 期望 x", obj.Name)
	}
	if obj.State != StateOwned {
		t.Errorf("初始状态 = %v, 期望 Owned", obj.State)
	}
	if obj.ScopeID != 0 {
		t.Errorf("ScopeID = %d, 期望 0", obj.ScopeID)
	}
}

func TestNewObjectIncrementsID(t *testing.T) {
	tracker := NewOwnershipTracker()
	id1 := tracker.NewObject("a", "int64", false, 1)
	id2 := tracker.NewObject("b", "int64", false, 2)
	if id1 == id2 {
		t.Error("连续创建的对象 ID 应不同")
	}
}

func TestGetObjectNonExistent(t *testing.T) {
	tracker := NewOwnershipTracker()
	if obj := tracker.GetObject("nonexistent"); obj != nil {
		t.Error("GetObject 对不存在的对象应返回 nil")
	}
}

func TestGetObjectByName(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("x", "int64", false, 1)
	found := tracker.GetObjectByName("x")
	if found != id {
		t.Errorf("GetObjectByName(x) = %q, 期望 %q", found, id)
	}
}

func TestGetObjectByNameNotFound(t *testing.T) {
	tracker := NewOwnershipTracker()
	if found := tracker.GetObjectByName("nonexistent"); found != "" {
		t.Errorf("GetObjectByName 对不存在的变量应返回空字符串，得到 %q", found)
	}
}

func TestGetObjectByNameScoped(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("x", "int64", false, 1)
	tracker.EnterScope()
	tracker.NewObject("y", "int64", false, 2)
	// 在内层作用域应能找到外层变量
	if found := tracker.GetObjectByName("x"); found == "" {
		t.Error("内层作用域应能通过名称找到外层变量")
	}
}

// ============================================================================
// 资源标记测试
// ============================================================================

func TestMarkAsResource(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("file", "File", false, 1)
	if !tracker.MarkAsResource(id, "file") {
		t.Fatal("MarkAsResource 返回 false，期望 true")
	}
	obj := tracker.GetObject(id)
	if !obj.IsResource {
		t.Error("标记后 IsResource 应为 true")
	}
	if obj.ResourceKind != "file" {
		t.Errorf("ResourceKind = %q, 期望 file", obj.ResourceKind)
	}
}

func TestMarkAsResourceNonExistent(t *testing.T) {
	tracker := NewOwnershipTracker()
	if tracker.MarkAsResource("nonexistent", "file") {
		t.Error("对不存在的对象 MarkAsResource 应返回 false")
	}
}

// ============================================================================
// 作用域管理测试
// ============================================================================

func TestEnterExitScope(t *testing.T) {
	tracker := NewOwnershipTracker()
	startScope := tracker.GetCurrentScope()
	tracker.EnterScope()
	if tracker.GetCurrentScope() != startScope+1 {
		t.Errorf("EnterScope 后作用域 = %d, 期望 %d", tracker.GetCurrentScope(), startScope+1)
	}
	tracker.ExitScope(0)
	if tracker.GetCurrentScope() != startScope {
		t.Errorf("ExitScope 后作用域 = %d, 期望 %d", tracker.GetCurrentScope(), startScope)
	}
}

func TestExitScopeResourceLeak(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.EnterScope()
	id := tracker.NewObject("leak", "File", false, 1)
	tracker.MarkAsResource(id, "file")
	tracker.ExitScope(10)
	if !tracker.HasErrors() {
		t.Fatal("期望资源泄漏错误，但没有错误")
	}
	errs := tracker.GetErrors()
	found := false
	for _, e := range errs {
		if e.Kind == ErrResourceLeak {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("错误中没有 ErrResourceLeak，得到: %v", errs)
	}
}

func TestExitScopeNoLeakWhenYielded(t *testing.T) {
	tracker := NewOwnershipTracker()
	// 创建资源并 yield 到另一变量，源对象变为 Moved
	tracker.EnterScope()
	srcID := tracker.NewObject("inner", "File", false, 1)
	tracker.MarkAsResource(srcID, "file")
	// yield 后 src 变为 Moved，不再持有资源
	dstID := tracker.Yield(srcID, "outer", 2)
	_ = dstID
	// 退出内层作用域
	tracker.ExitScope(3)
	// 检查：源对象 (inner/obj_1) 已 Moved，不应报其资源泄漏
	// 目标对象 (outer/obj_2) 是资源且仍在 Owned 状态，应报泄漏
	errs := tracker.GetErrors()
	innerLeak := false
	outerLeak := false
	for _, e := range errs {
		if e.Kind == ErrResourceLeak {
			if e.ObjectID == "obj_1" {
				innerLeak = true
			}
			if e.ObjectID == "obj_2" {
				outerLeak = true
			}
		}
	}
	if innerLeak {
		t.Error("Moved 状态的源对象不应报资源泄漏")
	}
	if !outerLeak {
		t.Error("Yield 获得所有权的目标对象仍持有资源，应报资源泄漏")
	}
}

func TestExitScopeGlobalCantExit(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.ExitScope(0) // 尝试退出全局作用域
	// 不应该崩溃或添加错误
	if tracker.HasErrors() {
		t.Error("退出全局作用域不应产生错误")
	}
}

// ============================================================================
// 线程管理测试
// ============================================================================

func TestSetThread(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.SetThread("worker1")
	if tracker.GetThread() != "worker1" {
		t.Errorf("GetThread() = %q, 期望 worker1", tracker.GetThread())
	}
}

// ============================================================================
// Yield 原语测试
// ============================================================================

func TestYieldSuccess(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("src", "int64", false, 1)
	dstID := tracker.Yield(srcID, "dst", 2)
	if dstID == "" {
		t.Fatal("Yield 失败，返回空 ID")
	}

	// 源对象应变为 Moved
	src := tracker.GetObject(srcID)
	if src == nil {
		t.Fatal("源对象不应为 nil")
	}
	if src.State != StateMoved {
		t.Errorf("Yield 后源状态 = %v, 期望 Moved", src.State)
	}

	// 目标对象应为 Owned
	dst := tracker.GetObject(dstID)
	if dst == nil {
		t.Fatal("目标对象不应为 nil")
	}
	if dst.State != StateOwned {
		t.Errorf("Yield 后目标状态 = %v, 期望 Owned", dst.State)
	}
}

func TestYieldInvalidSource(t *testing.T) {
	tracker := NewOwnershipTracker()
	dstID := tracker.Yield("nonexistent", "dst", 1)
	if dstID != "" {
		t.Error("对不存在的源对象 Yield 应返回空字符串")
	}
	if !tracker.HasErrors() {
		t.Error("期望有错误，但没有")
	}
}

func TestYieldNonOwnedSource(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("src", "int64", false, 1)
	// 先 yield 一次，使 src 变为 Moved
	tracker.Yield(srcID, "dst1", 1)
	// 再次 yield 同一个源应失败
	dst2ID := tracker.Yield(srcID, "dst2", 2)
	if dst2ID != "" {
		t.Error("对已 Moved 的对象 Yield 应返回空字符串")
	}
}

func TestYieldCompositeCopiesChildren(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("arr", "[]int", true, 1)
	childID := tracker.AddChild(srcID, "[0]", "int64", false, 1)
	if childID == "" {
		t.Fatal("AddChild 失败")
	}

	dstID := tracker.Yield(srcID, "arr2", 2)
	if dstID == "" {
		t.Fatal("Yield 复合对象失败")
	}

	dst := tracker.GetObject(dstID)
	if dst == nil {
		t.Fatal("目标对象为 nil")
	}
	if len(dst.Children) != 1 {
		t.Errorf("目标对象的子元素数 = %d, 期望 1", len(dst.Children))
	}
}

func TestYieldResourceCopiesResourceFlag(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("file", "File", false, 1)
	tracker.MarkAsResource(srcID, "file")
	dstID := tracker.Yield(srcID, "moved", 2)
	dst := tracker.GetObject(dstID)
	if dst == nil || !dst.IsResource {
		t.Error("Yield 资源类型后，目标对象应保持 IsResource 标记")
	}
}

// ============================================================================
// Release 原语测试
// ============================================================================

func TestReleaseSuccess(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("data", "Data", false, 1)
	holderIDs := tracker.Release(srcID, []string{"h1", "h2"}, 2)
	if len(holderIDs) != 2 {
		t.Fatalf("Release 返回 %d 个持有者, 期望 2", len(holderIDs))
	}

	src := tracker.GetObject(srcID)
	if src.State != StateReleased {
		t.Errorf("Release 后源状态 = %v, 期望 Released", src.State)
	}
}

func TestReleaseInvalidSource(t *testing.T) {
	tracker := NewOwnershipTracker()
	holders := tracker.Release("nonexistent", []string{"h1"}, 1)
	if holders != nil {
		t.Error("对不存在的源 Release 应返回 nil")
	}
}

func TestReleaseDoubleRelease(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("data", "Data", false, 1)
	tracker.Release(srcID, []string{"h1"}, 1)
	// 第二次 release 同一个持有者应报 double release 错误
	holders := tracker.Release(srcID, []string{"h1"}, 2)
	if len(holders) != 0 {
		t.Errorf("double release 后应返回 0 个持有者，得到 %d", len(holders))
	}
}

func TestReleaseCycleDetection(t *testing.T) {
	tracker2 := NewOwnershipTracker()
	xID := tracker2.NewObject("x", "Data", false, 1) // Owned
	yID := tracker2.NewObject("y", "Data", false, 2) // Owned
	zID := tracker2.NewObject("z", "Data", false, 3) // Owned

	// x release 给 y, x 变为 Released
	tracker2.Release(xID, []string{"y"}, 4)
	// y 现在是 Owned，release 给 z
	tracker2.Release(yID, []string{"z"}, 5)
	// z 是 Owned，release 给 x 构成环
	tracker2.Release(zID, []string{"x"}, 6)

	if !tracker2.HasErrors() {
		t.Fatal("期望检测到环，但没有错误")
	}
	errs := tracker2.GetErrors()
	found := false
	for _, e := range errs {
		if e.Kind == ErrCycleDetected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望 ErrCycleDetected，得到: %v", errs)
	}
}

// ============================================================================
// Extract 原语测试
// ============================================================================

func TestExtractSuccess(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("arr", "[]int", true, 1)
	elemID := tracker.Extract(srcID, "[0]", "first", 2)
	if elemID == "" {
		t.Fatal("Extract 失败，返回空 ID")
	}

	elem := tracker.GetObject(elemID)
	if elem == nil {
		t.Fatal("提取的元素为 nil")
	}
	if elem.State != StateOwned {
		t.Errorf("提取的元素状态 = %v, 期望 Owned", elem.State)
	}

	// 源位置应变为 hollow
	src := tracker.GetObject(srcID)
	if src.Children["[0]"] == "" {
		t.Error("源对象应保留子元素引用")
	}
	hollow := tracker.GetObject(src.Children["[0]"])
	if hollow == nil || hollow.State != StateHollow {
		t.Errorf("提取后源位置状态应为 Hollow，得到 %v", hollow)
	}
}

func TestExtractNonExistentSource(t *testing.T) {
	tracker := NewOwnershipTracker()
	elemID := tracker.Extract("nonexistent", "[0]", "elem", 1)
	if elemID != "" {
		t.Error("对不存在的源 Extract 应返回空字符串")
	}
}

func TestExtractAlreadyExtracted(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("arr", "[]int", true, 1)
	// 第一次 extract 成功
	elem1ID := tracker.Extract(srcID, "[0]", "first", 2)
	if elem1ID == "" {
		t.Fatal("第一次 Extract 失败")
	}
	// 再次 extract 同一位置应失败
	elem2ID := tracker.Extract(srcID, "[0]", "second", 3)
	if elem2ID != "" {
		t.Error("对已提取位置再次 Extract 应返回空字符串")
	}
}

func TestExtractNonOwnedSource(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("arr", "[]int", true, 1)
	// 先 release 掉，使状态变为 Released
	tracker.Release(srcID, []string{"h1"}, 2)
	// 尝试从 Released 状态 extract 应失败
	elemID := tracker.Extract(srcID, "[0]", "elem", 3)
	if elemID != "" {
		t.Error("从 Released 状态 Extract 应返回空字符串")
	}
}

// ============================================================================
// 权限检查测试
// ============================================================================

func TestCanRead(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("x", "int64", false, 1)
	if !tracker.CanRead(id) {
		t.Error("Owned 状态应可读")
	}
	// 释放后应仍可读
	tracker.Yield(id, "y", 2)
	yID := tracker.GetObjectByName("y")
	if yID == "" {
		t.Fatal("yield 目标未找到")
	}
	if !tracker.CanRead(yID) {
		t.Error("接收 yield 的对象应可读")
	}
}

func TestCanReadReleased(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("data", "Data", false, 1)
	holderIDs := tracker.Release(srcID, []string{"h1"}, 2)
	if len(holderIDs) == 0 {
		t.Fatal("Release 失败")
	}
	if !tracker.CanRead(holderIDs[0]) {
		t.Error("Released 持有者应可读")
	}
}

func TestCanReadNonExistent(t *testing.T) {
	tracker := NewOwnershipTracker()
	if tracker.CanRead("nonexistent") {
		t.Error("不存在的对象应不可读")
	}
}

func TestCanWrite(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("x", "int64", false, 1)
	if !tracker.CanWrite(id) {
		t.Error("Owned 状态应可写")
	}
}

func TestCanWriteReleased(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("data", "Data", false, 1)
	holderIDs := tracker.Release(srcID, []string{"h1"}, 2)
	if len(holderIDs) > 0 {
		if tracker.CanWrite(holderIDs[0]) {
			t.Error("Released 持有者应不可写")
		}
	}
}

func TestCanYield(t *testing.T) {
	tracker := NewOwnershipTracker()
	id := tracker.NewObject("x", "int64", false, 1)
	if !tracker.CanYield(id) {
		t.Error("Owned 状态应可 yield")
	}
	// yield 后源对象不可再 yield
	tracker.Yield(id, "y", 2)
	if tracker.CanYield(id) {
		t.Error("Moved 状态应不可 yield")
	}
}

func TestCheckAccessUseAfterMove(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("x", "int64", false, 1)
	tracker.Yield(srcID, "y", 2)
	// 使用已移动的对象
	ok := tracker.CheckAccess(srcID, AccessRead, 3)
	if ok {
		t.Error("use-after-move 应返回 false")
	}
}

func TestCheckAccessNullDereference(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("arr", "[]int", true, 1)
	tracker.Extract(srcID, "[0]", "elem", 2)
	// 访问空洞位置
	hollowID := tracker.GetChild(srcID, "[0]")
	if hollowID == "" {
		t.Fatal("hollow 对象 ID 不应为空")
	}
	ok := tracker.CheckAccess(hollowID, AccessRead, 3)
	if ok {
		t.Error("访问 hollow 位置应返回 false")
	}
}

func TestCheckAccessWriteOnReleased(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("data", "Data", false, 1)
	holders := tracker.Release(srcID, []string{"h1"}, 2)
	if len(holders) == 0 {
		t.Fatal("Release 失败")
	}
	ok := tracker.CheckAccess(holders[0], AccessWrite, 3)
	if ok {
		t.Error("写 released 持有者应返回 false")
	}
}

func TestCheckAccessNonExistent(t *testing.T) {
	tracker := NewOwnershipTracker()
	ok := tracker.CheckAccess("nonexistent", AccessRead, 1)
	if ok {
		t.Error("访问不存在的对象应返回 false")
	}
}

// ============================================================================
// 子元素管理测试
// ============================================================================

func TestAddChild(t *testing.T) {
	tracker := NewOwnershipTracker()
	parentID := tracker.NewObject("arr", "[]int", true, 1)
	childID := tracker.AddChild(parentID, "[0]", "int64", false, 1)
	if childID == "" {
		t.Fatal("AddChild 返回空字符串")
	}
	child := tracker.GetObject(childID)
	if child == nil {
		t.Fatal("子对象为 nil")
	}
	if child.Name != "arr[0]" {
		t.Errorf("子对象名称 = %q, 期望 arr[0]", child.Name)
	}
}

func TestAddChildNonComposite(t *testing.T) {
	tracker := NewOwnershipTracker()
	parentID := tracker.NewObject("x", "int64", false, 1)
	childID := tracker.AddChild(parentID, "[0]", "int64", false, 1)
	if childID != "" {
		t.Error("对非复合类型 AddChild 应返回空字符串")
	}
}

func TestGetChild(t *testing.T) {
	tracker := NewOwnershipTracker()
	parentID := tracker.NewObject("arr", "[]int", true, 1)
	childID := tracker.AddChild(parentID, "[0]", "int64", false, 1)
	found := tracker.GetChild(parentID, "[0]")
	if found != childID {
		t.Errorf("GetChild 返回 %q, 期望 %q", found, childID)
	}
}

func TestGetChildNonExistent(t *testing.T) {
	tracker := NewOwnershipTracker()
	parentID := tracker.NewObject("arr", "[]int", true, 1)
	if child := tracker.GetChild(parentID, "[999]"); child != "" {
		t.Error("获取不存在的子元素应返回空字符串")
	}
}

// ============================================================================
// 错误管理测试
// ============================================================================

func TestGetErrors(t *testing.T) {
	tracker := NewOwnershipTracker()
	if tracker.HasErrors() {
		t.Error("初始状态不应有错误")
	}
	tracker.Yield("nonexistent", "dst", 1)
	if !tracker.HasErrors() {
		t.Error("Yield 无效源后应有错误")
	}
	if len(tracker.GetErrors()) == 0 {
		t.Error("GetErrors() 不应为空")
	}
}

func TestClearErrors(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.Yield("nonexistent", "dst", 1)
	tracker.ClearErrors()
	if tracker.HasErrors() {
		t.Error("ClearErrors 后不应有错误")
	}
}

// ============================================================================
// 状态管理测试
// ============================================================================

func TestDumpState(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("x", "int64", false, 1)
	state := tracker.DumpState()
	if state == "" {
		t.Error("DumpState 不应返回空字符串")
	}
}

func TestGetDAG(t *testing.T) {
	tracker := NewOwnershipTracker()
	dag := tracker.GetDAG()
	if dag == nil {
		t.Error("GetDAG 不应返回 nil")
	}
}