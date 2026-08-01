package parser

import (
	"kaula-compiler/internal/ast"
	"kaula-compiler/internal/lexer"
	"testing"
)

// parseHelper creates a parser and parses the input, returning the program
func parseHelper(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := NewParser(l)
	p.EnableLogging(false)
	return p.Parse()
}

// --- Basic parsing ---

func TestParser_EmptyProgram(t *testing.T) {
	prog := parseHelper("")
	if prog == nil {
		t.Fatal("program should not be nil")
	}
	if len(prog.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Statements))
	}
}

func TestParser_SingleImport(t *testing.T) {
	prog := parseHelper("import std.string")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	imp, ok := prog.Statements[0].(*ast.ImportStatement)
	if !ok {
		t.Fatalf("expected ImportStatement, got %T", prog.Statements[0])
	}
	if imp.Module != "std.string" {
		t.Errorf("expected module 'std.string', got %q", imp.Module)
	}
}

func TestParser_MultipleImports(t *testing.T) {
	prog := parseHelper(`
import std.string
import std.io
import net.http
`)
	if len(prog.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(prog.Statements))
	}
	for i, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.ImportStatement); !ok {
			t.Errorf("statement %d: expected ImportStatement, got %T", i, stmt)
		}
	}
}

// --- Variable declarations ---

func TestParser_SimpleVariableDeclaration(t *testing.T) {
	prog := parseHelper("int x = 42")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	varDecl, ok := prog.Statements[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration, got %T", prog.Statements[0])
	}
	if varDecl.Name != "x" {
		t.Errorf("expected name 'x', got %q", varDecl.Name)
	}
	if varDecl.Type != "int" {
		t.Errorf("expected type 'int', got %q", varDecl.Type)
	}
	if varDecl.Value == nil {
		t.Error("expected value, got nil")
	}
}

func TestParser_ConstDeclaration(t *testing.T) {
	prog := parseHelper("const float PI = 3.14159")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	varDecl, ok := prog.Statements[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration, got %T", prog.Statements[0])
	}
	if !varDecl.IsConst {
		t.Error("expected IsConst to be true")
	}
	if varDecl.Type != "float" {
		t.Errorf("expected type 'float', got %q", varDecl.Type)
	}
}

// --- Function declarations ---

func TestParser_SimpleFunction(t *testing.T) {
	prog := parseHelper(`
fn add(int a, int b) int {
    return a + b
}
`)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", prog.Statements[0])
	}
	if fn.Name != "add" {
		t.Errorf("expected name 'add', got %q", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.ReturnType != "int" {
		t.Errorf("expected return type 'int', got %q", fn.ReturnType)
	}
	if len(fn.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(fn.Body))
	}
}

func TestParser_FunctionWithNoParams(t *testing.T) {
	prog := parseHelper(`
fn greet() {
    println("hello")
}
`)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", prog.Statements[0])
	}
	if fn.Name != "greet" {
		t.Errorf("expected name 'greet', got %q", fn.Name)
	}
	if len(fn.Params) != 0 {
		t.Errorf("expected 0 params, got %d", len(fn.Params))
	}
}

func TestParser_FunctionWithReturnType(t *testing.T) {
	prog := parseHelper(`
fn factorial(int n) int {
    return n
}
`)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", prog.Statements[0])
	}
	if fn.ReturnType != "int" {
		t.Errorf("expected return type 'int', got %q", fn.ReturnType)
	}
}

// --- Control flow ---

func TestParser_IfStatement(t *testing.T) {
	prog := parseHelper(`
fn test() {
    if (x > 0) {
        return x
    }
}
`)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn := prog.Statements[0].(*ast.FunctionStatement)
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(fn.Body))
	}
	ifStmt, ok := fn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", fn.Body[0])
	}
	if ifStmt.Condition == nil {
		t.Error("expected condition, got nil")
	}
	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 if-body statement, got %d", len(ifStmt.Body))
	}
}

