package core

import (
	"testing"
)

func TestPrefixContext_AddAndGetVariable(t *testing.T) {
	ctx := NewPrefixContext("test")
	ctx.AddVariable("x", VarTypeInt, 42, false, Position{Line: 1, Column: 1})

	v, ok := ctx.GetVariable("x")
	if !ok {
		t.Fatal("variable 'x' should exist")
	}
	if v.Value != 42 {
		t.Errorf("value = %v, want 42", v.Value)
	}
	if v.Type != VarTypeInt {
		t.Errorf("type = %v, want int", v.Type)
	}
}

func TestPrefixContext_GetVariable_NotFound(t *testing.T) {
	ctx := NewPrefixContext("test")
	_, ok := ctx.GetVariable("nonexistent")
	if ok {
		t.Error("should not find nonexistent variable")
	}
}

func TestPrefixContext_ParentLookup(t *testing.T) {
	parent := NewPrefixContext("parent")
	parent.AddVariable("inherited", VarTypeString, "hello", false, Position{})

	child := NewPrefixContext("child")
	child.SetParent(parent)

	v, ok := child.GetVariable("inherited")
	if !ok {
		t.Fatal("child should find variable from parent")
	}
	if v.Value != "hello" {
		t.Errorf("inherited value = %v, want 'hello'", v.Value)
	}
}

func TestPrefixContext_HasPrefixVar(t *testing.T) {
	ctx := NewPrefixContext("test")
	ctx.AddVariable("$x", VarTypeInt, 10, true, Position{})
	ctx.AddVariable("y", VarTypeInt, 20, false, Position{})

	if !ctx.HasPrefixVar("$x") {
		t.Error("$x should be a prefix var")
	}
	if ctx.HasPrefixVar("y") {
		t.Error("y should not be a prefix var")
	}
}

func TestPrefixContext_HasPrefixVar_ParentLookup(t *testing.T) {
	parent := NewPrefixContext("parent")
	parent.AddVariable("$p", VarTypeInt, 1, true, Position{})

	child := NewPrefixContext("child")
	child.SetParent(parent)

	if !child.HasPrefixVar("$p") {
		t.Error("child should find prefix var from parent")
	}
}

func TestPrefixContext_GetAllVariables(t *testing.T) {
	ctx := NewPrefixContext("test")
	ctx.AddVariable("a", VarTypeInt, 1, false, Position{})
	ctx.AddVariable("b", VarTypeFloat, 2.0, false, Position{})

	vars := ctx.GetAllVariables()
	if len(vars) != 2 {
		t.Errorf("GetAllVariables() = %d, want 2", len(vars))
	}
}

func TestPrefixContext_GetAllVariables_IncludesParent(t *testing.T) {
	parent := NewPrefixContext("parent")
	parent.AddVariable("p1", VarTypeInt, 1, false, Position{})

	child := NewPrefixContext("child")
	child.SetParent(parent)
	child.AddVariable("c1", VarTypeInt, 2, false, Position{})

	vars := child.GetAllVariables()
	if len(vars) != 2 {
		t.Errorf("GetAllVariables() = %d, want 2 (child + parent)", len(vars))
	}
}

func TestPrefixManager_CreatePrefix_Duplicate(t *testing.T) {
	pm := NewPrefixManager()
	_, err := pm.CreatePrefix("myprefix", PrefixAnnotationPrefix)
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	_, err = pm.CreatePrefix("myprefix", PrefixAnnotationPrefix)
	if err == nil {
		t.Error("duplicate prefix creation should fail")
	}
}

func TestPrefixManager_GetPrefix(t *testing.T) {
	pm := NewPrefixManager()
	pm.CreatePrefix("test", PrefixAnnotationPrefix)

	ctx := pm.GetPrefix("test")
	if ctx == nil {
		t.Fatal("GetPrefix should return the context")
	}
	if ctx.Name != "test" {
		t.Errorf("name = %q, want 'test'", ctx.Name)
	}

	missing := pm.GetPrefix("nonexistent")
	if missing != nil {
		t.Error("GetPrefix for nonexistent should return nil")
	}
}

func TestPrefixManager_SetActiveContext(t *testing.T) {
	pm := NewPrefixManager()
	pm.CreatePrefix("active_test", PrefixAnnotationPrefix)

	err := pm.SetActiveContext("active_test")
	if err != nil {
		t.Fatalf("SetActiveContext failed: %v", err)
	}

	active := pm.GetActiveContext()
	if active.Name != "active_test" {
		t.Errorf("active context = %q, want 'active_test'", active.Name)
	}
}

