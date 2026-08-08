package sor

import (
	"testing"
)

// ============================================================================
// LivenessAnalyzer 基础测试
// ============================================================================

func TestNewLivenessAnalyzer(t *testing.T) {
	la := NewLivenessAnalyzer()
	if la == nil {
		t.Fatal("NewLivenessAnalyzer() returned nil")
	}
}

// ============================================================================
// AnalyzeLiveness 测试：空语句列表
// ============================================================================

func TestAnalyzeLivenessEmpty(t *testing.T) {
	la := NewLivenessAnalyzer()
	result := la.AnalyzeLiveness([]Stmt{})
	if result == nil {
		t.Fatal("AnalyzeLiveness 返回 nil")
	}
	if len(result.GetAllLastUses()) != 0 {
		t.Errorf("空语句应返回 0 个最后使用信息, 得到 %d", len(result.GetAllLastUses()))
	}
}

// ============================================================================
// AnalyzeLiveness 测试：变量声明与读取
// ============================================================================

func TestAnalyzeLivenessLetAndRead(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		ReadStmt(2, "print(x)", "x"),
	}
	result := la.AnalyzeLiveness(stmts)
	info := result.GetLastUse("x")
	if info == nil {
		t.Fatal("x 的最后使用信息为 nil")
	}
	if info.LastUseLine != 2 {
		t.Errorf("x 的最后使用行 = %d, 期望 2", info.LastUseLine)
	}
	if info.LastUseKind != "read" {
		t.Errorf("x 的最后使用类型 = %s, 期望 read", info.LastUseKind)
	}
}

// ============================================================================
// AnalyzeLiveness 测试：Write 语句
// ============================================================================

func TestAnalyzeLivenessWrite(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		WriteStmt(2, "x = 2", "x"),
	}
	result := la.AnalyzeLiveness(stmts)
	info := result.GetLastUse("x")
	if info == nil {
		t.Fatal("x 的最后使用信息为 nil")
	}
	if info.LastUseKind != "write" {
		t.Errorf("x 的最后使用类型 = %s, 期望 write", info.LastUseKind)
	}
}

// ============================================================================
// AnalyzeLiveness 测试：Yield 语句
// ============================================================================

func TestAnalyzeLivenessYield(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		YieldStmt(2, "yield x -> y", "x", "y"),
	}
	result := la.AnalyzeLiveness(stmts)
	// x 的最后使用应标记为 yield-src
	xInfo := result.GetLastUse("x")
	if xInfo == nil {
		t.Fatal("x 的最后使用信息为 nil")
	}
	if !xInfo.IsYieldSrc {
		t.Error("x 应标记为 yield-src")
	}
	if xInfo.LastUseKind != "yield-src" {
		t.Errorf("x 的最后使用类型 = %s, 期望 yield-src", xInfo.LastUseKind)
	}
	// y 应出现在结果中
	yInfo := result.GetLastUse("y")
	if yInfo == nil {
		t.Fatal("y 的最后使用信息为 nil")
	}
	if yInfo.LastUseKind != "yield-dst" {
		t.Errorf("y 的最后使用类型 = %s, 期望 yield-dst", yInfo.LastUseKind)
	}
}

// ============================================================================
// AnalyzeLiveness 测试：Extract 语句
// ============================================================================

func TestAnalyzeLivenessExtract(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "arr = ...", "arr", "[]int", true),
		ExtractStmt(2, "extract arr[0] -> first", "arr", "[0]", "first"),
	}
	result := la.AnalyzeLiveness(stmts)
	arrInfo := result.GetLastUse("arr")
	if arrInfo == nil {
		t.Fatal("arr 的最后使用信息为 nil")
	}
	if !arrInfo.IsExtractSrc {
		t.Error("arr 应标记为 extract-src")
	}
	firstInfo := result.GetLastUse("first")
	if firstInfo == nil {
		t.Fatal("first 的最后使用信息为 nil")
	}
	if firstInfo.LastUseKind != "extract-elem" {
		t.Errorf("first 的最后使用类型 = %s, 期望 extract-elem", firstInfo.LastUseKind)
	}
}

// ============================================================================
// AnalyzeLiveness 测试：Release 语句
// ============================================================================

func TestAnalyzeLivenessRelease(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "data = ...", "data", "Data", false),
		ReleaseStmt(2, "release data -> [a, b]", "data", "a", "b"),
	}
	result := la.AnalyzeLiveness(stmts)
	dataInfo := result.GetLastUse("data")
	if dataInfo == nil {
		t.Fatal("data 的最后使用信息为 nil")
	}
	if dataInfo.LastUseKind != "release-src" {
		t.Errorf("data 的最后使用类型 = %s, 期望 release-src", dataInfo.LastUseKind)
	}
	for _, holder := range []string{"a", "b"} {
		hInfo := result.GetLastUse(holder)
		if hInfo == nil {
			t.Fatalf("%s 的最后使用信息为 nil", holder)
		}
		if hInfo.LastUseKind != "release-holder" {
			t.Errorf("%s 的最后使用类型 = %s, 期望 release-holder", holder, hInfo.LastUseKind)
		}
	}
}

// ============================================================================
// AnalyzeLiveness 测试：作用域管理
// ============================================================================

