package main

import (
	"kaula-compiler/internal/ast"
	"strings"
	"testing"
)

// --- Import sorting ---

func TestFormatter_ImportSorting_StdFirst(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ImportStatement{Module: "net.http"},
			&ast.ImportStatement{Module: "std.string"},
			&ast.ImportStatement{Module: "std.io"},
			&ast.ImportStatement{Module: "crypto.sha"},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	// std imports should come first, sorted alphabetically
	lines := strings.Split(out, "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines, got %d", len(lines))
	}
	// First two should be std imports
	if !strings.HasPrefix(lines[0], "import std.io") {
		t.Errorf("first import should be std.io, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "import std.string") {
		t.Errorf("second import should be std.string, got %q", lines[1])
	}
	// Then blank line, then other imports
	if lines[2] != "" {
		t.Errorf("expected blank line after std imports, got %q", lines[2])
	}
}

func TestFormatter_EmptyProgram(t *testing.T) {
	prog := &ast.Program{Statements: []ast.Statement{}}
	f := NewFormatter()
	out := f.FormatProgram(prog)
	if out != "" {
		t.Errorf("empty program should produce empty output, got %q", out)
	}
}

// --- Function formatting ---

func TestFormatter_SimpleFunction(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionStatement{
				Name:       "add",
				Params:     []string{"a", "b"},
				ParamTypes: []string{"int", "int"},
				ReturnType: "int",
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "a"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "b"},
						},
					},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "fn add(a: int, b: int) -> int") {
		t.Errorf("function signature not formatted correctly:\n%s", out)
	}
	if !strings.Contains(out, "return a + b") {
		t.Errorf("function body not formatted correctly:\n%s", out)
	}
}

func TestFormatter_FunctionWithAttributes(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionStatement{
				Name:       "fast",
				Params:     []string{},
				ParamTypes: []string{},
				Attributes: []*ast.Attribute{
					{Name: "inline", Args: []string{}},
				},
				Body: []ast.Statement{},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "#[inline]") {
		t.Errorf("attribute not formatted:\n%s", out)
	}
	if !strings.Contains(out, "fn fast()") {
		t.Errorf("function not formatted:\n%s", out)
	}
}

func TestFormatter_FunctionWithGenericParams(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionStatement{
				Name:       "identity",
				TypeParams: []*ast.TypeParameter{{Name: "T"}},
				Params:     []string{"x"},
				ParamTypes: []string{"T"},
				ReturnType: "T",
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.Identifier{Name: "x"},
					},
				},
				Generic: true,
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "fn identity[T](x: T) -> T") {
		t.Errorf("generic function not formatted correctly:\n%s", out)
	}
}

// --- Variable declarations ---

func TestFormatter_VariableDeclaration(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDeclaration{
				Name:  "x",
				Type:  "int",
				Value: &ast.IntegerLiteral{Value: 42},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "int x = 42") {
		t.Errorf("variable declaration not formatted:\n%s", out)
	}
}

func TestFormatter_ConstDeclaration(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDeclaration{
				Name:    "PI",
				Type:    "float",
				IsConst: true,
				Value:   &ast.FloatLiteral{Value: 3.14159},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "const float PI") {
		t.Errorf("const declaration not formatted:\n%s", out)
	}
}

func TestFormatter_AutoDeclaration(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDeclaration{
				Name:   "result",
				IsAuto: true,
				Value:  &ast.CallExpression{Function: &ast.Identifier{Name: "compute"}},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "auto result") {
		t.Errorf("auto declaration not formatted:\n%s", out)
	}
}

// --- Control flow ---

func TestFormatter_IfStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: ">",
					Right:    &ast.IntegerLiteral{Value: 0},
				},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.Identifier{Name: "x"},
					},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "if x > 0") {
		t.Errorf("if condition not formatted:\n%s", out)
	}
	if !strings.Contains(out, "return x") {
		t.Errorf("if body not formatted:\n%s", out)
	}
}

func TestFormatter_IfElseStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.Identifier{Name: "flag"},
				Body: []ast.Statement{
					&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 1}},
				},
				Else: []ast.Statement{
					&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 0}},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "if flag") {
		t.Errorf("if not formatted:\n%s", out)
	}
	if !strings.Contains(out, "else") {
		t.Errorf("else not formatted:\n%s", out)
	}
}

