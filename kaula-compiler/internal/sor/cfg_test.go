package sor

import (
	"testing"
)

func TestBuildCFG_Empty(t *testing.T) {
	cfg := BuildCFG([]Stmt{})

	if len(cfg.Blocks) != 1 {
		t.Errorf("Empty CFG should have 1 block, got %d", len(cfg.Blocks))
	}
	if cfg.Entry != cfg.Exit {
		t.Error("Empty CFG entry and exit should be the same block")
	}
}

func TestBuildCFG_LinearStatements(t *testing.T) {
	stmts := []Stmt{
		LetStmt(1, "x = 1", "x", "int64", false),
		ReadStmt(2, "x", "x"),
		LetStmt(3, "y = 2", "y", "int64", false),
	}

	cfg := BuildCFG(stmts)

	if len(cfg.Blocks) != 1 {
		t.Errorf("Linear statements should form 1 block, got %d", len(cfg.Blocks))
	}
	if len(cfg.Blocks[0].Stmts) != 3 {
		t.Errorf("Block should contain 3 statements, got %d", len(cfg.Blocks[0].Stmts))
	}
	if cfg.Entry != cfg.Exit {
		t.Error("Linear CFG entry and exit should be the same block")
	}
}

func TestBuildCFG_SimpleLoop(t *testing.T) {
	stmts := []Stmt{
		LetStmt(1, "i = 0", "i", "int64", false),
		LoopEnterStmt(2, "for (...)", 10),
		ReadStmt(3, "i", "i"),
		LoopExitStmt(4, "endfor"),
	}

	cfg := BuildCFG(stmts)

	// Should have: init block, loop header, loop body, exit block
	if len(cfg.Blocks) < 3 {
		t.Errorf("Simple loop should have at least 3 blocks, got %d", len(cfg.Blocks))
	}

	// Find loop header (block with LoopEnter statement)
	var loopHeader *BasicBlock
	for _, b := range cfg.Blocks {
		for _, s := range b.Stmts {
			if s.Kind == StmtLoopEnter {
				loopHeader = b
				break
			}
		}
	}

	if loopHeader == nil {
		t.Fatal("Could not find loop header block")
	}

	// Loop header should have a loop edge back to itself or a latch block
	hasLoopEdge := false
	for i, kind := range loopHeader.EdgeKinds {
		if kind == EdgeLoop {
			hasLoopEdge = true
			if i >= len(loopHeader.Succs) {
				t.Error("Loop edge index out of bounds")
			}
			break
		}
	}

	if !hasLoopEdge {
		t.Error("Loop header should have at least one loop edge")
	}
}

func TestBuildCFG_BranchIfElse(t *testing.T) {
	stmts := []Stmt{
		BranchEnterStmt(1, "if (...)"),
		LetStmt(2, "x = 1", "x", "int64", false),
		BranchElseStmt(3, "else"),
		LetStmt(4, "x = 2", "x", "int64", false),
		BranchExitStmt(5, "endif"),
	}

	cfg := BuildCFG(stmts)

	// Should have: branch header, if body, else body, merge block
	if len(cfg.Blocks) < 4 {
		t.Errorf("If-else should have at least 4 blocks, got %d", len(cfg.Blocks))
	}

	// Find branch header (block with BranchEnter statement)
	var branchHeader *BasicBlock
	for _, b := range cfg.Blocks {
		for _, s := range b.Stmts {
			if s.Kind == StmtBranchEnter {
				branchHeader = b
				break
			}
		}
	}

	if branchHeader == nil {
		t.Fatal("Could not find branch header block")
	}

	// Branch header should have 2 branch edges (if and else)
	branchCount := 0
	for _, kind := range branchHeader.EdgeKinds {
		if kind == EdgeBranch {
			branchCount++
		}
	}

	if branchCount != 2 {
		t.Errorf("Branch header should have 2 branch edges, got %d", branchCount)
	}
}

func TestBuildCFG_NestedLoops(t *testing.T) {
	stmts := []Stmt{
		LoopEnterStmt(1, "outer loop", 5),
		LetStmt(2, "i = 0", "i", "int64", false),
		LoopEnterStmt(3, "inner loop", 3),
		LetStmt(4, "j = 0", "j", "int64", false),
		LoopExitStmt(5, "end inner"),
		LoopExitStmt(6, "end outer"),
	}

	cfg := BuildCFG(stmts)

	// Should detect nested structure
	loops := cfg.GetLoopBlocks()

	if len(loops) < 2 {
		t.Errorf("Should detect at least 2 loops (outer and inner), got %d", len(loops))
	}
}

func TestBuildCFG_BranchMerge(t *testing.T) {
	stmts := []Stmt{
		LetStmt(1, "x = 0", "x", "int64", false),
		BranchEnterStmt(2, "if (x > 0)"),
		LetStmt(3, "y = 1", "y", "int64", false),
		BranchExitStmt(4, "endif"),
		ReadStmt(5, "y", "y"),
	}

	cfg := BuildCFG(stmts)

	// Find merge block (block with EdgeMerge incoming edge)
	var mergeBlock *BasicBlock
	for _, b := range cfg.Blocks {
		for _, pred := range b.Preds {
			for i, succ := range pred.Succs {
				if succ == b && i < len(pred.EdgeKinds) && pred.EdgeKinds[i] == EdgeMerge {
					mergeBlock = b
					break
				}
			}
		}
	}

	if mergeBlock == nil {
		t.Fatal("Could not find merge block after branch")
	}

	// Merge block should contain the statement after endif
	hasReadStmt := false
	for _, s := range mergeBlock.Stmts {
		if s.Kind == StmtRead {
			hasReadStmt = true
			break
		}
	}

	if !hasReadStmt {
		t.Error("Merge block should contain the read statement after endif")
	}
}

func TestCFG_GetBranchPairs(t *testing.T) {
	stmts := []Stmt{
		BranchEnterStmt(1, "if (...)"),
		LetStmt(2, "a = 1", "a", "int64", false),
		BranchElseStmt(3, "else"),
		LetStmt(4, "b = 2", "b", "int64", false),
		BranchExitStmt(5, "endif"),
	}

	cfg := BuildCFG(stmts)
	pairs := cfg.GetBranchPairs()

	if len(pairs) != 1 {
		t.Errorf("Should have 1 branch pair, got %d", len(pairs))
	}

	if len(pairs) > 0 {
		pair := pairs[0]
		if len(pair[0]) == 0 || len(pair[1]) == 0 {
			t.Error("Branch pair should have non-empty if and else bodies")
		}
	}
}
