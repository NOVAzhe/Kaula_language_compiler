package codegen

import (
	"kaula-compiler/internal/ast"
	"strings"
	"testing"
)

// ---- splitTopLevelCommas ----

func TestSplitTopLevelCommas_Simple(t *testing.T) {
	parts := splitTopLevelCommas("a, b, c")
	if len(parts) != 3 {
		t.Fatalf("got %d parts, want 3: %v", len(parts), parts)
	}
	if strings.TrimSpace(parts[0]) != "a" {
		t.Errorf("parts[0] = %q, want %q", parts[0], "a")
	}
}

func TestSplitTopLevelCommas_Nested(t *testing.T) {
	// nested void(...) should not split on inner commas
	parts := splitTopLevelCommas("void(void(i32)i32, i32)")
	if len(parts) != 1 {
		t.Fatalf("got %d parts, want 1 (no top-level commas): %v", len(parts), parts)
	}
}

func TestSplitTopLevelCommas_WithAngleBrackets(t *testing.T) {
	parts := splitTopLevelCommas("Box<int>, Box<string>")
	if len(parts) != 2 {
		t.Fatalf("got %d parts, want 2: %v", len(parts), parts)
	}
}

func TestSplitTopLevelCommas_WithBrackets(t *testing.T) {
	parts := splitTopLevelCommas("[]int, []string")
	if len(parts) != 2 {
		t.Fatalf("got %d parts, want 2: %v", len(parts), parts)
	}
}

func TestSplitTopLevelCommas_Empty(t *testing.T) {
	parts := splitTopLevelCommas("")
	if len(parts) != 1 || parts[0] != "" {
		t.Fatalf("got %v, want [\"\"]", parts)
	}
}

// ---- kaulaStructTag ----

func TestKaulaStructTag_Prefix(t *testing.T) {
	got := kaulaStructTag("Point")
	want := "K_Point"
	if got != want {
		t.Errorf("kaulaStructTag(Point) = %q, want %q", got, want)
	}
}

func TestKaulaStructTag_SpecialChars(t *testing.T) {
	got := kaulaStructTag("MyClass")
	want := "K_MyClass"
	if got != want {
		t.Errorf("kaulaStructTag(MyClass) = %q, want %q", got, want)
	}
}

// ---- generateStructAttributes ----

func TestGenerateStructAttributes_Empty(t *testing.T) {
	got := generateStructAttributes(nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestGenerateStructAttributes_Packed(t *testing.T) {
	attrs := []*ast.Attribute{{Name: "packed"}}
	got := generateStructAttributes(attrs)
	if !strings.Contains(got, "__attribute__((packed))") {
		t.Errorf("expected packed attribute, got %q", got)
	}
}

func TestGenerateStructAttributes_Aligned(t *testing.T) {
	attrs := []*ast.Attribute{{Name: "aligned", Args: []string{"16"}}}
	got := generateStructAttributes(attrs)
	if !strings.Contains(got, "__attribute__((aligned(16)))") {
		t.Errorf("expected aligned(16), got %q", got)
	}
}

func TestGenerateStructAttributes_AlignedNoArg(t *testing.T) {
	attrs := []*ast.Attribute{{Name: "aligned"}}
	got := generateStructAttributes(attrs)
	if !strings.Contains(got, "__attribute__((aligned))") {
		t.Errorf("expected aligned (no arg), got %q", got)
	}
}

// ---- parseVoidSignatureType ----

func TestParseVoidSignatureType_DataPtr(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}
	got, ok := tg.parseVoidSignatureType("void()")
	if !ok {
		t.Fatal("parseVoidSignatureType(void()) should succeed")
	}
	if got != "void*" {
		t.Errorf("got %q, want %q", got, "void*")
	}
}

func TestParseVoidSignatureType_ConstDataPtr(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}
	// parseVoidSignatureType only matches types starting with "void(";
	// "const void()" is handled by the const branch in MapKaulaTypeToC.
	got, ok := tg.parseVoidSignatureType("const void()")
	if ok {
		t.Fatalf("parseVoidSignatureType(const void()) should not match; got %q", got)
	}
	// Verify the full pipeline handles it correctly via MapKaulaTypeToC
	full := tg.MapKaulaTypeToC("const void()")
	if full != "const void*" {
		t.Errorf("MapKaulaTypeToC(const void()) = %q, want %q", full, "const void*")
	}
}

func TestParseVoidSignatureType_FuncPtr(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}
	got, ok := tg.parseVoidSignatureType("void(void())void")
	if !ok {
		t.Fatal("parseVoidSignatureType(void(void())void) should succeed")
	}
	if got != "void (*)(void*)" {
		t.Errorf("got %q, want %q", got, "void (*)(void*)")
	}
}

func TestParseVoidSignatureType_NonVoid(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}
	got, ok := tg.parseVoidSignatureType("int")
	if ok {
		t.Fatal("parseVoidSignatureType(int) should not match")
	}
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// ---- MapKaulaTypeToC ----