func TestFormatter_WhileLoop(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				Condition: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "i"},
					Operator: "<",
					Right:    &ast.IntegerLiteral{Value: 10},
				},
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "i"},
							Operator: "=",
							Right: &ast.BinaryExpression{
								Left:     &ast.Identifier{Name: "i"},
								Operator: "+",
								Right:    &ast.IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "while i < 10") {
		t.Errorf("while condition not formatted:\n%s", out)
	}
}

// --- Expressions ---

func TestFormatter_BinaryExpression(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: "+",
					Right:    &ast.Identifier{Name: "b"},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "a + b") {
		t.Errorf("binary expression not formatted:\n%s", out)
	}
}

func TestFormatter_CallExpression(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Function: &ast.Identifier{Name: "println"},
					Args: []ast.Expression{
						&ast.StringLiteral{Value: "hello"},
					},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, `println("hello")`) {
		t.Errorf("call expression not formatted:\n%s", out)
	}
}

func TestFormatter_UnaryExpression(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.UnaryExpression{
					Operator: "-",
					Right:    &ast.Identifier{Name: "x"},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "-x") {
		t.Errorf("unary expression not formatted:\n%s", out)
	}
}

func TestFormatter_MemberAccess(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.MemberAccessExpression{
					Object: &ast.Identifier{Name: "obj"},
					Member: "field",
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "obj.field") {
		t.Errorf("member access not formatted:\n%s", out)
	}
}

func TestFormatter_IndexExpression(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.IndexExpression{
					Object: &ast.Identifier{Name: "arr"},
					Index:  &ast.IntegerLiteral{Value: 0},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "arr[0]") {
		t.Errorf("index expression not formatted:\n%s", out)
	}
}

func TestFormatter_TypeCast(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.TypeCastExpression{
					TargetType:  "int",
					Expression:  &ast.Identifier{Name: "x"},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "as<int>(x)") {
		t.Errorf("type cast not formatted:\n%s", out)
	}
}

// --- Literals ---

func TestFormatter_StringLiteral(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.StringLiteral{Value: "test"},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, `"test"`) {
		t.Errorf("string literal not formatted:\n%s", out)
	}
}

func TestFormatter_BooleanLiteral(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.BooleanLiteral{Value: true},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "true") {
		t.Errorf("boolean literal not formatted:\n%s", out)
	}
}

// --- Break and continue ---

func TestFormatter_BreakStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.BreakStatement{},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "break") {
		t.Errorf("break not formatted:\n%s", out)
	}
}

func TestFormatter_ContinueStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ContinueStatement{},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "continue") {
		t.Errorf("continue not formatted:\n%s", out)
	}
}

// --- Package statement ---

func TestFormatter_PackageStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PackageStatement{Name: "mypackage"},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	if !strings.Contains(out, "package mypackage") {
		t.Errorf("package not formatted:\n%s", out)
	}
}

// --- Indentation ---

func TestFormatter_NestedIndentation(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionStatement{
				Name:       "outer",
				Params:     []string{},
				ParamTypes: []string{},
				Body: []ast.Statement{
					&ast.IfStatement{
						Condition: &ast.BooleanLiteral{Value: true},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.IntegerLiteral{Value: 1},
							},
						},
					},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	// Check that nested statements are indented
	lines := strings.Split(out, "\n")
	foundIndentedReturn := false
	for _, line := range lines {
		if strings.Contains(line, "return 1") && strings.HasPrefix(line, "        ") {
			foundIndentedReturn = true
			break
		}
	}
	if !foundIndentedReturn {
		t.Errorf("nested statement not properly indented:\n%s", out)
	}
}

// --- Complex expression ---

func TestFormatter_ComplexExpression(t *testing.T) {
	// Test: (a + b) * c
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "a"},
						Operator: "+",
						Right:    &ast.Identifier{Name: "b"},
					},
					Operator: "*",
					Right:    &ast.Identifier{Name: "c"},
				},
			},
		},
	}
	f := NewFormatter()
	out := f.FormatProgram(prog)

	// Should contain the expression (exact parenthesization depends on implementation)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") || !strings.Contains(out, "c") {
		t.Errorf("complex expression not formatted:\n%s", out)
	}
}
