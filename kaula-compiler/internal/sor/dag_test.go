package sor

import (
	"testing"
)

// ============================================================================
// DAG 环检测器基础功能测试
// ============================================================================

func TestNewDAGChecker(t *testing.T) {
	dag := NewDAGChecker()
	if dag == nil {
		t.Fatal("NewDAGChecker() returned nil")
	}
	if dag.GetNodeCount() != 0 {
		t.Errorf("初始节点数 = %d, 期望 0", dag.GetNodeCount())
	}
	if dag.GetEdgeCount() != 0 {
		t.Errorf("初始边数 = %d, 期望 0", dag.GetEdgeCount())
	}
}

// ============================================================================
// AddEdge 测试
// ============================================================================

func TestAddEdge(t *testing.T) {
	dag := NewDAGChecker()
	ok := dag.AddEdge("a", "b", 1)
	if !ok {
		t.Fatal("AddEdge 应返回 true")
	}
	if dag.GetNodeCount() != 2 {
		t.Errorf("节点数 = %d, 期望 2", dag.GetNodeCount())
	}
	if dag.GetEdgeCount() != 1 {
		t.Errorf("边数 = %d, 期望 1", dag.GetEdgeCount())
	}
}

func TestAddEdgeDuplicate(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	ok := dag.AddEdge("a", "b", 2)
	if ok {
		t.Error("重复添加边应返回 false")
	}
	if dag.GetEdgeCount() != 1 {
		t.Errorf("重复添加后边数 = %d, 期望 1", dag.GetEdgeCount())
	}
}

func TestAddEdgeMultiple(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("a", "c", 2)
	dag.AddEdge("b", "c", 3)
	if dag.GetNodeCount() != 3 {
		t.Errorf("节点数 = %d, 期望 3", dag.GetNodeCount())
	}
	if dag.GetEdgeCount() != 3 {
		t.Errorf("边数 = %d, 期望 3", dag.GetEdgeCount())
	}
}

// ============================================================================
// HasCycle 测试
// ============================================================================

func TestHasCycleNoEdges(t *testing.T) {
	dag := NewDAGChecker()
	if dag.HasCycle() {
		t.Error("空图不应有环")
	}
}

func TestHasCycleSingleEdge(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	if dag.HasCycle() {
		t.Error("单条边不应有环")
	}
}

func TestHasCycleLinearChain(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	dag.AddEdge("c", "d", 3)
	if dag.HasCycle() {
		t.Error("线性链不应有环")
	}
}

func TestHasCycleSimpleCycle(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	dag.AddEdge("c", "a", 3) // 构成环
	if !dag.HasCycle() {
		t.Error("a->b->c->a 应检测到环")
	}
}

func TestHasCycleSelfLoop(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "a", 1)
	if !dag.HasCycle() {
		t.Error("自环应检测到环")
	}
}

func TestHasCycleDiamond(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("a", "c", 2)
	dag.AddEdge("b", "d", 3)
	dag.AddEdge("c", "d", 4)
	if dag.HasCycle() {
		t.Error("菱形结构不应有环")
	}
}

func TestHasCycleComplexCycle(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	dag.AddEdge("c", "d", 3)
	dag.AddEdge("d", "e", 4)
	dag.AddEdge("e", "b", 5) // 构成环 b->c->d->e->b
	if !dag.HasCycle() {
		t.Error("复杂环应检测到")
	}
}

func TestHasCycleDisconnected(t *testing.T) {
	dag := NewDAGChecker()
	// 两个独立的子图，其中一个有环
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("c", "d", 2)
	dag.AddEdge("d", "e", 3)
	dag.AddEdge("e", "c", 4) // c->d->e->c 构成环
	if !dag.HasCycle() {
		t.Error("不连通图中的环应检测到")
	}
}

// ============================================================================
// GetCyclePath 测试
// ============================================================================

func TestGetCyclePathNoCycle(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	if path := dag.GetCyclePath(); path != nil {
		t.Errorf("无环图应返回 nil，得到 %v", path)
	}
}

func TestGetCyclePathSimple(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	dag.AddEdge("c", "a", 3)
	path := dag.GetCyclePath()
	if path == nil {
		t.Fatal("有环图应返回环路径")
	}
	// 路径应为 [a, b, c, a] 或类似
	if len(path) < 3 {
		t.Fatalf("环路径太短: %v", path)
	}
	if path[0] != path[len(path)-1] {
		t.Errorf("环路径首尾应相同: %v", path)
	}
}

// ============================================================================
// TopologicalSort 测试
// ============================================================================

func TestTopologicalSortAcyclic(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	dag.AddEdge("a", "c", 3)
	sorted, ok := dag.TopologicalSort()
	if !ok {
		t.Fatal("DAG 应能拓扑排序")
	}
	if len(sorted) != 3 {
		t.Fatalf("排序结果长度 = %d, 期望 3", len(sorted))
	}
	// a 必须在 b 和 c 之前
	aPos := indexOf(sorted, "a")
	bPos := indexOf(sorted, "b")
	cPos := indexOf(sorted, "c")
	if aPos > bPos || aPos > cPos {
		t.Errorf("a 应在 b 和 c 之前, 排序: %v", sorted)
	}
}

func TestTopologicalSortCyclic(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	dag.AddEdge("c", "a", 3)
	_, ok := dag.TopologicalSort()
	if ok {
		t.Error("有环图拓扑排序应返回 false")
	}
}

func TestTopologicalSortComplex(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("a", "c", 2)
	dag.AddEdge("b", "d", 3)
	dag.AddEdge("c", "d", 4)
	sorted, ok := dag.TopologicalSort()
	if !ok {
		t.Fatal("DAG 应能拓扑排序")
	}
	if len(sorted) != 4 {
		t.Fatalf("排序结果长度 = %d, 期望 4", len(sorted))
	}
	// a 必须在 b 和 c 之前
	aPos := indexOf(sorted, "a")
	bPos := indexOf(sorted, "b")
	cPos := indexOf(sorted, "c")
	dPos := indexOf(sorted, "d")
	if aPos > bPos || aPos > cPos {
		t.Errorf("a 应在 b 和 c 之前, 排序: %v, 各位置: a=%d b=%d c=%d d=%d", sorted, aPos, bPos, cPos, dPos)
	}
	if bPos > dPos || cPos > dPos {
		t.Errorf("b 和 c 应在 d 之前, 排序: %v", sorted)
	}
}

// ============================================================================
// GetAllEdges 测试
// ============================================================================

func TestGetAllEdges(t *testing.T) {
	dag := NewDAGChecker()
	dag.AddEdge("a", "b", 1)
	dag.AddEdge("b", "c", 2)
	edges := dag.GetAllEdges()
	if len(edges) != 2 {
		t.Fatalf("边数 = %d, 期望 2", len(edges))
	}
	if edges[0].From != "a" || edges[0].To != "b" {
		t.Errorf("边[0] = %v, 期望 From=a To=b", edges[0])
	}
	if edges[1].From != "b" || edges[1].To != "c" {
		t.Errorf("边[1] = %v, 期望 From=b To=c", edges[1])
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}