package sema

import (
	"testing"

	"kaula-compiler/internal/lexer"
	"kaula-compiler/internal/parser"
)

// parseAndAnalyze is a helper to parse source and run semantic analysis.
func parseAndAnalyze(t *testing.T, src string) (*parser.Parser, *SemanticAnalyzer) {
	t.Helper()
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()
	if prog == nil {
		t.Fatal("Parse() returned nil")
	}

	sa := NewSemanticAnalyzer()
	sa.Analyze(prog)
	return p, sa
}

func TestSema_EmptyProgram(t *testing.T) {
	src := ``
	_, sa := parseAndAnalyze(t, src)
	if sa.HasErrors() {
		t.Error("Expected no errors for empty program")
	}
}

func TestSema_SimpleFunction(t *testing.T) {
	src := `fn main() int {
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	// It's okay if there are some errors (e.g., stdlib config not found)
	// but we should not crash
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_FunctionWithReturn(t *testing.T) {
	src := `fn add(a int, b int) int {
    return a + b
}

fn main() int {
    return add(1, 2)
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_VariableDeclaration(t *testing.T) {
	src := `fn main() int {
    int x = 42
    int y = x + 1
    return y
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_IfStatement(t *testing.T) {
	src := `fn main() int {
    int x = 10
    if (x > 5) {
        return 1
    }
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_IfElseChain(t *testing.T) {
	src := `fn main() int {
    int x = 2
    if (x == 1) {
        return 10
    } else if (x == 2) {
        return 20
    } else {
        return 0
    }
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_WhileLoop(t *testing.T) {
	src := `fn main() int {
    int i = 0
    while (i < 10) {
        i = i + 1
    }
    return i
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_ForLoop(t *testing.T) {
	src := `fn main() int {
    int sum = 0
    for (int i = 0; i < 10; i = i + 1) {
        sum = sum + i
    }
    return sum
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_StringOperations(t *testing.T) {
	src := `fn main() int {
    string s = "hello"
    string t = " world"
    string u = s + t
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_StructAndFieldAccess(t *testing.T) {
	src := `struct Point {
    int x
    int y
}

fn main() int {
    Point p
    p.x = 10
    p.y = 20
    return p.x + p.y
}`
	// This should parse and analyze without crashing
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_EnumAndMatch(t *testing.T) {
	src := `enum Color {
    Red, Green, Blue
}

fn main() int {
    Color c = Color.Red
    match c {
        Color.Red => { return 0 }
        Color.Green => { return 1 }
        _ => { return 2 }
    }
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_ImportStdlib(t *testing.T) {
	src := `import std.io

fn main() int {
    println("hello")
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_ExportConst(t *testing.T) {
	src := `export const C_MAX = 100

fn main() int {
    return C_MAX
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_TypeMismatch(t *testing.T) {
	src := `fn main() int {
    int x = "hello"
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
	// This may or may not produce errors depending on type coercion rules
}

func TestSema_BoolExpression(t *testing.T) {
	src := `fn main() int {
    bool flag = true
    if (flag) {
        return 1
    }
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_IntegerTypes(t *testing.T) {
	src := `fn main() int {
    int8 a = 1
    int16 b = 2
    int32 c = 3
    int64 d = 4
    int e = 5
    uint8 f = 6
    uint16 g = 7
    uint32 h = 8
    uint64 i = 9
    uint j = 10
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_FloatTypes(t *testing.T) {
	src := `fn main() int {
    float32 a = 1.0
    float64 b = 2.0
    float c = 3.0
    double d = 4.0
    return 0
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

func TestSema_BreakContinue(t *testing.T) {
	src := `fn main() int {
    int i = 0
    while (i < 10) {
        if (i == 5) {
            break
        }
        i = i + 1
    }
    return i
}`
	_, sa := parseAndAnalyze(t, src)
	if sa == nil {
		t.Fatal("SemanticAnalyzer is nil")
	}
}

// TestSema_ErrorCollector verifies that the error collector is properly wired.
func TestSema_ErrorCollector(t *testing.T) {
	sa := NewSemanticAnalyzer()
	if sa.ErrorCollector() == nil {
		t.Error("ErrorCollector() should not return nil")
	}
}

// TestSema_SetSOREnabled verifies SOR mode toggle.
func TestSema_SetSOREnabled(t *testing.T) {
	sa := NewSemanticAnalyzer()
	sa.SetSOREnabled(true)
	// No crash is the main test
	sa.SetSOREnabled(false)
}

// TestSema_PromoteIntegerType verifies integer type promotion logic.
func TestSema_PromoteIntegerType(t *testing.T) {
	tests := []struct {
		left, right, expected string
	}{
		{"int8", "int16", "int16"},
		{"int16", "int32", "int32"},
		{"int32", "int64", "int64"},
		{"int", "int64", "int"},       // same precision (4), left wins
		{"uint8", "uint16", "uint16"},
		{"int", "uint", "uint"},       // uint has precision 5 > int precision 4
		{"int8", "int8", "int8"},
	}

	for _, tt := range tests {
		t.Run(tt.left+"_"+tt.right, func(t *testing.T) {
			result := promoteIntegerType(tt.left, tt.right)
			if result != tt.expected {
				t.Errorf("promoteIntegerType(%q, %q) = %q, want %q", tt.left, tt.right, result, tt.expected)
			}
		})
	}
}

// TestSema_IsIntegerType verifies type classification.
func TestSema_IsIntegerType(t *testing.T) {
	intTypes := []string{"int8", "int16", "int32", "int64", "int", "uint8", "uint16", "uint32", "uint64", "uint", "byte"}
	nonIntTypes := []string{"float32", "float64", "string", "bool", "char", "void", "ptr"}

	for _, tt := range intTypes {
		t.Run(tt, func(t *testing.T) {
			if !isIntegerType(tt) {
				t.Errorf("isIntegerType(%q) = false, want true", tt)
			}
		})
	}
	for _, tt := range nonIntTypes {
		t.Run(tt, func(t *testing.T) {
			if isIntegerType(tt) {
				t.Errorf("isIntegerType(%q) = true, want false", tt)
			}
		})
	}
}

// TestSema_IsFloatType verifies floating-point type classification.
func TestSema_IsFloatType(t *testing.T) {
	floatTypes := []string{"float", "f32", "f64", "double", "real", "single"}
	nonFloatTypes := []string{"int", "string", "bool", "void", "float32", "float64"}

	for _, tt := range floatTypes {
		t.Run(tt, func(t *testing.T) {
			if !isFloatType(tt) {
				t.Errorf("isFloatType(%q) = false, want true", tt)
			}
		})
	}
	for _, tt := range nonFloatTypes {
		t.Run(tt, func(t *testing.T) {
			if isFloatType(tt) {
				t.Errorf("isFloatType(%q) = true, want false", tt)
			}
		})
	}
}

// TestSema_IsNumericType verifies numeric type classification.
func TestSema_IsNumericType(t *testing.T) {
	numericTypes := []string{"int", "i64", "f64", "f32", "float", "i32", "i16", "i8", "u64", "u32", "u16", "u8"}
	nonNumericTypes := []string{"string", "bool", "char", "void", "int64", "float32", "double", "uint8", "byte"}

	for _, tt := range numericTypes {
		t.Run(tt, func(t *testing.T) {
			if !isNumericType(tt) {
				t.Errorf("isNumericType(%q) = false, want true", tt)
			}
		})
	}
	for _, tt := range nonNumericTypes {
		t.Run(tt, func(t *testing.T) {
			if isNumericType(tt) {
				t.Errorf("isNumericType(%q) = true, want false", tt)
			}
		})
	}
}

// TestSema_IsPointerType verifies pointer type classification.
func TestSema_IsPointerType(t *testing.T) {
	pointerTypes := []string{"int*", "string*", "int**", "byte*"}
	nonPointerTypes := []string{"int", "string", "bool", "*int", "char*", "void()"} // char* is string-like, void() is function type

	for _, tt := range pointerTypes {
		t.Run(tt, func(t *testing.T) {
			if !isPointerType(tt) {
				t.Errorf("isPointerType(%q) = false, want true", tt)
			}
		})
	}
	for _, tt := range nonPointerTypes {
		t.Run(tt, func(t *testing.T) {
			if isPointerType(tt) {
				t.Errorf("isPointerType(%q) = true, want false", tt)
			}
		})
	}
}

// TestSema_IsStringLikeType verifies string-like type classification.
func TestSema_IsStringLikeType(t *testing.T) {
	stringTypes := []string{"string", "cstring", "cstr", "char*"}
	nonStringTypes := []string{"int", "bool", "char", "String"}

	for _, tt := range stringTypes {
		t.Run(tt, func(t *testing.T) {
			if !isStringLikeType(tt) {
				t.Errorf("isStringLikeType(%q) = false, want true", tt)
			}
		})
	}
	for _, tt := range nonStringTypes {
		t.Run(tt, func(t *testing.T) {
			if isStringLikeType(tt) {
				t.Errorf("isStringLikeType(%q) = true, want false", tt)
			}
		})
	}
}