func TestParser_IfElseStatement(t *testing.T) {
	prog := parseHelper(`
fn test() {
    if (flag) {
        return 1
    } else {
        return 0
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	ifStmt := fn.Body[0].(*ast.IfStatement)
	if !ifStmt.HasElse() {
		t.Error("expected else branch")
	}
	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

func TestParser_WhileLoop(t *testing.T) {
	prog := parseHelper(`
fn test() {
    while (i < 10) {
        i = i + 1
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	whileStmt, ok := fn.Body[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", fn.Body[0])
	}
	if whileStmt.Condition == nil {
		t.Error("expected condition, got nil")
	}
	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestParser_ForInLoop(t *testing.T) {
	prog := parseHelper(`
fn test() {
    for i in range(10) {
        println(i)
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	forInStmt, ok := fn.Body[0].(*ast.ForInStatement)
	if !ok {
		t.Fatalf("expected ForInStatement, got %T", fn.Body[0])
	}
	if forInStmt.Variable == nil {
		t.Error("expected variable, got nil")
	}
	if forInStmt.Iterable == nil {
		t.Error("expected iterable, got nil")
	}
}

// --- Expressions ---

func TestParser_CallExpression(t *testing.T) {
	prog := parseHelper(`
fn test() {
    println("hello")
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	exprStmt := fn.Body[0].(*ast.ExpressionStatement)
	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", exprStmt.Expression)
	}
	if callExpr.Function == nil {
		t.Error("expected function, got nil")
	}
	if len(callExpr.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(callExpr.Args))
	}
}

// --- Break and continue ---

func TestParser_BreakStatement(t *testing.T) {
	prog := parseHelper(`
fn test() {
    while (true) {
        break
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	whileStmt := fn.Body[0].(*ast.WhileStatement)
	breakStmt, ok := whileStmt.Body[0].(*ast.BreakStatement)
	if !ok {
		t.Fatalf("expected BreakStatement, got %T", whileStmt.Body[0])
	}
	if breakStmt == nil {
		t.Error("break statement is nil")
	}
}

func TestParser_ContinueStatement(t *testing.T) {
	prog := parseHelper(`
fn test() {
    while (true) {
        continue
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	whileStmt := fn.Body[0].(*ast.WhileStatement)
	continueStmt, ok := whileStmt.Body[0].(*ast.ContinueStatement)
	if !ok {
		t.Fatalf("expected ContinueStatement, got %T", whileStmt.Body[0])
	}
	if continueStmt == nil {
		t.Error("continue statement is nil")
	}
}

// --- Complex scenarios ---

func TestParser_NestedIfStatements(t *testing.T) {
	prog := parseHelper(`
fn test() {
    if (a) {
        if (b) {
            return 1
        }
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	outerIf := fn.Body[0].(*ast.IfStatement)
	innerIf, ok := outerIf.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected nested IfStatement, got %T", outerIf.Body[0])
	}
	if len(innerIf.Body) != 1 {
		t.Errorf("expected 1 inner if body statement, got %d", len(innerIf.Body))
	}
}

func TestParser_FunctionWithMultipleStatements(t *testing.T) {
	prog := parseHelper(`
fn test() {
    int x = 1
    int y = 2
    int z = x + y
    return z
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	if len(fn.Body) != 4 {
		t.Errorf("expected 4 body statements, got %d", len(fn.Body))
	}
}

// --- Package statement ---

func TestParser_PackageStatement(t *testing.T) {
	prog := parseHelper("package mypackage")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	pkgStmt, ok := prog.Statements[0].(*ast.PackageStatement)
	if !ok {
		t.Fatalf("expected PackageStatement, got %T", prog.Statements[0])
	}
	if pkgStmt.Name != "mypackage" {
		t.Errorf("expected package name 'mypackage', got %q", pkgStmt.Name)
	}
}

// --- Return statement ---

func TestParser_ReturnStatement(t *testing.T) {
	prog := parseHelper(`
fn test() {
    return 42
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	returnStmt, ok := fn.Body[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", fn.Body[0])
	}
	if returnStmt.Value == nil {
		t.Error("expected return value, got nil")
	}
}

func TestParser_ReturnWithoutValue(t *testing.T) {
	prog := parseHelper(`
fn test() {
    return
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	returnStmt, ok := fn.Body[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", fn.Body[0])
	}
	// Return without value should have nil Value
	if returnStmt.Value != nil {
		t.Error("expected nil return value")
	}
}

// --- Edge cases ---

func TestParser_EmptyFunctionBody(t *testing.T) {
	prog := parseHelper(`
fn empty() {
}
`)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", prog.Statements[0])
	}
	if len(fn.Body) != 0 {
		t.Errorf("expected 0 body statements, got %d", len(fn.Body))
	}
}

func TestParser_MultipleFunctions(t *testing.T) {
	prog := parseHelper(`
fn foo() {
    return 1
}
fn bar() {
    return 2
}
`)
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}
	for i, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.FunctionStatement); !ok {
			t.Errorf("statement %d: expected FunctionStatement, got %T", i, stmt)
		}
	}
}

func TestParser_IfElseIfChain(t *testing.T) {
	prog := parseHelper(`
fn test() {
    if (x > 0) {
        return 1
    } else if (x < 0) {
        return -1
    } else {
        return 0
    }
}
`)
	fn := prog.Statements[0].(*ast.FunctionStatement)
	ifStmt := fn.Body[0].(*ast.IfStatement)
	if !ifStmt.HasElse() {
		t.Error("expected else branch")
	}
	// The else branch should contain another IfStatement
	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifStmt.Else))
	}
	innerIf, ok := ifStmt.Else[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected nested IfStatement in else, got %T", ifStmt.Else[0])
	}
	if !innerIf.HasElse() {
		t.Error("expected inner if to have else branch")
	}
}
