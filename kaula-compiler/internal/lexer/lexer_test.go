package lexer

import (
	"testing"
)

func TestLexer_Keywords(t *testing.T) {
	input := `spend call fn if else while for in return import export pub break continue class struct enum match yield release`
	l := NewLexer(input)

	expected := []struct {
		typ TokenType
		val string
	}{
		{TOKEN_SPEND, "spend"},
		{TOKEN_CALL, "call"},
		{TOKEN_FUNC, "fn"},
		{TOKEN_IF, "if"},
		{TOKEN_ELSE, "else"},
		{TOKEN_WHILE, "while"},
		{TOKEN_FOR, "for"},
		{TOKEN_IN, "in"},
		{TOKEN_RETURN, "return"},
		{TOKEN_IMPORT, "import"},
		{TOKEN_EXPORT, "export"},
		{TOKEN_PUB, "pub"},
		{TOKEN_BREAK, "break"},
		{TOKEN_CONTINUE, "continue"},
		{TOKEN_CLASS, "class"},
		{TOKEN_STRUCT, "struct"},
		{TOKEN_ENUM, "enum"},
		{TOKEN_MATCH, "match"},
		{TOKEN_YIELD, "yield"},
		{TOKEN_RELEASE, "release"},
	}

	for i, exp := range expected {
		tok := l.Next()
		if tok.Type != exp.typ {
			t.Errorf("token[%d] type = %v, want %v (value=%q)", i, TokenTypeToString(tok.Type), TokenTypeToString(exp.typ), tok.Value)
		}
		if tok.Value != exp.val {
			t.Errorf("token[%d] value = %q, want %q", i, tok.Value, exp.val)
		}
	}
}

func TestLexer_TypeKeywords(t *testing.T) {
	input := `int float double bool char string void auto const static extern`
	l := NewLexer(input)

	expected := []TokenType{
		TOKEN_TYPE_INT, TOKEN_TYPE_FLOAT, TOKEN_TYPE_DOUBLE,
		TOKEN_TYPE_BOOL, TOKEN_TYPE_CHAR, TOKEN_TYPE_STRING,
		TOKEN_TYPE_VOID, TOKEN_AUTO, TOKEN_CONST, TOKEN_STATIC, TOKEN_EXTERN,
	}

	for i, exp := range expected {
		tok := l.Next()
		if tok.Type != exp {
			t.Errorf("token[%d] type = %v, want %v", i, TokenTypeToString(tok.Type), TokenTypeToString(exp))
		}
	}
}