func TestMapKaulaTypeToC_BasicTypes(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	tests := []struct {
		kaula string
		want  string
	}{
		{"int", "int64_t"},
		{"i64", "int64_t"},
		{"i32", "int32_t"},
		{"i16", "int16_t"},
		{"i8", "int8_t"},
		{"u64", "uint64_t"},
		{"u32", "uint32_t"},
		{"u16", "uint16_t"},
		{"u8", "uint8_t"},
		{"float", "float"},
		{"f32", "float"},
		{"double", "double"},
		{"f64", "double"},
		{"bool", "bool"},
		{"char", "char"},
		{"byte", "uint8_t"},
		{"string", "String"},
		{"cstring", "const char*"},
		{"cint", "int"},
		{"void", "void"},
		{"size", "size_t"},
		{"object", "Object*"},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := tg.MapKaulaTypeToC(tt.kaula)
			if got != tt.want {
				t.Errorf("MapKaulaTypeToC(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}

func TestMapKaulaTypeToC_PointerTypes(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	tests := []struct {
		kaula string
		want  string
	}{
		{"*int", "int64_t*"},
		{"*i64", "int64_t*"},
		{"*float", "float*"},
		{"*double", "double*"},
		{"*bool", "bool*"},
		{"*char", "char*"},
		{"*string", "String*"},
		{"int*", "int64_t*"},
		{"float*", "float*"},
		{"string*", "String*"},
		{"ptr", "ptr"},
		{"int64ptr", "int64ptr"},
		{"*cstring", "const char**"},
		{"cstring*", "const char**"},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := tg.MapKaulaTypeToC(tt.kaula)
			if got != tt.want {
				t.Errorf("MapKaulaTypeToC(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}

func TestMapKaulaTypeToC_ArrayTypes(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	tests := []struct {
		kaula string
		want  string
	}{
		{"[10]int", "int64_t[10]"},
		{"[5]byte", "uint8_t[5]"},
		{"[3]float", "float[3]"},
		{"[]int", "int64_t*"},
		{"[]byte", "uint8_t*"},
		{"[]string", "String*"},
		{"[]cstring", "const char**"},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := tg.MapKaulaTypeToC(tt.kaula)
			if got != tt.want {
				t.Errorf("MapKaulaTypeToC(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}

func TestMapKaulaTypeToC_ConstTypes(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	tests := []struct {
		kaula string
		want  string
	}{
		{"const int", "const int64_t"},
		{"const string", "String"},
		{"const void()", "const void*"},
		{"const char", "const char"},
		{"const double", "const double"},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := tg.MapKaulaTypeToC(tt.kaula)
			if got != tt.want {
				t.Errorf("MapKaulaTypeToC(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}

func TestMapKaulaTypeToC_StructType(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   map[string]bool{"Point": true},
		activeTypeMap: nil,
	}

	got := tg.MapKaulaTypeToC("Point")
	want := "K_Point"
	if got != want {
		t.Errorf("MapKaulaTypeToC(Point) = %q, want %q", got, want)
	}
}

func TestMapKaulaTypeToC_ClassType(t *testing.T) {
	// Class types resolve to K_ClassName* (pointer)
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   map[string]bool{"MyClass": true},
		activeTypeMap: nil,
		codegen: &CodeGenerator{
			genericCache:        make(map[string]*GenericInstanceCache),
			genericInstantiated: make(map[string]bool),
			genericTypeCache:    make(map[string]string),
			callStack:           make(map[string]bool),
			constTable:          make(map[string]string),
			arrayLens:           make(map[string]int),
			trackedModules:      make(map[string]bool),
			usedThirdPartyLibs:  make(map[string]bool),
			localImportFuncs:    make(map[string]bool),
		},
	}

	// IsClassType returns false by default for CodeGenerator
	// So the struct branch will be taken -> K_MyClass (without *)
	got := tg.MapKaulaTypeToC("MyClass")
	want := "K_MyClass"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// ---- GetTypeSize ----

func TestGetTypeSize_BasicTypes(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	tests := []struct {
		kaula string
		want  int
	}{
		{"i8", 1}, {"int8", 1}, {"byte", 1}, {"bool", 1}, {"char", 1},
		{"i16", 2}, {"int16", 2}, {"short", 2},
		{"i32", 4}, {"int32", 4}, {"int", 4}, {"float", 4}, {"f32", 4},
		{"i64", 8}, {"int64", 8}, {"long", 8}, {"double", 8}, {"f64", 8},
		{"string", 16}, {"str", 16},
		{"cstring", 8},
		{"*int", 8},
		{"int*", 8},
		{"void()", 8},
		{"void(void())void", 8},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got, ok := tg.GetTypeSize(tt.kaula)
			if !ok {
				t.Errorf("GetTypeSize(%q) returned ok=false", tt.kaula)
				return
			}
			if got != tt.want {
				t.Errorf("GetTypeSize(%q) = %d, want %d", tt.kaula, got, tt.want)
			}
		})
	}
}

func TestGetTypeSize_Array(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	got, ok := tg.GetTypeSize("[10]int64")
	if !ok {
		t.Fatal("GetTypeSize([10]int64) returned ok=false")
	}
	if got != 80 {
		t.Errorf("GetTypeSize([10]int64) = %d, want 80", got)
	}
}

func TestGetTypeSize_Unknown(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	_, ok := tg.GetTypeSize("MyStruct")
	if ok {
		t.Error("GetTypeSize(MyStruct) should return ok=false for unknown type")
	}
}

// ---- convertType ----

func TestConvertType_Nullable(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   map[string]string{"File": "FILE*"},
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	got := tg.convertType("int", true)
	want := "int64_t*"
	if got != want {
		t.Errorf("convertType(int, nullable=true) = %q, want %q", got, want)
	}
}

func TestConvertType_ClibType(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   map[string]string{"File": "FILE*"},
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	got := tg.convertType("File", false)
	want := "FILE*"
	if got != want {
		t.Errorf("convertType(File) = %q, want %q", got, want)
	}
}

// ---- GetTypeSize edge cases ----

func TestGetTypeSize_EmptyType(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	_, ok := tg.GetTypeSize("")
	if ok {
		t.Error("GetTypeSize('') should return ok=false")
	}
}

// ---- methodHasReturn ----

func TestMethodHasReturn_WithReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 0}},
	}
	if !methodHasReturn(stmts) {
		t.Error("methodHasReturn should find return statement")
	}
}

func TestMethodHasReturn_WithoutReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.VariableDeclaration{Name: "x", Type: "int"},
	}
	if methodHasReturn(stmts) {
		t.Error("methodHasReturn should not find return statement")
	}
}

func TestMethodHasReturn_Nil(t *testing.T) {
	if methodHasReturn(nil) {
		t.Error("methodHasReturn(nil) should be false")
	}
}

func TestMethodHasReturn_BlockWithReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 42}},
			},
		},
	}
	if !methodHasReturn(stmts) {
		t.Error("methodHasReturn should find return inside block")
	}
}

func TestMethodHasReturn_IfWithReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IfStatement{
			Body: []ast.Statement{
				&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 1}},
			},
		},
	}
	if !methodHasReturn(stmts) {
		t.Error("methodHasReturn should find return inside if body")
	}
}

