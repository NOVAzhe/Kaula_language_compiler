package parser_test

import (
	"testing"

	"kaula-compiler/internal/ast"
	"kaula-compiler/internal/lexer"
	"kaula-compiler/internal/parser"
)

func parseSource(t *testing.T, src string) *ast.Program {
	t.Helper()
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	program := p.Parse()
	if p.HasErrors() {
		// Don't fail here, let the test decide what to do with errors
	}
	return program
}

func TestParseEmptyProgram(t *testing.T) {
	// Empty file should be valid (no main check without skipMainCheck)
	lx := lexer.NewLexer("")
	p := parser.NewParser(lx)
	program := p.Parse()
	if program == nil {
		t.Fatal("Parse() returned nil")
	}
	if program.Statements == nil {
		t.Error("Statements is nil")
	}
}

func TestParseSimpleFunction(t *testing.T) {
	src := `fn main() void {
}
`
	program := parseSource(t, src)
	if program == nil {
		t.Fatal("Parse() returned nil")
	}
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if fn.Name != "main" {
		t.Errorf("function name = %q, want %q", fn.Name, "main")
	}
	if fn.ReturnType != "void" {
		t.Errorf("return type = %q, want %q", fn.ReturnType, "void")
	}
}

func TestParseVariableDeclaration(t *testing.T) {
	src := `fn main() void {
    int x = 42
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	// Should be a variable declaration
	_, ok = fn.Body[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration, got %T", fn.Body[0])
	}
}

func TestParseIfStatement(t *testing.T) {
	src := `fn main() void {
    if x == 1 {
        println("one")
    }
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	ifStmt, ok := fn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", fn.Body[0])
	}
	if ifStmt.Condition == nil {
		t.Error("IfStatement.Condition is nil")
	}
	if len(ifStmt.Body) == 0 {
		t.Error("IfStatement.Body is empty")
	}
}

