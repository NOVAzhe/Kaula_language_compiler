package lexer

import (
	"testing"
)

func tokenTypeName(tt TokenType) string {
	return TokenTypeToString(tt)
}

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
	}{
		{"fn", TOKEN_FUNC},
		{"if", TOKEN_IF},
		{"else", TOKEN_ELSE},
		{"while", TOKEN_WHILE},
		{"for", TOKEN_FOR},
		{"in", TOKEN_IN},
		{"return", TOKEN_RETURN},
		{"import", TOKEN_IMPORT},
		{"export", TOKEN_EXPORT},
		{"package", TOKEN_PACKAGE},
		{"pub", TOKEN_PUB},
		{"self", TOKEN_SELF},
		{"nonlocal", TOKEN_NONLOCAL},
		{"break", TOKEN_BREAK},
		{"continue", TOKEN_CONTINUE},
		{"class", TOKEN_CLASS},
		{"interface", TOKEN_LITERAL_INTERFACE},
		{"implements", TOKEN_IMPLEMENTS},
		{"constructor", TOKEN_CONSTRUCTOR},
		{"struct", TOKEN_STRUCT},
		{"enum", TOKEN_ENUM},
		{"match", TOKEN_MATCH},
		{"auto", TOKEN_AUTO},
		{"as", TOKEN_AS},
		{"yield", TOKEN_YIELD},
		{"release", TOKEN_RELEASE},
		{"extract", TOKEN_EXTRACT},
		{"spend", TOKEN_SPEND},
		{"call", TOKEN_CALL},
		{"default", TOKEN_DEFAULT},
		{"prefix", TOKEN_PREFIX},
		{"tree", TOKEN_TREE},
		{"object", TOKEN_OBJECT},
		{"extern", TOKEN_EXTERN},
		{"static", TOKEN_STATIC},
		{"const", TOKEN_CONST},
		{"type", TOKEN_TYPE},
		{"sizeof", TOKEN_SIZEOF},
		{"alignof", TOKEN_ALIGNOF},
		{"offsetof", TOKEN_OFFSETOF},
		{"comptime", TOKEN_COMPTIME},
		{"true", TOKEN_TRUE},
		{"false", TOKEN_FALSE},
		{"null", TOKEN_NULL},
		// Type keywords
		{"int", TOKEN_TYPE_INT},
		{"float", TOKEN_TYPE_FLOAT},
		{"double", TOKEN_TYPE_DOUBLE},
		{"bool", TOKEN_TYPE_BOOL},
		{"char", TOKEN_TYPE_CHAR},
		{"string", TOKEN_TYPE_STRING},
		{"void", TOKEN_TYPE_VOID},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.typ {
				t.Errorf("Next(%q) = %s, want %s", tt.input, tokenTypeName(tok.Type), tokenTypeName(tt.typ))
			}
			// Should be followed by EOF
			eof := l.Next()
			if eof.Type != TOKEN_EOF {
				t.Errorf("after %q, expected EOF, got %s", tt.input, tokenTypeName(eof.Type))
			}
		})
	}
}