func TestPrefixManager_SetActiveContext_NotFound(t *testing.T) {
	pm := NewPrefixManager()
	err := pm.SetActiveContext("nonexistent")
	if err == nil {
		t.Error("SetActiveContext for nonexistent should fail")
	}
}

func TestPrefixManager_PushPopContext(t *testing.T) {
	pm := NewPrefixManager()

	ctx, err := pm.PushContext("inner")
	if err != nil {
		t.Fatalf("PushContext failed: %v", err)
	}
	if ctx.Name != "inner" {
		t.Errorf("pushed context name = %q, want 'inner'", ctx.Name)
	}

	active := pm.GetActiveContext()
	if active.Name != "inner" {
		t.Errorf("active after push = %q, want 'inner'", active.Name)
	}

	err = pm.PopContext()
	if err != nil {
		t.Fatalf("PopContext failed: %v", err)
	}

	active = pm.GetActiveContext()
	if active.Name != "root" {
		t.Errorf("active after pop = %q, want 'root'", active.Name)
	}
}

func TestPrefixManager_PopContext_RootError(t *testing.T) {
	pm := NewPrefixManager()
	err := pm.PopContext()
	if err == nil {
		t.Error("popping root context should fail")
	}
}

func TestPrefixManager_SetAndGetVariable(t *testing.T) {
	pm := NewPrefixManager()
	pm.CreatePrefix("ns", PrefixAnnotationPrefix)

	pos := Position{Line: 1, Column: 1}
	err := pm.SetVariable("ns", "count", VarTypeInt, 42, false, pos)
	if err != nil {
		t.Fatalf("SetVariable failed: %v", err)
	}

	v, ok := pm.GetVariable("ns", "count")
	if !ok {
		t.Fatal("GetVariable should find 'count'")
	}
	if v.Value != 42 {
		t.Errorf("value = %v, want 42", v.Value)
	}
}

func TestPrefixManager_SetVariable_PrefixNotFound(t *testing.T) {
	pm := NewPrefixManager()
	err := pm.SetVariable("nonexistent", "x", VarTypeInt, 0, false, Position{})
	if err == nil {
		t.Error("SetVariable for nonexistent prefix should fail")
	}
}

func TestPrefixManager_ListPrefixes(t *testing.T) {
	pm := NewPrefixManager()
	pm.CreatePrefix("alpha", PrefixAnnotationPrefix)
	pm.CreatePrefix("beta", PrefixAnnotationTree)

	prefixes := pm.ListPrefixes()
	// root + alpha + beta = 3
	if len(prefixes) != 3 {
		t.Errorf("ListPrefixes() = %d, want 3", len(prefixes))
	}
}

func TestPrefixAnnotation_String(t *testing.T) {
	tests := []struct {
		ann  PrefixAnnotation
		want string
	}{
		{PrefixAnnotationNone, "none"},
		{PrefixAnnotationPrefix, "prefix"},
		{PrefixAnnotationTree, "tree"},
		{PrefixAnnotationPrefixTree, "prefix,tree"},
	}
	for _, tt := range tests {
		if got := tt.ann.String(); got != tt.want {
			t.Errorf("PrefixAnnotation(%d).String() = %q, want %q", tt.ann, got, tt.want)
		}
	}
}

func TestPrefixVarType_String(t *testing.T) {
	tests := []struct {
		vt   PrefixVarType
		want string
	}{
		{VarTypeUnknown, "unknown"},
		{VarTypeInt, "int"},
		{VarTypeFloat, "float"},
		{VarTypeString, "string"},
		{VarTypeBool, "bool"},
		{VarTypeObject, "object"},
	}
	for _, tt := range tests {
		if got := tt.vt.String(); got != tt.want {
			t.Errorf("PrefixVarType(%d).String() = %q, want %q", tt.vt, got, tt.want)
		}
	}
}

func TestPrefixVariable_String(t *testing.T) {
	pv := &PrefixVariable{Name: "x", Type: VarTypeInt, IsPrefixVar: true}
	s := pv.String()
	if s != "$x:int" {
		t.Errorf("PrefixVariable.String() = %q, want \"$x:int\"", s)
	}

	pv2 := &PrefixVariable{Name: "y", Type: VarTypeString, IsPrefixVar: false}
	s2 := pv2.String()
	if s2 != "y:string" {
		t.Errorf("PrefixVariable.String() = %q, want \"y:string\"", s2)
	}
}
