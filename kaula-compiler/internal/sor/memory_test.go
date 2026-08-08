package sor

import (
	"testing"
)

// ============================================================================
// AllocKind 字符串表示测试
// ============================================================================

func TestAllocKindString(t *testing.T) {
	tests := []struct {
		kind AllocKind
		want string
	}{
		{AllocStack, "Stack"},
		{AllocBumpPool, "BumpPool"},
		{AllocArenaTiny, "ArenaTiny"},
		{AllocArenaSmall, "ArenaSmall"},
		{AllocArenaMedium, "ArenaMedium"},
	}
	for _, tc := range tests {
		got := tc.kind.String()
		if got != tc.want {
			t.Errorf("AllocKind(%d).String() = %q, 期望 %q", int(tc.kind), got, tc.want)
		}
	}
}

func TestAllocKindUnknown(t *testing.T) {
	kind := AllocKind(999)
	s := kind.String()
	if s == "" {
		t.Error("未知 AllocKind 的 String() 不应为空")
	}
}

// ============================================================================
// DropAction 字符串表示测试
// ============================================================================

func TestDropActionString(t *testing.T) {
	tests := []struct {
		action DropAction
		want   string
	}{
		{DropNone, "None"},
		{DropScopeEnd, "ScopeEnd"},
		{DropHollow, "Hollow"},
	}
	for _, tc := range tests {
		got := tc.action.String()
		if got != tc.want {
			t.Errorf("DropAction(%d).String() = %q, 期望 %q", int(tc.action), got, tc.want)
		}
	}
}

// ============================================================================
// MemoryDecision 基础测试
// ============================================================================

func TestMemoryDecisionString(t *testing.T) {
	d := &MemoryDecision{
		VarName:    "x",
		ObjID:      "obj_1",
		AllocKind:  AllocStack,
		DropAction: DropScopeEnd,
		FinalState: StateOwned,
		ScopeID:    0,
	}
	s := d.String()
	if s == "" {
		t.Error("MemoryDecision.String() 不应为空")
	}
}

func TestMemoryDecisionStringComposite(t *testing.T) {
	d := &MemoryDecision{
		VarName:           "arr",
		ObjID:             "obj_1",
		AllocKind:         AllocBumpPool,
		DropAction:        DropHollow,
		FinalState:        StateOwned,
		ScopeID:           0,
		IsComposite:       true,
		ExtractedChildren: map[string]bool{"[0]": true},
	}
	s := d.String()
	if s == "" {
		t.Error("MemoryDecision.String() 不应为空")
	}
}

// ============================================================================
// NewMemoryAnalyzer 测试
// ============================================================================

func TestNewMemoryAnalyzer(t *testing.T) {
	ma := NewMemoryAnalyzer()
	if ma == nil {
		t.Fatal("NewMemoryAnalyzer() returned nil")
	}
}

// ============================================================================
// AnalyzeMemory 测试：空 tracker
// ============================================================================

func TestAnalyzeMemoryEmptyTracker(t *testing.T) {
	ma := NewMemoryAnalyzer()
	decisions := ma.AnalyzeMemory(NewOwnershipTracker())
	if decisions != nil {
		t.Errorf("空 tracker 的 AnalyzeMemory 应返回 nil, 得到 %d 个决策", len(decisions))
	}
}

// ============================================================================
// AnalyzeMemory 测试：单对象
// ============================================================================

func TestAnalyzeMemorySingleObject(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("x", "int64", false, 1)

	ma := NewMemoryAnalyzer()
	decisions := ma.AnalyzeMemory(tracker)
	if len(decisions) != 1 {
		t.Fatalf("期望 1 个决策, 得到 %d", len(decisions))
	}

	d := decisions[0]
	if d.VarName != "x" {
		t.Errorf("VarName = %q, 期望 x", d.VarName)
	}
	if d.FinalState != StateOwned {
		t.Errorf("FinalState = %v, 期望 Owned", d.FinalState)
	}
	if d.DropAction != DropScopeEnd {
		t.Errorf("DropAction = %v, 期望 ScopeEnd", d.DropAction)
	}
	// int64 应分配在栈上
	if d.AllocKind != AllocStack {
		t.Errorf("int64 的 AllocKind = %v, 期望 Stack", d.AllocKind)
	}
}