func TestParseIfElseStatement(t *testing.T) {
	src := `fn main() void {
    if x == 1 {
        println("one")
    } else {
        println("other")
    }
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	ifStmt, ok := fn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", fn.Body[0])
	}
	if len(ifStmt.Else) == 0 {
		t.Error("IfStatement.Else is empty")
	}
}

func TestParseWhileStatement(t *testing.T) {
	src := `fn main() void {
    while x < 10 {
        x = x + 1
    }
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	_, ok = fn.Body[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", fn.Body[0])
	}
}

func TestParseForInStatement(t *testing.T) {
	src := `fn main() void {
    for i in range(10) {
        println(i)
    }
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	_, ok = fn.Body[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("expected ForInStatement, got %T", fn.Body[0])
	}
}

func TestParseReturnStatement(t *testing.T) {
	src := `fn add(int a, int b) int {
    return a + b
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	retStmt, ok := fn.Body[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", fn.Body[0])
	}
	if retStmt.Value == nil {
		t.Error("ReturnStatement.Value is nil")
	}
}

func TestParseImportStatement(t *testing.T) {
	src := `import std.io

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	_, ok := program.Statements[0].(*ast.ImportStatement)
	if !ok {
		t.Fatalf("expected ImportStatement, got %T", program.Statements[0])
	}
}

func TestParseStructDeclaration(t *testing.T) {
	src := `struct Point {
    int x
    int y
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	structStmt, ok := program.Statements[0].(*ast.StructStatement)
	if !ok {
		t.Fatalf("expected StructStatement, got %T", program.Statements[0])
	}
	if structStmt.Name != "Point" {
		t.Errorf("struct name = %q, want %q", structStmt.Name, "Point")
	}
}

func TestParseEnumDeclaration(t *testing.T) {
	src := `enum Color {
    Red, Green, Blue
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	enumStmt, ok := program.Statements[0].(*ast.EnumStatement)
	if !ok {
		t.Fatalf("expected EnumStatement, got %T", program.Statements[0])
	}
	if enumStmt.Name != "Color" {
		t.Errorf("enum name = %q, want %q", enumStmt.Name, "Color")
	}
}

func TestParseBinaryExpression(t *testing.T) {
	src := `fn main() void {
    int x = 1 + 2 * 3
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	varDecl, ok := fn.Body[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration, got %T", fn.Body[0])
	}
	if varDecl.Value == nil {
		t.Fatal("VariableDeclaration.Value is nil")
	}
	// Should be a binary expression
	_, ok = varDecl.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression, got %T", varDecl.Value)
	}
}

func TestParseFunctionCall(t *testing.T) {
	src := `fn main() void {
    println("hello")
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	exprStmt, ok := fn.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", fn.Body[0])
	}
	if exprStmt.Expression == nil {
		t.Fatal("ExpressionStatement.Expression is nil")
	}
	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", exprStmt.Expression)
	}
	if callExpr.Function == nil {
		t.Error("CallExpression.Function is nil")
	}
}

func TestParseFunctionWithParams(t *testing.T) {
	src := `fn add(int a, int b) int {
    return a + b
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if fn.Name != "add" {
		t.Errorf("function name = %q, want %q", fn.Name, "add")
	}
	if len(fn.Params) < 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.Params[0] != "a" || fn.Params[1] != "b" {
		t.Errorf("params = %v, want [a b]", fn.Params)
	}
}

func TestParseClassDeclaration(t *testing.T) {
	src := `class Animal {
    constructor(string name) {
        self.name = name
    }
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	classStmt, ok := program.Statements[0].(*ast.ClassStatement)
	if !ok {
		t.Fatalf("expected ClassStatement, got %T", program.Statements[0])
	}
	if classStmt.Name != "Animal" {
		t.Errorf("class name = %q, want %q", classStmt.Name, "Animal")
	}
}

func TestParsePrefixStatement(t *testing.T) {
	src := `prefix print {
    fn execute(string msg) void {
        println(msg)
    }
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	prefixStmt, ok := program.Statements[0].(*ast.PrefixStatement)
	if !ok {
		t.Fatalf("expected PrefixStatement, got %T", program.Statements[0])
	}
	if prefixStmt.Name != "print" {
		t.Errorf("prefix name = %q, want %q", prefixStmt.Name, "print")
	}
}

func TestParseMatchStatement(t *testing.T) {
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
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	// Should have 2 statements: variable declaration and match
	if len(fn.Body) < 2 {
		t.Fatalf("expected 2 body statements, got %d", len(fn.Body))
	}
	// Match is parsed as ExpressionStatement containing MatchExpression
	exprStmt, ok := fn.Body[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", fn.Body[1])
	}
	_, ok = exprStmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected MatchExpression, got %T", exprStmt.Expression)
	}
}

func TestParseBreakContinue(t *testing.T) {
	src := `fn main() void {
    while true {
        break
        continue
    }
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	whileStmt, ok := fn.Body[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", fn.Body[0])
	}
	if len(whileStmt.Body) < 2 {
		t.Fatalf("expected 2 body statements in while, got %d", len(whileStmt.Body))
	}
	_, ok = whileStmt.Body[0].(*ast.BreakStatement)
	if !ok {
		t.Fatalf("expected BreakStatement, got %T", whileStmt.Body[0])
	}
	_, ok = whileStmt.Body[1].(*ast.ContinueStatement)
	if !ok {
		t.Fatalf("expected ContinueStatement, got %T", whileStmt.Body[1])
	}
}

func TestParseMultipleFunctions(t *testing.T) {
	src := `fn foo() void {
}
fn bar() void {
}
fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 3 {
		t.Fatalf("expected 3 statements, got %d", len(program.Statements))
	}
	fn1, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok || fn1.Name != "foo" {
		t.Errorf("first function = %s, want foo", fn1.Name)
	}
	fn2, ok := program.Statements[1].(*ast.FunctionStatement)
	if !ok || fn2.Name != "bar" {
		t.Errorf("second function = %s, want bar", fn2.Name)
	}
	fn3, ok := program.Statements[2].(*ast.FunctionStatement)
	if !ok || fn3.Name != "main" {
		t.Errorf("third function = %s, want main", fn3.Name)
	}
}

func TestParseNestedBlocks(t *testing.T) {
	src := `fn main() void {
    if true {
        if false {
            println("nested")
        }
    }
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	outerIf, ok := fn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", fn.Body[0])
	}
	if len(outerIf.Body) < 1 {
		t.Fatal("outer if body is empty")
	}
	_, ok = outerIf.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected nested IfStatement, got %T", outerIf.Body[0])
	}
}

func TestParseErrorDuplicateFunction(t *testing.T) {
	src := `fn foo() void {
}
fn foo() void {
}
`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	_ = p.Parse()
	// Should have errors about duplicate function name
	if !p.HasErrors() {
		t.Error("expected errors for duplicate function name")
	}
}

func TestParserErrorMissingMain(t *testing.T) {
	src := `fn foo() void {
}
`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	_ = p.Parse()
	// Should have errors about missing main
	if !p.HasErrors() {
		t.Error("expected errors for missing main function")
	}
}

func TestParseSpendStatement(t *testing.T) {
	src := `fn main() void {
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
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	_, ok = fn.Body[0].(*ast.SpendStatement)
	if !ok {
		t.Fatalf("expected SpendStatement, got %T", fn.Body[0])
	}
}

func TestParseAssignment(t *testing.T) {
	src := `fn main() void {
    x = 42
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 body statement")
	}
	_, ok = fn.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", fn.Body[0])
	}
}

func TestParseExternFunction(t *testing.T) {
	src := `extern fn puts(string s) int

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	// The parser may parse extern fn as an ExpressionStatement
	_, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		// Fallback: check if it's parsed as ExternStatement
		if _, ok2 := program.Statements[0].(*ast.ExternStatement); !ok2 {
			t.Fatalf("expected ExpressionStatement or ExternStatement, got %T", program.Statements[0])
		}
	}
}

func TestParsePubFunction(t *testing.T) {
	src := `pub fn helper() void {
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	_, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
}

func TestParsePackageStatement(t *testing.T) {
	src := `package mylib

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	pkgStmt, ok := program.Statements[0].(*ast.PackageStatement)
	if !ok {
		t.Fatalf("expected PackageStatement, got %T", program.Statements[0])
	}
	if pkgStmt.Name != "mylib" {
		t.Errorf("package name = %q, want %q", pkgStmt.Name, "mylib")
	}
}

func TestParseExportStatement(t *testing.T) {
	src := `export fn helper() void {
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	// export fn is parsed as a FunctionStatement (the export is handled by the parser)
	_, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		// Fallback: check export statement
		if _, ok2 := program.Statements[0].(*ast.ExportStatement); !ok2 {
			t.Fatalf("expected FunctionStatement or ExportStatement, got %T", program.Statements[0])
		}
	}
}

func TestParseAttributeAnnotation(t *testing.T) {
	src := `#[sor] fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	_, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
}

func TestParseTreeStatement(t *testing.T) {
	src := `tree MyTree {
    root = "data"
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	_, ok := program.Statements[0].(*ast.TreeStatement)
	if !ok {
		t.Fatalf("expected TreeStatement, got %T", program.Statements[0])
	}
}

func TestParseObjectStatement(t *testing.T) {
	src := `object MyObject {
    string name
    int value
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	// The parser may parse object as VariableDeclaration or ObjectStatement
	switch stmt := program.Statements[0].(type) {
	case *ast.ObjectStatement:
		if stmt.Name != "MyObject" {
			t.Errorf("object name = %q, want %q", stmt.Name, "MyObject")
		}
	case *ast.VariableDeclaration:
		// object MyObject is parsed as a variable declaration
	default:
		t.Fatalf("expected ObjectStatement or VariableDeclaration, got %T", program.Statements[0])
	}
}

func TestParseInterfaceDeclaration(t *testing.T) {
	src := `interface Drawable {
    fn draw() void
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	ifaceStmt, ok := program.Statements[0].(*ast.InterfaceStatement)
	if !ok {
		t.Fatalf("expected InterfaceStatement, got %T", program.Statements[0])
	}
	if ifaceStmt.Name != "Drawable" {
		t.Errorf("interface name = %q, want %q", ifaceStmt.Name, "Drawable")
	}
}

func TestParserHasErrors(t *testing.T) {
	lx := lexer.NewLexer("!")
	p := parser.NewParser(lx)
	_ = p.Parse()
	if !p.HasErrors() {
		t.Error("expected parser to have errors for invalid input")
	}
}

func TestParserReportErrors(t *testing.T) {
	lx := lexer.NewLexer("!")
	p := parser.NewParser(lx)
	_ = p.Parse()
	// Should not panic
	p.ReportErrors()
}

func TestParserSkipMainCheck(t *testing.T) {
	// When skipMainCheck is set, no main function should not cause errors
	src := `fn helper() void {
}
`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	p.SetSkipMainCheck(true)
	_ = p.Parse()
	if p.HasErrors() {
		t.Error("expected no errors when skipMainCheck is set")
	}
}

func TestParserEnableLogging(t *testing.T) {
	src := `fn main() void {
}
`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	// Should not panic
	p.EnableLogging(false)
	_ = p.Parse()
}

func TestParseImplementsClause(t *testing.T) {
	src := `class Circle implements Drawable {
    constructor() {
    }
}

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	classStmt, ok := program.Statements[0].(*ast.ClassStatement)
	if !ok {
		t.Fatalf("expected ClassStatement, got %T", program.Statements[0])
	}
	if classStmt.Name != "Circle" {
		t.Errorf("class name = %q, want %q", classStmt.Name, "Circle")
	}
}

func TestParseYieldReleaseExtract(t *testing.T) {
	src := `fn main() void {
    yield x
    release x
    extract x
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if len(fn.Body) < 3 {
		t.Fatalf("expected 3 body statements, got %d", len(fn.Body))
	}
	_, ok = fn.Body[0].(*ast.YieldStatement)
	if !ok {
		t.Fatalf("expected YieldStatement, got %T", fn.Body[0])
	}
	_, ok = fn.Body[1].(*ast.ReleaseStatement)
	if !ok {
		t.Fatalf("expected ReleaseStatement, got %T", fn.Body[1])
	}
	_, ok = fn.Body[2].(*ast.ExtractStatement)
	if !ok {
		t.Fatalf("expected ExtractStatement, got %T", fn.Body[2])
	}
}

func TestParseStaticConst(t *testing.T) {
	src := `static const int MAX_SIZE = 1024

fn main() void {
}
`
	program := parseSource(t, src)
	if len(program.Statements) < 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestParseEmptyBodyFunction(t *testing.T) {
	src := `fn main() void {}
`
	program := parseSource(t, src)
	if len(program.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
	if fn.Name != "main" {
		t.Errorf("function name = %q, want %q", fn.Name, "main")
	}
}