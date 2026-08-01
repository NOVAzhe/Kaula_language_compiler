package core

import (
	"testing"
)

func TestParseAnnotation(t *testing.T) {
	tests := []struct {
		input string
		want  TreeAnnotation
	}{
		{"prefix", AnnotationPrefix},
		{"tree", AnnotationTree},
		{"prefix,tree", AnnotationPrefixTree},
		{"tree,prefix", AnnotationPrefixTree},
		{"root", AnnotationRoot},
		{"root,tree", AnnotationRootTree},
		{"tree,root", AnnotationRootTree},
		{"unknown", AnnotationNone},
		{"", AnnotationNone},
	}

	for _, tt := range tests {
		got := ParseAnnotation(tt.input)
		if got != tt.want {
			t.Errorf("ParseAnnotation(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestTreeAnnotation_String(t *testing.T) {
	tests := []struct {
		ann  TreeAnnotation
		want string
	}{
		{AnnotationNone, "none"},
		{AnnotationPrefix, "prefix"},
		{AnnotationTree, "tree"},
		{AnnotationPrefixTree, "prefix,tree"},
		{AnnotationRoot, "root"},
		{AnnotationRootTree, "root,tree"},
	}
	for _, tt := range tests {
		if got := tt.ann.String(); got != tt.want {
			t.Errorf("TreeAnnotation(%d).String() = %q, want %q", tt.ann, got, tt.want)
		}
	}
}

func TestTreeNodeType_String(t *testing.T) {
	tests := []struct {
		nt   TreeNodeType
		want string
	}{
		{NodeTypeGeneric, "generic"},
		{NodeTypeStatement, "statement"},
		{NodeTypeFunction, "function"},
		{NodeTypeVariable, "variable"},
		{NodeTypeBlock, "block"},
		{NodeTypeCondition, "condition"},
		{NodeTypeLoop, "loop"},
		{NodeTypeReturn, "return"},
	}
	for _, tt := range tests {
		if got := tt.nt.String(); got != tt.want {
			t.Errorf("TreeNodeType(%d).String() = %q, want %q", tt.nt, got, tt.want)
		}
	}
}

func TestTreeNode_AddChild(t *testing.T) {
	parent := NewTreeNode("parent", NodeTypeFunction)
	child := NewTreeNode("child", NodeTypeVariable)

	parent.AddChild(child)

	children := parent.GetChildren()
	if len(children) != 1 {
		t.Fatalf("children count = %d, want 1", len(children))
	}
	if children[0].Name != "child" {
		t.Errorf("child name = %q, want 'child'", children[0].Name)
	}
	if child.Parent != parent {
		t.Error("child.Parent should point to parent")
	}
}

func TestTreeNode_RemoveChild(t *testing.T) {
	parent := NewTreeNode("parent", NodeTypeFunction)
	child := NewTreeNode("child", NodeTypeVariable)
	parent.AddChild(child)

	ok := parent.RemoveChild(child)
	if !ok {
		t.Error("RemoveChild should return true")
	}
	if len(parent.GetChildren()) != 0 {
		t.Error("children should be empty after removal")
	}
	if child.Parent != nil {
		t.Error("removed child's Parent should be nil")
	}
}

func TestTreeNode_RemoveChild_NotFound(t *testing.T) {
	parent := NewTreeNode("parent", NodeTypeFunction)
	other := NewTreeNode("other", NodeTypeVariable)

	ok := parent.RemoveChild(other)
	if ok {
		t.Error("RemoveChild should return false for non-child")
	}
}

func TestTreeNode_MatchesConstraint_NilConstraint(t *testing.T) {
	node := NewTreeNode("x", NodeTypeVariable)
	if !node.MatchesConstraint() {
		t.Error("node with nil constraint should match")
	}
}

func TestTreeNode_MatchesConstraint_RequiredType(t *testing.T) {
	node := NewTreeNode("x", NodeTypeVariable)
	constraint := &TreeConstraint{
		Required: true,
		NodeType: NodeTypeVariable,
	}
	node.SetConstraint(constraint)

	if !node.MatchesConstraint() {
		t.Error("node should match constraint with matching type")
	}

	wrongConstraint := &TreeConstraint{
		Required: true,
		NodeType: NodeTypeFunction,
	}
	node.SetConstraint(wrongConstraint)
	if node.MatchesConstraint() {
		t.Error("node should not match constraint with wrong type")
	}
}

func TestTreeNode_MatchesConstraint_Pattern(t *testing.T) {
	node := NewTreeNode("myFunc", NodeTypeFunction)
	constraint := &TreeConstraint{
		Pattern: "myFunc",
	}
	node.SetConstraint(constraint)

	if !node.MatchesConstraint() {
		t.Error("node should match constraint with matching pattern")
	}

	wrongPattern := &TreeConstraint{
		Pattern: "otherFunc",
	}
	node.SetConstraint(wrongPattern)
	if node.MatchesConstraint() {
		t.Error("node should not match constraint with wrong pattern")
	}
}

func TestTreeNode_MatchesConstraint_Children(t *testing.T) {
	parent := NewTreeNode("parent", NodeTypeFunction)
	child1 := NewTreeNode("arg1", NodeTypeParameter)
	child2 := NewTreeNode("arg2", NodeTypeParameter)
	parent.AddChild(child1)
	parent.AddChild(child2)

	constraint := &TreeConstraint{
		Required: true,
		NodeType: NodeTypeFunction,
		Children: []*TreeConstraint{
			{Required: true, NodeType: NodeTypeParameter},
			{Required: true, NodeType: NodeTypeParameter},
		},
	}
	parent.SetConstraint(constraint)

	if !parent.MatchesConstraint() {
		t.Error("parent with matching children should match")
	}
}

func TestTreeNode_MatchesConstraint_TooFewChildren(t *testing.T) {
	parent := NewTreeNode("parent", NodeTypeFunction)
	child := NewTreeNode("arg1", NodeTypeParameter)
	parent.AddChild(child)

	constraint := &TreeConstraint{
		Children: []*TreeConstraint{
			{Required: true, NodeType: NodeTypeParameter},
			{Required: true, NodeType: NodeTypeParameter},
		},
	}
	parent.SetConstraint(constraint)

	if parent.MatchesConstraint() {
		t.Error("parent with too few children should not match")
	}
}

func TestNewTree(t *testing.T) {
	tree := NewTree()
	if tree == nil {
		t.Fatal("NewTree() should not return nil")
	}
	if tree.Root == nil {
		t.Error("tree.Root should not be nil")
	}
}