func TestLexer_NumberLiterals_Decimal(t *testing.T) {
	l := NewLexer("42")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "42" {
		t.Errorf("decimal: got %s %q, want INT \"42\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_NumberLiterals_Hex(t *testing.T) {
	l := NewLexer("0xFF")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0xFF" {
		t.Errorf("hex: got %s %q, want INT \"0xFF\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_NumberLiterals_Octal(t *testing.T) {
	l := NewLexer("0o77")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0o77" {
		t.Errorf("octal: got %s %q, want INT \"0o77\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_NumberLiterals_Binary(t *testing.T) {
	l := NewLexer("0b1010")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0b1010" {
		t.Errorf("binary: got %s %q, want INT \"0b1010\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_NumberLiterals_Float(t *testing.T) {
	l := NewLexer("3.14")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_FLOAT || tok.Value != "3.14" {
		t.Errorf("float: got %s %q, want FLOAT \"3.14\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_StringLiteral(t *testing.T) {
	l := NewLexer(`"hello world"`)
	tok := l.Next()
	if tok.Type != TOKEN_STRING || tok.Value != "hello world" {
		t.Errorf("string: got %s %q, want STRING \"hello world\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_StringLiteral_EscapeQuote(t *testing.T) {
	// String containing an escaped quote: "he\"llo"
	l := NewLexer(`"he\"llo"`)
	tok := l.Next()
	if tok.Type != TOKEN_STRING {
		t.Errorf("expected STRING token, got %s", TokenTypeToString(tok.Type))
	}
	// Value should contain the raw content between quotes (with backslash)
	if tok.Value != `he\"llo` {
		t.Errorf("string with escape: got %q, want %q", tok.Value, `he\"llo`)
	}
}

func TestLexer_UnterminatedString(t *testing.T) {
	l := NewLexer(`"unterminated`)
	tok := l.Next()
	if tok.Type != TOKEN_STRING {
		t.Errorf("expected STRING token for unterminated string, got %s", TokenTypeToString(tok.Type))
	}
	if !l.HasErrors() {
		t.Error("expected lexer error for unterminated string")
	}
}

func TestLexer_CharLiteral(t *testing.T) {
	l := NewLexer("'a'")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_CHAR || tok.Value != "a" {
		t.Errorf("char: got %s %q, want CHAR \"a\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_CharLiteral_Escape(t *testing.T) {
	l := NewLexer(`'\n'`)
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_CHAR || tok.Value != "\\n" {
		t.Errorf("char escape: got %s %q, want CHAR \"\\n\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_Operators(t *testing.T) {
	input := `+ - * / % = == != < > <= >= << >> && || & | ^ ~ . ?`
	l := NewLexer(input)

	expected := []TokenType{
		TOKEN_PLUS, TOKEN_MINUS, TOKEN_MULTIPLY, TOKEN_DIVIDE, TOKEN_MOD,
		TOKEN_ASSIGN, TOKEN_EQ, TOKEN_NE, TOKEN_LT, TOKEN_GT,
		TOKEN_LE, TOKEN_GE, TOKEN_LSHIFT, TOKEN_RSHIFT,
		TOKEN_AND, TOKEN_OR, TOKEN_AMPERSAND, TOKEN_PIPE,
		TOKEN_XOR, TOKEN_TILDE, TOKEN_DOT, TOKEN_QUESTION,
	}

	for i, exp := range expected {
		tok := l.Next()
		if tok.Type != exp {
			t.Errorf("operator[%d] = %v, want %v", i, TokenTypeToString(tok.Type), TokenTypeToString(exp))
		}
	}
}

func TestLexer_Arrow(t *testing.T) {
	l := NewLexer("=>")
	tok := l.Next()
	if tok.Type != TOKEN_ARROW || tok.Value != "=>" {
		t.Errorf("arrow: got %s %q, want ARROW \"=>\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_Delimiters(t *testing.T) {
	input := `( ) { } [ ] ; , : :: .`
	l := NewLexer(input)

	expected := []TokenType{
		TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_LBRACE, TOKEN_RBRACE,
		TOKEN_LBRACKET, TOKEN_RBRACKET, TOKEN_SEMICOLON, TOKEN_COMMA,
		TOKEN_COLON, TOKEN_DOUBLE_COLON, TOKEN_DOT,
	}

	for i, exp := range expected {
		tok := l.Next()
		if tok.Type != exp {
			t.Errorf("delimiter[%d] = %v, want %v", i, TokenTypeToString(tok.Type), TokenTypeToString(exp))
		}
	}
}

func TestLexer_AtAndDollar(t *testing.T) {
	l := NewLexer("@ $")
	tok := l.Next()
	if tok.Type != TOKEN_AT {
		t.Errorf("@: got %v, want AT", TokenTypeToString(tok.Type))
	}
	tok = l.Next()
	if tok.Type != TOKEN_PREFIX_REF {
		t.Errorf("$: got %v, want PREFIX_REF", TokenTypeToString(tok.Type))
	}
}

func TestLexer_BooleanAndNull(t *testing.T) {
	l := NewLexer("true false null")

	tok := l.Next()
	if tok.Type != TOKEN_TRUE {
		t.Errorf("true: got %v", TokenTypeToString(tok.Type))
	}
	tok = l.Next()
	if tok.Type != TOKEN_FALSE {
		t.Errorf("false: got %v", TokenTypeToString(tok.Type))
	}
	tok = l.Next()
	if tok.Type != TOKEN_NULL {
		t.Errorf("null: got %v", TokenTypeToString(tok.Type))
	}
}

func TestLexer_Comments(t *testing.T) {
	input := `x // this is a comment
y`
	l := NewLexer(input)
	tok := l.Next()
	if tok.Type != TOKEN_IDENT || tok.Value != "x" {
		t.Errorf("before comment: got %v %q", TokenTypeToString(tok.Type), tok.Value)
	}
	tok = l.Next()
	if tok.Type != TOKEN_IDENT || tok.Value != "y" {
		t.Errorf("after comment: got %v %q", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_Attribute(t *testing.T) {
	l := NewLexer("#[inline]")
	tok := l.Next()
	if tok.Type != TOKEN_ATTRIBUTE {
		t.Errorf("attribute: got %v, want ATTRIBUTE", TokenTypeToString(tok.Type))
	}
	if tok.Value != "#[inline]" {
		t.Errorf("attribute value: got %q, want \"#[inline]\"", tok.Value)
	}
}

func TestLexer_LineTracking(t *testing.T) {
	input := "a\nb\nc"
	l := NewLexer(input)

	tok := l.Next()
	if tok.Line != 1 {
		t.Errorf("token 'a' line = %d, want 1", tok.Line)
	}
	tok = l.Next()
	if tok.Line != 2 {
		t.Errorf("token 'b' line = %d, want 2", tok.Line)
	}
	tok = l.Next()
	if tok.Line != 3 {
		t.Errorf("token 'c' line = %d, want 3", tok.Line)
	}
}

func TestLexer_SaveRestoreState(t *testing.T) {
	l := NewLexer("a b c")
	tok1 := l.Next() // a
	if tok1.Value != "a" {
		t.Fatalf("first token = %q, want 'a'", tok1.Value)
	}

	state := l.SaveState()
	tok2 := l.Next() // b
	if tok2.Value != "b" {
		t.Fatalf("second token = %q, want 'b'", tok2.Value)
	}

	l.RestoreState(state)
	tok2Again := l.Next() // should be b again
	if tok2Again.Value != "b" {
		t.Errorf("after restore: got %q, want 'b'", tok2Again.Value)
	}
}

func TestLexer_EOF(t *testing.T) {
	l := NewLexer("")
	tok := l.Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("empty input: got %v, want EOF", TokenTypeToString(tok.Type))
	}
}

func TestLexer_Identifier(t *testing.T) {
	l := NewLexer("myVar_123")
	tok := l.Next()
	if tok.Type != TOKEN_IDENT || tok.Value != "myVar_123" {
		t.Errorf("ident: got %s %q, want IDENT \"myVar_123\"", TokenTypeToString(tok.Type), tok.Value)
	}
}

func TestLexer_ComptimeKeywords(t *testing.T) {
	input := `comptime sizeof alignof offsetof type_name field_count field_name field_type type_kind`
	l := NewLexer(input)

	expected := []TokenType{
		TOKEN_COMPTIME, TOKEN_SIZEOF, TOKEN_ALIGNOF, TOKEN_OFFSETOF,
		TOKEN_TYPE_NAME, TOKEN_FIELD_COUNT, TOKEN_FIELD_NAME,
		TOKEN_FIELD_TYPE, TOKEN_TYPE_KIND,
	}

	for i, exp := range expected {
		tok := l.Next()
		if tok.Type != exp {
			t.Errorf("comptime keyword[%d] = %v, want %v", i, TokenTypeToString(tok.Type), TokenTypeToString(exp))
		}
	}
}

func TestLexer_UnexpectedCharacter(t *testing.T) {
	l := NewLexer("`")
	l.Next()
	if !l.HasErrors() {
		t.Error("expected error for unexpected character '`'")
	}
}

func TestLexer_ScanUntilRbrace(t *testing.T) {
	l := NewLexer("mov rax, rbx}")
	// Skip to the content
	content := l.ScanUntilRbrace()
	if content != "mov rax, rbx" {
		t.Errorf("ScanUntilRbrace: got %q, want \"mov rax, rbx\"", content)
	}
}

func TestTokenTypeToString(t *testing.T) {
	if s := TokenTypeToString(TOKEN_FUNC); s != "FUNC" {
		t.Errorf("TokenTypeToString(FUNC) = %q, want \"FUNC\"", s)
	}
	if s := TokenTypeToString(TOKEN_EOF); s != "EOF" {
		t.Errorf("TokenTypeToString(EOF) = %q, want \"EOF\"", s)
	}
}

func TestToken_String(t *testing.T) {
	tok := Token{Type: TOKEN_IDENT, Value: "foo", Line: 1, Column: 5}
	s := tok.String()
	if s == "" {
		t.Error("Token.String() should not be empty")
	}
}