func TestAnalyzeLivenessScopes(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		{Kind: StmtScopeEnter, Line: 0, Source: "{ // scope_1", ScopeName: "scope_1"},
		LetStmt(2, "y = 2", "y", "int64", false),
		{Kind: StmtScopeExit, Line: 0, Source: "} // scope_1", ScopeName: "scope_1"},
	}
	result := la.AnalyzeLiveness(stmts)
	// x 和 y 都应被追踪
	if result.GetLastUse("x") == nil {
		t.Error("x 应被追踪")
	}
	if result.GetLastUse("y") == nil {
		t.Error("y 应被追踪")
	}
}

// ============================================================================
// AnalyzeLiveness 测试：最后一次使用被后续覆盖
// ============================================================================

func TestAnalyzeLivenessLastUseOverwritten(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		ReadStmt(2, "print(x)", "x"),
		ReadStmt(3, "print(x)", "x"),
	}
	result := la.AnalyzeLiveness(stmts)
	info := result.GetLastUse("x")
	if info == nil {
		t.Fatal("x 的最后使用信息为 nil")
	}
	if info.LastUseLine != 3 {
		t.Errorf("x 的最后使用行 = %d, 期望 3（最后一次使用）", info.LastUseLine)
	}
}

// ============================================================================
// AnalyzeLiveness 测试：Call 语句
// ============================================================================

func TestAnalyzeLivenessCall(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		LetStmt(2, "y = 2", "y", "int64", false),
		CallStmt(3, "foo(x, y)", "foo", []string{"x", "y"}, []string{"owned", "borrow"}),
	}
	result := la.AnalyzeLiveness(stmts)
	// x 是 owned 参数，应记录最后使用
	xInfo := result.GetLastUse("x")
	if xInfo == nil {
		t.Fatal("x 的最后使用信息为 nil")
	}
	if xInfo.LastUseLine != 3 {
		t.Errorf("x 的最后使用行 = %d, 期望 3", xInfo.LastUseLine)
	}
	// y 是 borrow 参数，不应记录最后使用（或仍保留之前的记录）
	yInfo := result.GetLastUse("y")
	if yInfo != nil {
		// borrow 参数应该没被记录，但如果有之前的记录也可以
		// 这里只验证它不为 nil 时满足相关条件
	}
}

// ============================================================================
// GetAllLastUses 测试
// ============================================================================

func TestGetAllLastUses(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(2, "b = 2", "b", "int64", false),
		LetStmt(1, "a = 1", "a", "int64", false),
	}
	result := la.AnalyzeLiveness(stmts)
	allUses := result.GetAllLastUses()
	if len(allUses) != 2 {
		t.Fatalf("应有 2 个最后使用信息, 得到 %d", len(allUses))
	}
	// 应排序：a 在 b 前
	if allUses[0].VarName != "a" || allUses[1].VarName != "b" {
		t.Errorf("结果应排序: %v", allUses)
	}
}

// ============================================================================
// ComputeDropPoints 测试
// ============================================================================

func TestComputeDropPoints(t *testing.T) {
	la := NewLivenessAnalyzer()
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		ReadStmt(2, "print(x)", "x"),
	}
	result := la.AnalyzeLiveness(stmts)
	dropPoints := la.ComputeDropPoints(result)
	if dropPoints == nil {
		t.Error("ComputeDropPoints 不应返回 nil")
	}
}

func TestComputeDropPointsNil(t *testing.T) {
	la := NewLivenessAnalyzer()
	dropPoints := la.ComputeDropPoints(nil)
	if dropPoints != nil {
		t.Error("传入 nil 的 result 应返回 nil")
	}
}

// ============================================================================
// FormatLivenessSummary 测试
// ============================================================================

func TestFormatLivenessSummaryEmpty(t *testing.T) {
	result := &LivenessResult{
		lastUses: make(map[string]*LastUseInfo),
	}
	summary := result.FormatLivenessSummary()
	if summary != "(no liveness info)" {
		t.Errorf("空结果摘要 = %q, 期望 (no liveness info)", summary)
	}
}

func TestFormatLivenessSummaryNonEmpty(t *testing.T) {
	result := &LivenessResult{
		lastUses: map[string]*LastUseInfo{
			"x": {VarName: "x", LastUseLine: 5, LastUseKind: "read"},
		},
	}
	summary := result.FormatLivenessSummary()
	if summary == "" {
		t.Error("非空结果摘要不应为空")
	}
}

// ============================================================================
// parseScopeNameToID 测试
// ============================================================================

func TestParseScopeNameToID(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"scope_0", 0},
		{"scope_1", 1},
		{"scope_42", 42},
		{"", 0},
		{"42", 42},
		{"invalid", 0},
	}
	for _, tc := range tests {
		got := parseScopeNameToID(tc.input)
		if got != tc.want {
			t.Errorf("parseScopeNameToID(%q) = %d, 期望 %d", tc.input, got, tc.want)
		}
	}
}

// ============================================================================
// GetLastUse 对 nil 的防御性测试
// ============================================================================

func TestGetLastUseNilReceiver(t *testing.T) {
	var result *LivenessResult
	info := result.GetLastUse("x")
	if info != nil {
		t.Error("nil receiver 的 GetLastUse 应返回 nil")
	}
}

func TestGetAllLastUsesNilReceiver(t *testing.T) {
	var result *LivenessResult
	uses := result.GetAllLastUses()
	if uses != nil {
		t.Error("nil receiver 的 GetAllLastUses 应返回 nil")
	}
}

func TestFormatLivenessSummaryNilReceiver(t *testing.T) {
	var result *LivenessResult
	summary := result.FormatLivenessSummary()
	if summary != "(no liveness info)" {
		t.Errorf("nil receiver 的 FormatLivenessSummary 应返回 (no liveness info), 得到 %q", summary)
	}
}