func TestLexer_Identifiers(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"x", "x"},
		{"myVar", "myVar"},
		{"_private", "_private"},
		{"camelCase", "camelCase"},
		{"PascalCase", "PascalCase"},
		{"snake_case", "snake_case"},
		{"a123", "a123"},
		{"_123", "_123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_IDENT {
				t.Errorf("Next(%q) = %s, want IDENT", tt.input, tokenTypeName(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_Integers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		value string
		typ   TokenType
	}{
		{"decimal", "42", "42", TOKEN_LITERAL_INT},
		{"zero", "0", "0", TOKEN_LITERAL_INT},
		{"hex", "0xFF", "0xFF", TOKEN_LITERAL_INT},
		{"hex_lower", "0xdeadbeef", "0xdeadbeef", TOKEN_LITERAL_INT},
		{"octal", "0o777", "0o777", TOKEN_LITERAL_INT},
		{"binary", "0b1010", "0b1010", TOKEN_LITERAL_INT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.typ {
				t.Errorf("Next(%q) = %s, want %s", tt.input, tokenTypeName(tok.Type), tokenTypeName(tt.typ))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_Floats(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"10.0", "10.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_LITERAL_FLOAT {
				t.Errorf("Next(%q) = %s, want FLOAT", tt.input, tokenTypeName(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_Strings(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{`"hello"`, "hello"},
		{`""`, ""},
		{`"hello world"`, "hello world"},
		{`"quote\"inside"`, `quote\"inside`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_STRING {
				t.Errorf("Next(%q) = %s, want STRING", tt.input, tokenTypeName(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_CharLiterals(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"'a'", "a"},
		{"'\\n'", "\\n"},
		{"'\\t'", "\\t"},
		{"'\\''", "\\'"},
		{"'\\\\'", "\\\\"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_LITERAL_CHAR {
				t.Errorf("Next(%q) = %s, want CHAR", tt.input, tokenTypeName(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_Operators(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
		value string
	}{
		{"+", TOKEN_PLUS, "+"},
		{"-", TOKEN_MINUS, "-"},
		{"*", TOKEN_MULTIPLY, "*"},
		{"/", TOKEN_DIVIDE, "/"},
		{"%", TOKEN_MOD, "%"},
		{"=", TOKEN_ASSIGN, "="},
		{"==", TOKEN_EQ, "=="},
		{"!=", TOKEN_NE, "!="},
		{"<", TOKEN_LT, "<"},
		{">", TOKEN_GT, ">"},
		{"<=", TOKEN_LE, "<="},
		{">=", TOKEN_GE, ">="},
		{"&&", TOKEN_AND, "&&"},
		{"||", TOKEN_OR, "||"},
		{"&", TOKEN_AMPERSAND, "&"},
		{"|", TOKEN_PIPE, "|"},
		{"^", TOKEN_XOR, "^"},
		{"~", TOKEN_TILDE, "~"},
		{"<<", TOKEN_LSHIFT, "<<"},
		{">>", TOKEN_RSHIFT, ">>"},
		{"->", TOKEN_ARROW, "->"},
		{"=>", TOKEN_ARROW, "=>"},
		{"$", TOKEN_PREFIX_REF, "$"},
		{"@", TOKEN_AT, "@"},
		{"?", TOKEN_QUESTION, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.typ {
				t.Errorf("Next(%q) = %s, want %s", tt.input, tokenTypeName(tok.Type), tokenTypeName(tt.typ))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_Delimiters(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
		value string
	}{
		{"(", TOKEN_LPAREN, "("},
		{")", TOKEN_RPAREN, ")"},
		{"{", TOKEN_LBRACE, "{"},
		{"}", TOKEN_RBRACE, "}"},
		{"[", TOKEN_LBRACKET, "["},
		{"]", TOKEN_RBRACKET, "]"},
		{";", TOKEN_SEMICOLON, ";"},
		{",", TOKEN_COMMA, ","},
		{":", TOKEN_COLON, ":"},
		{"::", TOKEN_DOUBLE_COLON, "::"},
		{".", TOKEN_DOT, "."},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.typ {
				t.Errorf("Next(%q) = %s, want %s", tt.input, tokenTypeName(tok.Type), tokenTypeName(tt.typ))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q) value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_SkipComments(t *testing.T) {
	input := `// this is a comment
42`
	l := NewLexer(input)
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_INT || tok.Value != "42" {
		t.Errorf("Expected int 42 after comment, got %s(%q)", tokenTypeName(tok.Type), tok.Value)
	}
}

func TestLexer_MultiLineProgram(t *testing.T) {
	input := `fn main() int {
    int x = 42
    return x
}`
	l := NewLexer(input)
	tokens := []TokenType{}
	for {
		tok := l.Next()
		if tok.Type == TOKEN_EOF {
			break
		}
		tokens = append(tokens, tok.Type)
	}

	expected := []TokenType{
		TOKEN_FUNC, TOKEN_IDENT, TOKEN_LPAREN, TOKEN_RPAREN,
		TOKEN_TYPE_INT, TOKEN_LBRACE,
		TOKEN_TYPE_INT, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_LITERAL_INT,
		TOKEN_RETURN, TOKEN_IDENT,
		TOKEN_RBRACE,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
		return
	}
	for i, tok := range tokens {
		if tok != expected[i] {
			t.Errorf("Token[%d] = %s, want %s", i, tokenTypeName(tok), tokenTypeName(expected[i]))
		}
	}
}

func TestLexer_SaveRestoreState(t *testing.T) {
	input := `fn main() { }`
	l := NewLexer(input)

	// Consume first token
	tok1 := l.Next()
	if tok1.Type != TOKEN_FUNC {
		t.Fatalf("First token should be FUNC, got %s", tokenTypeName(tok1.Type))
	}

	// Save state after "fn"
	state := l.SaveState()

	// Consume the rest
	l.Next() // main
	l.Next() // (
	l.Next() // )
	l.Next() // {
	l.Next() // }

	// Restore to after "fn"
	l.RestoreState(state)

	// Should get "main" again
	tok := l.Next()
	if tok.Type != TOKEN_IDENT || tok.Value != "main" {
		t.Errorf("After restore, expected IDENT(main), got %s(%q)", tokenTypeName(tok.Type), tok.Value)
	}
}

func TestLexer_AttributeAnnotation(t *testing.T) {
	input := "#[sor]"
	l := NewLexer(input)
	tok := l.Next()
	if tok.Type != TOKEN_ATTRIBUTE {
		t.Errorf("Expected ATTRIBUTE, got %s", tokenTypeName(tok.Type))
	}
	if tok.Value != "#[sor]" {
		t.Errorf("Expected value '#[sor]', got %q", tok.Value)
	}
}

func TestLexer_UnterminatedString(t *testing.T) {
	input := `"unterminated`
	l := NewLexer(input)
	tok := l.Next()
	if tok.Type != TOKEN_STRING {
		t.Errorf("Expected STRING (with error), got %s", tokenTypeName(tok.Type))
	}
	if !l.HasErrors() {
		t.Error("Expected error for unterminated string")
	}
}

func TestLexer_UnterminatedChar(t *testing.T) {
	input := `'a`
	l := NewLexer(input)
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_CHAR {
		t.Errorf("Expected CHAR, got %s", tokenTypeName(tok.Type))
	}
	if !l.HasErrors() {
		t.Error("Expected error for unterminated char literal")
	}
}

func TestLexer_ScanUntilRbrace(t *testing.T) {
	input := `some raw content }`
	l := NewLexer(input)
	result := l.ScanUntilRbrace()
	if result != "some raw content " {
		t.Errorf("ScanUntilRbrace = %q, want %q", result, "some raw content ")
	}
}

func TestLexer_Println(t *testing.T) {
	l := NewLexer("println")
	tok := l.Next()
	if tok.Type != TOKEN_PRINTLN {
		t.Errorf("Expected PRINTLN, got %s", tokenTypeName(tok.Type))
	}
}

func TestLexer_PositionTracking(t *testing.T) {
	input := "fn\n  x"
	l := NewLexer(input)

	tok := l.Next() // fn at line 1, col 1
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("fn: expected (1,1), got (%d,%d)", tok.Line, tok.Column)
	}

	tok = l.Next() // x at line 2, col 3
	if tok.Line != 2 || tok.Column != 3 {
		t.Errorf("x: expected (2,3), got (%d,%d)", tok.Line, tok.Column)
	}
}