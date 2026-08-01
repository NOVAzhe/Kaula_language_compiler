package lexer

import (
	"testing"
)

// collectTokens lexes the full input and returns all tokens up to (and including) EOF.
func collectTokens(input string) []Token {
	l := NewLexer(input)
	var tokens []Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

// --- Keywords ---

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"fn", TOKEN_FUNC},
		{"if", TOKEN_IF},
		{"else", TOKEN_ELSE},
		{"while", TOKEN_WHILE},
		{"for", TOKEN_FOR},
		{"in", TOKEN_IN},
		{"return", TOKEN_RETURN},
		{"import", TOKEN_IMPORT},
		{"vo", TOKEN_VO},
		{"spend", TOKEN_SPEND},
		{"call", TOKEN_CALL},
		{"task", TOKEN_TASK},
		{"prefix", TOKEN_PREFIX},
		{"tree", TOKEN_TREE},
		{"object", TOKEN_OBJECT},
		{"class", TOKEN_CLASS},
		{"struct", TOKEN_STRUCT},
		{"enum", TOKEN_ENUM},
		{"match", TOKEN_MATCH},
		{"yield", TOKEN_YIELD},
		{"release", TOKEN_RELEASE},
		{"extract", TOKEN_EXTRACT},
		{"extern", TOKEN_EXTERN},
		{"static", TOKEN_STATIC},
		{"const", TOKEN_CONST},
		{"pub", TOKEN_PUB},
		{"break", TOKEN_BREAK},
		{"continue", TOKEN_CONTINUE},
		{"as", TOKEN_AS},
		{"auto", TOKEN_AUTO},
	}
	for _, tt := range tests {
		tok := NewLexer(tt.input).Next()
		if tok.Type != tt.expected {
			t.Errorf("keyword %q: got type %d, want %d", tt.input, tok.Type, tt.expected)
		}
		if tok.Value != tt.input {
			t.Errorf("keyword %q: got value %q", tt.input, tok.Value)
		}
	}
}

func TestLexer_TypeKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"int", TOKEN_TYPE_INT},
		{"float", TOKEN_TYPE_FLOAT},
		{"double", TOKEN_TYPE_DOUBLE},
		{"bool", TOKEN_TYPE_BOOL},
		{"char", TOKEN_TYPE_CHAR},
		{"string", TOKEN_TYPE_STRING},
		{"void", TOKEN_TYPE_VOID},
	}
	for _, tt := range tests {
		tok := NewLexer(tt.input).Next()
		if tok.Type != tt.expected {
			t.Errorf("type keyword %q: got type %d, want %d", tt.input, tok.Type, tt.expected)
		}
	}
}

func TestLexer_BooleanAndNull(t *testing.T) {
	cases := []struct {
		input string
		typ   TokenType
	}{
		{"true", TOKEN_TRUE},
		{"false", TOKEN_FALSE},
		{"null", TOKEN_NULL},
	}
	for _, c := range cases {
		tok := NewLexer(c.input).Next()
		if tok.Type != c.typ {
			t.Errorf("%q: got type %d, want %d", c.input, tok.Type, c.typ)
		}
	}
}

// --- Identifiers ---

func TestLexer_Identifier(t *testing.T) {
	tok := NewLexer("myVar_1").Next()
	if tok.Type != TOKEN_IDENT || tok.Value != "myVar_1" {
		t.Errorf("identifier: got (%d, %q), want (TOKEN_IDENT, %q)", tok.Type, tok.Value, "myVar_1")
	}
}

func TestLexer_IdentifierNotKeyword(t *testing.T) {
	// "fnx" should be an identifier, not the "fn" keyword
	tok := NewLexer("fnx").Next()
	if tok.Type != TOKEN_IDENT {
		t.Errorf("fnx should be IDENT, got %d", tok.Type)
	}
}

// --- Number literals ---