// ============================================================================
// AnalyzeMemory 测试：释放后的对象
// ============================================================================

func TestAnalyzeMemoryReleasedObject(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("data", "Data", false, 1)
	tracker.Release(srcID, []string{"h1"}, 2)

	ma := NewMemoryAnalyzer()
	decisions := ma.AnalyzeMemory(tracker)

	dataFound := false
	for _, d := range decisions {
		if d.VarName == "data" {
			dataFound = true
			if d.FinalState != StateReleased {
				t.Errorf("data 的 FinalState = %v, 期望 Released", d.FinalState)
			}
			break
		}
	}
	if !dataFound {
		t.Error("data 未出现在决策中")
	}
}

// ============================================================================
// AnalyzeMemory 测试：Moved 对象
// ============================================================================

func TestAnalyzeMemoryMovedObject(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("x", "int64", false, 1)
	tracker.Yield(srcID, "y", 2)

	ma := NewMemoryAnalyzer()
	decisions := ma.AnalyzeMemory(tracker)

	xFound := false
	for _, d := range decisions {
		if d.VarName == "x" {
			xFound = true
			if d.DropAction != DropNone {
				t.Errorf("Moved 对象 x 的 DropAction = %v, 期望 None", d.DropAction)
			}
			break
		}
	}
	if !xFound {
		t.Error("x 未出现在决策中")
	}
}

// ============================================================================
// AnalyzeMemory 测试：复合类型 extract
// ============================================================================

func TestAnalyzeMemoryExtractHollow(t *testing.T) {
	tracker := NewOwnershipTracker()
	srcID := tracker.NewObject("arr", "[]int", true, 1)
	tracker.Extract(srcID, "[0]", "first", 2)

	ma := NewMemoryAnalyzer()
	decisions := ma.AnalyzeMemory(tracker)

	arrFound := false
	for _, d := range decisions {
		if d.VarName == "arr" {
			arrFound = true
			if d.IsComposite != true {
				t.Error("arr 应标记为复合类型")
			}
			if len(d.ExtractedChildren) == 0 {
				t.Error("arr 应有 extracted children")
			}
			break
		}
	}
	if !arrFound {
		t.Error("arr 未出现在决策中")
	}
}

// ============================================================================
// GetDecision 测试
// ============================================================================

func TestGetDecision(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("x", "int64", false, 1)

	ma := NewMemoryAnalyzer()
	ma.AnalyzeMemory(tracker)

	d := ma.GetDecision("x")
	if d == nil {
		t.Fatal("GetDecision(x) 返回 nil")
	}
	if d.VarName != "x" {
		t.Errorf("VarName = %q, 期望 x", d.VarName)
	}
}

func TestGetDecisionNotFound(t *testing.T) {
	ma := NewMemoryAnalyzer()
	d := ma.GetDecision("nonexistent")
	if d != nil {
		t.Error("GetDecision 对不存在的变量应返回 nil")
	}
}

// ============================================================================
// determineAllocKind 间接测试（通过 AnalyzeMemory）
// ============================================================================

func TestAnalyzeMemoryAllocKindByType(t *testing.T) {
	tests := []struct {
		typeName string
		wantKind AllocKind
	}{
		{"int64", AllocStack},
		{"int32", AllocStack},
		{"float64", AllocStack},
		{"bool", AllocStack},
		{"[]int", AllocBumpPool},
		{"string", AllocBumpPool},
		{"str", AllocBumpPool},
		{"CustomStruct", AllocBumpPool},
	}
	for _, tc := range tests {
		tracker := NewOwnershipTracker()
		tracker.NewObject("v", tc.typeName, false, 1)
		ma := NewMemoryAnalyzer()
		decisions := ma.AnalyzeMemory(tracker)
		if len(decisions) == 0 {
			t.Errorf("类型 %s 未产生决策", tc.typeName)
			continue
		}
		if decisions[0].AllocKind != tc.wantKind {
			t.Errorf("类型 %s 的 AllocKind = %v, 期望 %v", tc.typeName, decisions[0].AllocKind, tc.wantKind)
		}
	}
}

