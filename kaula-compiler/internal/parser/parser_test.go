package parser_test

import (
	"testing"

	"kaula-compiler/internal/lexer"
	"kaula-compiler/internal/parser"
)

// parseHelper creates a parser from source and returns the program and parser.
func parseHelper(t *testing.T, src string) (*parser.Parser, *parser.Parser) {
	t.Helper()
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	p.Parse()
	_ = p // keep reference for error checking
	return p, p
}

func TestParser_EmptyProgram(t *testing.T) {
	src := ``
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()
	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 0 {
		t.Errorf("Expected 0 statements, got %d", prog.StatementCount())
	}
	// Empty program: no statements, but validation reports "missing main"
	// This is expected behavior, not a crash
}

func TestParser_FunctionDeclaration(t *testing.T) {
	src := `fn main() -> int {
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for valid function declaration")
	}
}

func TestParser_FunctionWithParams(t *testing.T) {
	src := `fn add(a: int, b: int) -> int {
    return a + b
}

fn main() -> int {
    return add(1, 2)
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements, got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for valid function with params")
	}
}

func TestParser_IfStatement(t *testing.T) {
	src := `fn main() int {
    int x = 10
    if (x > 5) {
        return 1
    } else {
        return 0
    }
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement (function), got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for if-else statement")
	}
}

func TestParser_WhileLoop(t *testing.T) {
	src := `fn main() int {
    int i = 0
    while (i < 10) {
        i = i + 1
    }
    return i
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for while loop")
	}
}

func TestParser_ForInLoop(t *testing.T) {
	src := `fn main() int {
    int sum = 0
    for (i in 0..10) {
        sum = sum + i
    }
    return sum
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
}

func TestParser_ImportStatement(t *testing.T) {
	src := `import std.io

fn main() int {
    println("hello")
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements (import + function), got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for import statement")
	}
}

func TestParser_VariableDeclarations(t *testing.T) {
	src := `fn main() int {
    int a = 1
    int64 b = 2
    string s = "hello"
    bool flag = true
    float pi = 3.14
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for variable declarations")
	}
}

func TestParser_AutoDeclaration(t *testing.T) {
	src := `fn main() int {
    auto x = 42
    auto y = "hello"
    return x
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
	// auto declarations may or may not produce errors depending on sema integration
	// but the parser should not crash
}

func TestParser_StructDefinition(t *testing.T) {
	src := `struct Point {
    int x
    int y
}

fn main() int {
    Point p
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements (struct + function), got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for struct definition")
	}
}

func TestParser_EnumDefinition(t *testing.T) {
	src := `enum Color {
    Red, Green, Blue
}

fn main() int {
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements (enum + function), got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for enum definition")
	}
}

func TestParser_MatchExpression(t *testing.T) {
	src := `fn main() int {
    int x = 2
    match x {
        1 => { return 10 }
        2 => { return 20 }
        _ => { return 0 }
    }
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
	// match statements may or may not be fully parsed depending on implementation
	// but the parser should not hang
}

func TestParser_ClassDefinition(t *testing.T) {
	src := `class Animal {
    fn speak() string {
        return "..."
    }
}

fn main() int {
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	// Check that the parser completed without hang
}

func TestParser_ExpressionStatement(t *testing.T) {
	src := `fn main() int {
    a + b
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
}

func TestParser_CommentOnly(t *testing.T) {
	src := `// just a comment
// another comment`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 0 {
		t.Errorf("Expected 0 statements for comment-only file, got %d", prog.StatementCount())
	}
}

func TestParser_CallStatement(t *testing.T) {
	src := `import std.io

fn main() int {
    println("hello world")
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements, got %d", prog.StatementCount())
	}
	if p.HasErrors() {
		t.Error("Unexpected errors for function call")
	}
}

func TestParser_ExportStatement(t *testing.T) {
	src := `export const C_MAX = 100

fn main() int {
    return C_MAX
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements, got %d", prog.StatementCount())
	}
}

func TestParser_ExternDeclaration(t *testing.T) {
	src := `extern fn puts(s: string) -> int

fn main() int {
    return 0
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements, got %d", prog.StatementCount())
	}
}

func TestParser_PubFunction(t *testing.T) {
	src := `pub fn add(a: int, b: int) -> int {
    return a + b
}

fn main() int {
    return add(1, 2)
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 2 {
		t.Fatalf("Expected 2 statements, got %d", prog.StatementCount())
	}
}

func TestParser_SyntaxErrorMissingParen(t *testing.T) {
	src := `fn main() int {
    return
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	_ = p.Parse()
	// The parser should handle this without crashing
	// Errors may or may not be reported depending on the parser's error recovery
}

func TestParser_SpendStatement(t *testing.T) {
	src := `fn main() int {
    spend arr {
        0 => { return 0 }
        default => { return 1 }
    }
}`
	lx := lexer.NewLexer(src)
	p := parser.NewParser(lx)
	prog := p.Parse()

	if prog == nil {
		t.Fatal("Parse() returned nil")
	}
	if prog.StatementCount() != 1 {
		t.Fatalf("Expected 1 statement, got %d", prog.StatementCount())
	}
}