func TestLexer_DecimalInteger(t *testing.T) {
	tok := NewLexer("42").Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "42" {
		t.Errorf("decimal int: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_HexInteger(t *testing.T) {
	tok := NewLexer("0xFF").Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0xFF" {
		t.Errorf("hex int: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_OctalInteger(t *testing.T) {
	tok := NewLexer("0o77").Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0o77" {
		t.Errorf("octal int: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_BinaryInteger(t *testing.T) {
	tok := NewLexer("0b1010").Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0b1010" {
		t.Errorf("binary int: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_FloatLiteral(t *testing.T) {
	tok := NewLexer("3.14").Next()
	if tok.Type != TOKEN_LITERAL_FLOAT || tok.Value != "3.14" {
		t.Errorf("float: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_ZeroIsInt(t *testing.T) {
	tok := NewLexer("0").Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "0" {
		t.Errorf("zero: got (%d, %q)", tok.Type, tok.Value)
	}
}

// --- String and char literals ---

func TestLexer_SimpleString(t *testing.T) {
	tok := NewLexer(`"hello"`).Next()
	if tok.Type != TOKEN_STRING || tok.Value != "hello" {
		t.Errorf("string: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_StringWithEscape(t *testing.T) {
	// The lexer keeps escape sequences raw
	tok := NewLexer(`"line\nbreak"`).Next()
	if tok.Type != TOKEN_STRING || tok.Value != `line\nbreak` {
		t.Errorf("string with escape: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_CharLiteral(t *testing.T) {
	tok := NewLexer("'a'").Next()
	if tok.Type != TOKEN_LITERAL_CHAR || tok.Value != "a" {
		t.Errorf("char literal: got (%d, %q)", tok.Type, tok.Value)
	}
}

func TestLexer_CharLiteralEscape(t *testing.T) {
	tok := NewLexer(`'\n'`).Next()
	if tok.Type != TOKEN_LITERAL_CHAR || tok.Value != `\n` {
		t.Errorf("char escape: got (%d, %q)", tok.Type, tok.Value)
	}
}

// --- Operators ---

func TestLexer_SingleCharOperators(t *testing.T) {
	cases := []struct {
		input string
		typ   TokenType
	}{
		{"+", TOKEN_PLUS},
		{"-", TOKEN_MINUS},
		{"*", TOKEN_MULTIPLY},
		{"/", TOKEN_DIVIDE},
		{"%", TOKEN_MOD},
		{"(", TOKEN_LPAREN},
		{")", TOKEN_RPAREN},
		{"{", TOKEN_LBRACE},
		{"}", TOKEN_RBRACE},
		{"[", TOKEN_LBRACKET},
		{"]", TOKEN_RBRACKET},
		{";", TOKEN_SEMICOLON},
		{",", TOKEN_COMMA},
		{".", TOKEN_DOT},
		{"?", TOKEN_QUESTION},
		{"^", TOKEN_XOR},
		{"~", TOKEN_TILDE},
		{"$", TOKEN_PREFIX_REF},
		{"@", TOKEN_AT},
	}
	for _, c := range cases {
		tok := NewLexer(c.input).Next()
		if tok.Type != c.typ {
			t.Errorf("%q: got type %d, want %d", c.input, tok.Type, c.typ)
		}
	}
}

func TestLexer_MultiCharOperators(t *testing.T) {
	cases := []struct {
		input string
		typ   TokenType
		val   string
	}{
		{"==", TOKEN_EQ, "=="},
		{"!=", TOKEN_NE, "!="},
		{"<=", TOKEN_LE, "<="},
		{">=", TOKEN_GE, ">="},
		{"<<", TOKEN_LSHIFT, "<<"},
		{">>", TOKEN_RSHIFT, ">>"},
		{"&&", TOKEN_AND, "&&"},
		{"||", TOKEN_OR, "||"},
		{"::", TOKEN_DOUBLE_COLON, "::"},
		{"=>", TOKEN_ARROW, "=>"},
	}
	for _, c := range cases {
		tok := NewLexer(c.input).Next()
		if tok.Type != c.typ || tok.Value != c.val {
			t.Errorf("%q: got (%d, %q), want (%d, %q)", c.input, tok.Type, tok.Value, c.typ, c.val)
		}
	}
}

func TestLexer_ColonVsDoubleColon(t *testing.T) {
	tokens := collectTokens(": ::")
	// Expect: COLON, DOUBLE_COLON, EOF
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TOKEN_COLON {
		t.Errorf("first token: got %d, want TOKEN_COLON", tokens[0].Type)
	}
	if tokens[1].Type != TOKEN_DOUBLE_COLON {
		t.Errorf("second token: got %d, want TOKEN_DOUBLE_COLON", tokens[1].Type)
	}
}

// --- Comments ---

func TestLexer_LineCommentSkipped(t *testing.T) {
	tokens := collectTokens("a // comment\nb")
	// Expect: IDENT(a), IDENT(b), EOF
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %+v", len(tokens), tokens)
	}
	if tokens[0].Value != "a" || tokens[1].Value != "b" {
		t.Errorf("unexpected values: %q, %q", tokens[0].Value, tokens[1].Value)
	}
}

func TestLexer_HashCommentSkipped(t *testing.T) {
	tokens := collectTokens("a # comment\nb")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
}

func TestLexer_Attribute(t *testing.T) {
	tok := NewLexer("#[sor]").Next()
	if tok.Type != TOKEN_ATTRIBUTE || tok.Value != "#[sor]" {
		t.Errorf("attribute: got (%d, %q), want (TOKEN_ATTRIBUTE, #[sor])", tok.Type, tok.Value)
	}
}

// --- Line/column tracking ---

func TestLexer_LineTracking(t *testing.T) {
	tokens := collectTokens("a\nb\nc")
	// a=line1, b=line2, c=line3
	if tokens[0].Line != 1 {
		t.Errorf("a line: got %d, want 1", tokens[0].Line)
	}
	if tokens[1].Line != 2 {
		t.Errorf("b line: got %d, want 2", tokens[1].Line)
	}
	if tokens[2].Line != 3 {
		t.Errorf("c line: got %d, want 3", tokens[2].Line)
	}
}

// --- SaveState / RestoreState ---

func TestLexer_SaveRestoreState(t *testing.T) {
	l := NewLexer("a b c")
	tok1 := l.Next() // a
	if tok1.Value != "a" {
		t.Fatalf("expected 'a', got %q", tok1.Value)
	}
	state := l.SaveState()
	tok2 := l.Next() // b
	if tok2.Value != "b" {
		t.Fatalf("expected 'b', got %q", tok2.Value)
	}
	l.RestoreState(state)
	tok2again := l.Next() // should be b again
	if tok2again.Value != "b" {
		t.Errorf("after restore expected 'b', got %q", tok2again.Value)
	}
}

// --- Mixed expression ---

func TestLexer_SimpleExpression(t *testing.T) {
	tokens := collectTokens("x = 10 + y")
	// Expected: IDENT(x) ASSIGN INT(10) PLUS IDENT(y) EOF
	expected := []struct {
		typ TokenType
		val string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_LITERAL_INT, "10"},
		{TOKEN_PLUS, "+"},
		{TOKEN_IDENT, "y"},
		{TOKEN_EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, e := range expected {
		if tokens[i].Type != e.typ || tokens[i].Value != e.val {
			t.Errorf("token[%d]: got (%d, %q), want (%d, %q)", i, tokens[i].Type, tokens[i].Value, e.typ, e.val)
		}
	}
}

// --- Edge cases ---

func TestLexer_EmptyInput(t *testing.T) {
	tok := NewLexer("").Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("empty input should yield EOF, got %d", tok.Type)
	}
}

func TestLexer_OnlyWhitespace(t *testing.T) {
	tok := NewLexer("   \n\t\n  ").Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("whitespace-only input should yield EOF, got %d", tok.Type)
	}
}