// ============================================================================
// FormatDecisionsSummary 测试
// ============================================================================

func TestFormatDecisionsSummaryEmpty(t *testing.T) {
	summary := FormatDecisionsSummary(nil)
	if summary != "(no decisions)" {
		t.Errorf("空输入摘要 = %q, 期望 (no decisions)", summary)
	}
}

func TestFormatDecisionsSummaryNonEmpty(t *testing.T) {
	decisions := []*MemoryDecision{
		{VarName: "x", ObjID: "obj_1", AllocKind: AllocStack, DropAction: DropScopeEnd, FinalState: StateOwned, ScopeID: 0},
	}
	summary := FormatDecisionsSummary(decisions)
	if summary == "" {
		t.Error("非空决策摘要不应为空")
	}
}

// ============================================================================
// AnalyzeASTWithMemory 测试（无需 AST 的简化版）
// ============================================================================

// testProgram 是一个简单的测试用 program 类型，实现 GetStmts() 接口
type testProgram struct {
	stmts []Stmt
}

func (p *testProgram) GetStmts() []Stmt {
	return p.stmts
}

func TestAnalyzeASTWithMemory(t *testing.T) {
	program := &testProgram{
		stmts: []Stmt{
			LetStmt(1, "x = 1", "x", "int64", false),
		},
	}
	errors, decisions, execLog := AnalyzeASTWithMemory(program)
	if errors == nil {
		t.Error("errors 不应为 nil")
	}
	if decisions == nil {
		t.Error("decisions 不应为 nil")
	}
	if execLog == nil {
		t.Error("execLog 不应为 nil")
	}
	if len(execLog) == 0 {
		t.Error("execLog 不应为空")
	}
}

// ============================================================================
// EstimatePoolCapacityFromAST 测试：nil 输入
// ============================================================================

func TestEstimatePoolCapacityFromASTNil(t *testing.T) {
	capacity := EstimatePoolCapacityFromAST(nil)
	if capacity != 0 {
		t.Errorf("nil 输入的容量估算 = %d, 期望 0", capacity)
	}
}

// ============================================================================
// MemoryDecision 对象序列化测试
// ============================================================================

func TestMemoryDecisionOrder(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("b", "int64", false, 1)
	tracker.NewObject("a", "int64", false, 2)
	tracker.NewObject("c", "int64", false, 3)

	ma := NewMemoryAnalyzer()
	decisions := ma.AnalyzeMemory(tracker)

	if len(decisions) != 3 {
		t.Fatalf("期望 3 个决策, 得到 %d", len(decisions))
	}
	// 确保按名称排序
	if decisions[0].VarName != "a" || decisions[1].VarName != "b" || decisions[2].VarName != "c" {
		t.Errorf("决策应按 VarName 排序: %v", decisions)
	}
}

// ============================================================================
// OwnershipTracker.GetAllObjects 测试（为 MemoryAnalyzer 提供数据）
// ============================================================================

func TestGetAllObjects(t *testing.T) {
	tracker := NewOwnershipTracker()
	tracker.NewObject("x", "int64", false, 1)
	tracker.NewObject("y", "int64", false, 2)
	objects := tracker.GetAllObjects()
	if len(objects) != 2 {
		t.Fatalf("期望 2 个对象, 得到 %d", len(objects))
	}
	// 确保按 ID 排序
	if objects[0].ID > objects[1].ID {
		t.Errorf("对象应按 ID 排序: %v", objects)
	}
}