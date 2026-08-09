package sema

import (
	"testing"

	"kaula-compiler/internal/ast"
	"kaula-compiler/internal/errors"
	"kaula-compiler/internal/lexer"
	"kaula-compiler/internal/parser"
)

// Helper: parse source into AST
func parseSource(t *testing.T, src string) *ast.Program {
	t.Helper()
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	program := p.Parse()
	return program
}

func TestNewSemanticAnalyzer(t *testing.T) {
	sa := NewSemanticAnalyzer()
	if sa == nil {
		t.Fatal("NewSemanticAnalyzer() returned nil")
	}
	if sa.ErrorCollector() == nil {
		t.Error("ErrorCollector() returned nil")
	}
	if sa.HasErrors() {
		t.Error("new analyzer should not have errors")
	}
}

func TestAnalyzeEmptyProgram(t *testing.T) {
	sa := NewSemanticAnalyzer()
	program := &ast.Program{
		Statements: make([]ast.Statement, 0),
		Source:     "",
	}
	// Should not panic
	sa.Analyze(program)
}

func TestAnalyzeSimpleFunction(t *testing.T) {
	src := `fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	if sa.HasErrors() {
		t.Errorf("unexpected errors: %v", sa.ErrorCollector().Errors())
	}
}

func TestAnalyzeUndefinedVariable(t *testing.T) {
	src := `fn main() void {
    x = 42
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	if !sa.HasErrors() {
		t.Error("expected errors for undefined variable 'x'")
	}
}

func TestAnalyzeDuplicateFunction(t *testing.T) {
	// The parser should catch duplicates, but we test the full pipeline
	src := `fn foo() void {
}
fn foo() void {
}
`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	program := p.Parse()
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// At minimum, should not panic
}

func TestAnalyzeStructFieldAccess(t *testing.T) {
	// This requires a struct type and field access
	src := `struct Point {
    int x
    int y
}

fn main() void {
    Point p
    p.x = 10
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic - may or may not have errors depending on implementation
	_ = sa.HasErrors()
}

func TestAnalyzeFunctionCall(t *testing.T) {
	src := `fn helper() void {
}

fn main() void {
    helper()
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeImportStatement(t *testing.T) {
	// import statements are typically resolved with stdlib config
	src := `import std.io

fn main() void {
    println("hello")
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeIfStatement(t *testing.T) {
	src := `fn main() void {
    if true {
        println("yes")
    } else {
        println("no")
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeWhileLoop(t *testing.T) {
	src := `fn main() void {
    while true {
        println("looping")
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeForInLoop(t *testing.T) {
	src := `fn main() void {
    for i in range(10) {
        println(i)
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeReturnStatement(t *testing.T) {
	src := `fn main() void {
    return
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeBinaryExpression(t *testing.T) {
	src := `fn main() void {
    int x = 1 + 2
    bool y = x == 3
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeNestedScopes(t *testing.T) {
	src := `fn main() void {
    int x = 1
    if true {
        int y = 2
        x = y
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeClassDeclaration(t *testing.T) {
	src := `class Animal {
    constructor(string name) {
        self.name = name
    }
}

fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeEnumDeclaration(t *testing.T) {
	src := `enum Color {
    Red, Green, Blue
}

fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzePrefixStatement(t *testing.T) {
	src := `prefix log {
    fn execute(string msg) void {
        println(msg)
    }
}

fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeMatchStatement(t *testing.T) {
	src := `fn main() void {
    int x = 3
    match x {
        1 => println("one"),
        2 => println("two"),
        _ => println("many")
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeSpendStatement(t *testing.T) {
	src := `fn main() void {
    int x = 3
    spend x {
        call 0 {
            println("zero")
        }
        call 1 {
            println("one")
        }
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
	// Should not panic; may or may not have errors
}

func TestAnalyzeWithStdlibConfig(t *testing.T) {
	sa := NewSemanticAnalyzer()
	// SetStdlibConfig with nil should not panic
	sa.SetStdlibConfig(nil)
	if sa.GetStdlibConfig() != nil {
		t.Error("GetStdlibConfig() should return nil")
	}
}

func TestAnalyzeSetSOREnabled(t *testing.T) {
	sa := NewSemanticAnalyzer()
	sa.SetSOREnabled(true)

	src := `fn main() void {
    yield x
    release x
}
`
	program := parseSource(t, src)
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeWithErrorCollector(t *testing.T) {
	ec := errors.NewErrorCollector()
	sa := NewSemanticAnalyzerWithConfig("kaula-compiler/stdlib.json", ec)
	if sa == nil {
		t.Fatal("NewSemanticAnalyzerWithConfig returned nil")
	}
	if sa.ErrorCollector() != ec {
		t.Error("ErrorCollector() should return the same instance")
	}
}

func TestAnalyzeLocalImportFuncs(t *testing.T) {
	sa := NewSemanticAnalyzer()
	funcs := map[string]bool{"helper": true}
	sa.SetLocalImportFuncs(funcs)
	sa.SetLocalModuleFuncs(map[string]bool{"helper": true, "internal_helper": true})
	sa.SetLocalPubVars(map[string]bool{"config": true})

	src := `fn main() void {
}
`
	program := parseSource(t, src)
	sa.Analyze(program)
	// Should not panic
}

func TestAnalyzeBinaryExpressionIntOps(t *testing.T) {
	src := `fn main() void {
    int a = 10
    int b = 20
    int c = a + b
    int d = a - b
    int e = a * b
    int f = a / b
    int g = a % b
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeComparisonOperators(t *testing.T) {
	src := `fn main() void {
    bool a = 1 == 2
    bool b = 1 != 2
    bool c = 1 < 2
    bool d = 1 > 2
    bool e = 1 <= 2
    bool f = 1 >= 2
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeLogicalOperators(t *testing.T) {
	src := `fn main() void {
    bool a = true && false
    bool b = true || false
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeVariableDeclarationWithInit(t *testing.T) {
	src := `fn main() void {
    int x = 42
    string s = "hello"
    bool b = true
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeArrayAccess(t *testing.T) {
	src := `fn main() void {
    int[10] arr
    int x = arr[0]
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeTypeCast(t *testing.T) {
	src := `fn main() void {
    float x = 3.14
    int y = x as int
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeExternFunction(t *testing.T) {
	src := `extern fn puts(string s) int

fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeInterfaceDeclaration(t *testing.T) {
	src := `interface Drawable {
    fn draw() void
}

fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeStructDeclaration(t *testing.T) {
	src := `struct Point {
    int x
    int y
}

fn main() void {
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeNestedIfElse(t *testing.T) {
	src := `fn main() void {
    int x = 5
    if x > 0 {
        if x > 10 {
            println(">10")
        } else {
            println("<=10")
        }
    }
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeFunctionWithParams(t *testing.T) {
	src := `fn add(int a, int b) int {
    return a + b
}

fn main() void {
    int result = add(1, 2)
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeYieldReleaseExtract(t *testing.T) {
	src := `fn main() void {
    int x = 42
    yield x
    release x
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeExtractStatement(t *testing.T) {
	src := `fn main() void {
    int x = 42
    extract x
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}

func TestAnalyzeNonLocalStatement(t *testing.T) {
	src := `fn main() void {
    nonlocal x = 42
}
`
	program := parseSource(t, src)
	sa := NewSemanticAnalyzer()
	sa.Analyze(program)
}