// ---- tgReturnTypeToC ----

func TestTgReturnTypeToC(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
		codegen: &CodeGenerator{
			genericCache:        make(map[string]*GenericInstanceCache),
			genericInstantiated: make(map[string]bool),
			genericTypeCache:    make(map[string]string),
			callStack:           make(map[string]bool),
			constTable:          make(map[string]string),
			arrayLens:           make(map[string]int),
			trackedModules:      make(map[string]bool),
			usedThirdPartyLibs:  make(map[string]bool),
			localImportFuncs:    make(map[string]bool),
		},
	}

	tests := []struct {
		kaula string
		want  string
	}{
		{"", "void"},
		{"int", "int"},
		{"i64", "int64_t"},
		{"u64", "uint64_t"},
		{"i32", "int32_t"},
		{"u32", "uint32_t"},
		{"i16", "int16_t"},
		{"u16", "uint16_t"},
		{"i8", "int8_t"},
		{"u8", "uint8_t"},
		{"float", "float"},
		{"f32", "float"},
		{"double", "double"},
		{"f64", "double"},
		{"bool", "bool"},
		{"char", "char"},
		{"void", "void"},
		{"string", "String"},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := tgReturnTypeToC(tg, tt.kaula)
			if got != tt.want {
				t.Errorf("tgReturnTypeToC(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}

// ---- Generic Type Cache ----

func TestGenericTypeCache_InitialState(t *testing.T) {
	cg := &CodeGenerator{
		genericCache:        make(map[string]*GenericInstanceCache),
		genericInstantiated: make(map[string]bool),
		genericTypeCache:    make(map[string]string),
		callStack:           make(map[string]bool),
		constTable:          make(map[string]string),
		arrayLens:           make(map[string]int),
		trackedModules:      make(map[string]bool),
		usedThirdPartyLibs:  make(map[string]bool),
		localImportFuncs:    make(map[string]bool),
	}

	if len(cg.genericTypeCache) != 0 {
		t.Error("genericTypeCache should be empty initially")
	}
}

// ---- convertTypeAliasToCType ----

func TestConvertTypeAliasToCType(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	tests := []struct {
		kaula string
		want  string
	}{
		{"string", "char*"},
		{"*string", "char**"},
		{"*int", "int*"},
		{"*float", "float*"},
		{"*double", "double*"},
		{"*bool", "bool*"},
		{"*char", "char*"},
		{"[]string", "char**"},
		{"void()", "void*"},
		{"void(void())void", "void (*)(void*)"},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := tg.convertTypeAliasToCType(tt.kaula)
			if got != tt.want {
				t.Errorf("convertTypeAliasToCType(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}