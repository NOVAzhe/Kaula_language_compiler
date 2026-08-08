package sor

import (
	"testing"
)

// ============================================================================
// EscapeAnalyzer 基础测试
// ============================================================================

func TestNewEscapeAnalyzer(t *testing.T) {
	ea := NewEscapeAnalyzer()
	if ea == nil {
		t.Fatal("NewEscapeAnalyzer() returned nil")
	}
}

// ============================================================================
// AnalyzeEscape 测试：空语句列表
// ============================================================================

func TestAnalyzeEscapeEmpty(t *testing.T) {
	ea := NewEscapeAnalyzer()
	results := ea.AnalyzeEscape([]Stmt{})
	if len(results) != 0 {
		t.Errorf("空语句列表应返回空结果, 得到 %d 个结果", len(results))
	}
}

// ============================================================================
// AnalyzeEscape 测试：Yield 语句
// ============================================================================

func TestAnalyzeEscapeYield(t *testing.T) {
	ea := NewEscapeAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		YieldStmt(2, "yield x -> y", "x", "y"),
	}
	results := ea.AnalyzeEscape(stmts)
	// x 和 y 都应出现在结果中
	if _, ok := results["x"]; !ok {
		t.Error("x 应出现在逃逸分析结果中")
	}
	if _, ok := results["y"]; !ok {
		t.Error("y 应出现在逃逸分析结果中")
	}
}

// ============================================================================
// AnalyzeEscape 测试：Release 语句
// ============================================================================

func TestAnalyzeEscapeRelease(t *testing.T) {
	ea := NewEscapeAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "data = ...", "data", "Data", false),
		ReleaseStmt(2, "release data -> [a, b]", "data", "a", "b"),
	}
	results := ea.AnalyzeEscape(stmts)
	for _, name := range []string{"data", "a", "b"} {
		if _, ok := results[name]; !ok {
			t.Errorf("%s 应出现在逃逸分析结果中", name)
		}
	}
}

// ============================================================================
// AnalyzeEscape 测试：Extract 语句
// ============================================================================

func TestAnalyzeEscapeExtract(t *testing.T) {
	ea := NewEscapeAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "arr = ...", "arr", "[]int", true),
		ExtractStmt(2, "extract arr[0] -> first", "arr", "[0]", "first"),
	}
	results := ea.AnalyzeEscape(stmts)
	if _, ok := results["arr"]; !ok {
		t.Error("arr 应出现在逃逸分析结果中")
	}
	if _, ok := results["first"]; !ok {
		t.Error("first 应出现在逃逸分析结果中")
	}
}

// ============================================================================
// AnalyzeEscape 测试：Call 语句传参逃逸
// ============================================================================

func TestAnalyzeEscapeCallArg(t *testing.T) {
	ea := NewEscapeAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		CallStmt(2, "foo(x)", "foo", []string{"x"}, []string{"owned"}),
	}
	results := ea.AnalyzeEscape(stmts)
	if level, ok := results["x"]; ok {
		if level < EscArg {
			t.Errorf("作为函数参数传递后，x 的逃逸级别应为 Arg 或更高, 得到 %v", level)
		}
	}
}

// ============================================================================
// AnalyzeEscape 测试：逃逸传播
// ============================================================================

func TestAnalyzeEscapePropagation(t *testing.T) {
	ea := NewEscapeAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		LetStmt(2, "y = 2", "y", "int64", false),
		YieldStmt(3, "yield x -> y", "x", "y"),
		CallStmt(4, "foo(y)", "foo", []string{"y"}, []string{"owned"}),
	}
	results := ea.AnalyzeEscape(stmts)
	// y 作为参数传递给 foo，应达到 EscArg
	if level, ok := results["y"]; ok {
		if level < EscArg {
			t.Errorf("作为参数调用后 y 应达到 Arg 级别, 得到 %v", level)
		}
	}
	// x 通过 yield 传递给 y，但传播是单向的（From→To），
	// 所以 x 保持为 EscNone（除非有反向传播）
	// 这里仅验证 x 存在在结果中
	if _, ok := results["x"]; !ok {
		t.Error("x 应出现在逃逸分析结果中")
	}
}

// ============================================================================
// AnalyzeEscape 测试：Release 传播
// ============================================================================

func TestAnalyzeEscapeReleasePropagation(t *testing.T) {
	ea := NewEscapeAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		CallStmt(2, "foo(x)", "foo", []string{"x"}, []string{"owned"}),
		ReleaseStmt(3, "release x -> [a, b]", "x", "a", "b"),
	}
	results := ea.AnalyzeEscape(stmts)
	// x 作为参数被调用，应达到 Arg 级别
	// a 和 b 通过 release 从 x 传播，应至少达到 Arg 级别
	for _, name := range []string{"a", "b"} {
		if level, ok := results[name]; ok {
			if level < EscArg {
				t.Errorf("通过 release 传播后 %s 应达到 Arg 级别, 得到 %v", name, level)
			}
		}
	}
}

// ============================================================================
// EscapeToAlloc 映射测试
// ============================================================================

func TestEscapeToAlloc(t *testing.T) {
	tests := []struct {
		level EscapeLevel
		want  AllocKind
	}{
		{EscNone, AllocStack},
		{EscArg, AllocArenaSmall},
		{EscCrossScope, AllocArenaSmall},
		{EscReturn, AllocBumpPool},
		{EscGlobal, AllocBumpPool},
		{EscHeap, AllocBumpPool},
	}
	for _, tc := range tests {
		got := EscapeToAlloc(tc.level)
		if got != tc.want {
			t.Errorf("EscapeToAlloc(%v) = %v, 期望 %v", tc.level, got, tc.want)
		}
	}
}

// ============================================================================
// EscapeForcesAlloc 测试
// ============================================================================

func TestEscapeForcesAlloc(t *testing.T) {
	tests := []struct {
		level EscapeLevel
		want  bool
	}{
		{EscNone, false},
		{EscArg, false},
		{EscCrossScope, true},  // EscCrossScope(3) >= EscReturn(3) → true
		{EscReturn, true},
		{EscGlobal, true},
		{EscHeap, true},
	}
	for _, tc := range tests {
		got := EscapeForcesAlloc(tc.level)
		if got != tc.want {
			t.Errorf("EscapeForcesAlloc(%v) = %v, 期望 %v", tc.level, got, tc.want)
		}
	}
}

// ============================================================================
// FormatEscapeSummary 测试
// ============================================================================

func TestFormatEscapeSummary(t *testing.T) {
	summary := FormatEscapeSummary(map[string]EscapeLevel{})
	if summary != "(no escape results)" {
		t.Errorf("空结果摘要 = %q, 期望 (no escape results)", summary)
	}
}

func TestFormatEscapeSummaryNonEmpty(t *testing.T) {
	results := map[string]EscapeLevel{
		"x": EscNone,
		"y": EscHeap,
	}
	summary := FormatEscapeSummary(results)
	if summary == "" {
		t.Error("非空结果摘要不应为空")
	}
}

// ============================================================================
// GetFlowEdges 测试
// ============================================================================

func TestGetFlowEdges(t *testing.T) {
	ea := NewEscapeAnalyzer()
	ea.AnalyzeEscape([]Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		YieldStmt(2, "yield x -> y", "x", "y"),
	})
	edges := ea.GetFlowEdges()
	if len(edges) == 0 {
		t.Error("yield 应产生数据流边")
	}
	found := false
	for _, e := range edges {
		if e.From == "x" && e.To == "y" && e.Kind == "yield" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("未找到 x->y 的 yield 数据流边, 得到: %v", edges)
	}
}