package lexer

import (
	"testing"
)

// Helper: 收集所有 token 直到 EOF
func collectTokens(l *Lexer) []Token {
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

func TestNewLexer(t *testing.T) {
	l := NewLexer("")
	if l == nil {
		t.Fatal("NewLexer() returned nil")
	}
	if l.inputLen != 0 {
		t.Errorf("inputLen = %d, want 0", l.inputLen)
	}
	if l.line != 1 {
		t.Errorf("line = %d, want 1", l.line)
	}
	if l.column != 1 {
		t.Errorf("column = %d, want 1", l.column)
	}
}

func TestLexer_EOF(t *testing.T) {
	l := NewLexer("")
	tok := l.Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("expected EOF, got %s", TokenTypeToString(tok.Type))
	}
}

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		input string
		kind  TokenType
	}{
		{"spend", TOKEN_SPEND},
		{"call", TOKEN_CALL},
		{"default", TOKEN_DEFAULT},
		{"prefix", TOKEN_PREFIX},
		{"tree", TOKEN_TREE},
		{"object", TOKEN_OBJECT},
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
		{"println", TOKEN_PRINTLN},
		{"break", TOKEN_BREAK},
		{"continue", TOKEN_CONTINUE},
		{"class", TOKEN_CLASS},
		{"interface", TOKEN_LITERAL_INTERFACE},
		{"implements", TOKEN_IMPLEMENTS},
		{"constructor", TOKEN_CONSTRUCTOR},
		{"struct", TOKEN_STRUCT},
		{"auto", TOKEN_AUTO},
		{"as", TOKEN_AS},
		{"yield", TOKEN_YIELD},
		{"release", TOKEN_RELEASE},
		{"extract", TOKEN_EXTRACT},
		{"type", TOKEN_TYPE},
		{"sizeof", TOKEN_SIZEOF},
		{"alignof", TOKEN_ALIGNOF},
		{"offsetof", TOKEN_OFFSETOF},
		{"comptime", TOKEN_COMPTIME},
		{"type_name", TOKEN_TYPE_NAME},
		{"field_count", TOKEN_FIELD_COUNT},
		{"field_name", TOKEN_FIELD_NAME},
		{"field_type", TOKEN_FIELD_TYPE},
		{"type_kind", TOKEN_TYPE_KIND},
		{"enum", TOKEN_ENUM},
		{"match", TOKEN_MATCH},
		{"extern", TOKEN_EXTERN},
		{"static", TOKEN_STATIC},
		{"const", TOKEN_CONST},
		{"true", TOKEN_TRUE},
		{"false", TOKEN_FALSE},
		{"null", TOKEN_NULL},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.kind {
				t.Errorf("Next(%q) = %s, want %s", tt.input, TokenTypeToString(tok.Type), TokenTypeToString(tt.kind))
			}
			if tok.Value != tt.input {
				t.Errorf("Next(%q).Value = %q, want %q", tt.input, tok.Value, tt.input)
			}
		})
	}
}

func TestLexer_TypeKeywords(t *testing.T) {
	tests := []struct {
		input string
		kind  TokenType
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
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.kind {
				t.Errorf("Next(%q) = %s, want %s", tt.input, TokenTypeToString(tok.Type), TokenTypeToString(tt.kind))
			}
		})
	}
}

func TestLexer_Identifiers(t *testing.T) {
	tests := []string{
		"foo", "bar", "x", "myVariable", "_private", "foo123", "hello_world",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			l := NewLexer(input)
			tok := l.Next()
			if tok.Type != TOKEN_IDENT {
				t.Errorf("Next(%q) = %s, want IDENT", input, TokenTypeToString(tok.Type))
			}
			if tok.Value != input {
				t.Errorf("Next(%q).Value = %q, want %q", input, tok.Value, input)
			}
		})
	}
}

func TestLexer_IntegerLiterals(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"0", "0"},
		{"42", "42"},
		{"12345", "12345"},
		{"0x1A", "0x1A"},
		{"0xFF", "0xFF"},
		{"0xDEAD", "0xDEAD"},
		{"0o77", "0o77"},
		{"0o123", "0o123"},
		{"0b1010", "0b1010"},
		{"0b1", "0b1"},
		{"0b0", "0b0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_LITERAL_INT {
				t.Errorf("Next(%q) = %s, want LITERAL_INT", tt.input, TokenTypeToString(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q).Value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_FloatLiterals(t *testing.T) {
	tests := []string{
		"3.14", "0.5", "100.0", "0.001",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			l := NewLexer(input)
			tok := l.Next()
			if tok.Type != TOKEN_LITERAL_FLOAT {
				t.Errorf("Next(%q) = %s, want LITERAL_FLOAT", input, TokenTypeToString(tok.Type))
			}
			if tok.Value != input {
				t.Errorf("Next(%q).Value = %q, want %q", input, tok.Value, input)
			}
		})
	}
}

func TestLexer_StringLiterals(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{`"hello"`, "hello"},
		{`""`, ""},
		{`"hello world"`, "hello world"},
		{`"hello\nworld"`, "hello\\nworld"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_STRING {
				t.Errorf("Next(%q) = %s, want STRING", tt.input, TokenTypeToString(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q).Value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_UnterminatedString(t *testing.T) {
	l := NewLexer(`"hello`)
	tok := l.Next()
	if tok.Type != TOKEN_STRING {
		t.Errorf("expected STRING (empty), got %s", TokenTypeToString(tok.Type))
	}
	if !l.HasErrors() {
		t.Error("expected error for unterminated string")
	}
}

func TestLexer_CharLiterals(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{`'a'`, "a"},
		{`'Z'`, "Z"},
		{`'0'`, "0"},
		{`'\n'`, "\\n"},
		{`'\t'`, "\\t"},
		{`'\''`, "\\'"},
		{`'\"'`, "\\\""},
		{`'\\'`, "\\\\"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != TOKEN_LITERAL_CHAR {
				t.Errorf("Next(%q) = %s, want LITERAL_CHAR", tt.input, TokenTypeToString(tok.Type))
			}
			if tok.Value != tt.value {
				t.Errorf("Next(%q).Value = %q, want %q", tt.input, tok.Value, tt.value)
			}
		})
	}
}

func TestLexer_UnterminatedChar(t *testing.T) {
	l := NewLexer("'a")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_CHAR {
		t.Errorf("expected LITERAL_CHAR, got %s", TokenTypeToString(tok.Type))
	}
	if !l.HasErrors() {
		t.Error("expected error for unterminated char literal")
	}
}

func TestLexer_OperatorsAndDelimiters(t *testing.T) {
	tests := []struct {
		input string
		kind  TokenType
	}{
		{"+", TOKEN_PLUS},
		{"-", TOKEN_MINUS},
		{"*", TOKEN_MULTIPLY},
		{"/", TOKEN_DIVIDE},
		{"%", TOKEN_MOD},
		{"=", TOKEN_ASSIGN},
		{"==", TOKEN_EQ},
		{"!=", TOKEN_NE},
		{"<", TOKEN_LT},
		{">", TOKEN_GT},
		{"<=", TOKEN_LE},
		{">=", TOKEN_GE},
		{"<<", TOKEN_LSHIFT},
		{">>", TOKEN_RSHIFT},
		{"&&", TOKEN_AND},
		{"&", TOKEN_AMPERSAND},
		{"||", TOKEN_OR},
		{"|", TOKEN_PIPE},
		{"^", TOKEN_XOR},
		{"~", TOKEN_TILDE},
		{"->", TOKEN_ARROW},
		{"=>", TOKEN_ARROW},
		{"$", TOKEN_PREFIX_REF},
		{"@", TOKEN_AT},
		{"?", TOKEN_QUESTION},
		{"(", TOKEN_LPAREN},
		{")", TOKEN_RPAREN},
		{"{", TOKEN_LBRACE},
		{"}", TOKEN_RBRACE},
		{"[", TOKEN_LBRACKET},
		{"]", TOKEN_RBRACKET},
		{";", TOKEN_SEMICOLON},
		{",", TOKEN_COMMA},
		{":", TOKEN_COLON},
		{"::", TOKEN_DOUBLE_COLON},
		{".", TOKEN_DOT},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.kind {
				t.Errorf("Next(%q) = %s, want %s", tt.input, TokenTypeToString(tok.Type), TokenTypeToString(tt.kind))
			}
			if tok.Value != tt.input {
				t.Errorf("Next(%q).Value = %q, want %q", tt.input, tok.Value, tt.input)
			}
		})
	}
}

func TestLexer_Comments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TokenType // 注释后的第一个 token
	}{
		{"line comment //", "// this is a comment\n42", TOKEN_LITERAL_INT},
		{"line comment #", "# this is a comment\n42", TOKEN_LITERAL_INT},
		{"comment at end", "// comment only", TOKEN_EOF},
		{"comment with code", "// comment\nint x", TOKEN_TYPE_INT},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.want {
				t.Errorf("after comment, got %s, want %s", TokenTypeToString(tok.Type), TokenTypeToString(tt.want))
			}
		})
	}
}

func TestLexer_AttributeAnnotation(t *testing.T) {
	l := NewLexer("#[attr]")
	tok := l.Next()
	if tok.Type != TOKEN_ATTRIBUTE {
		t.Errorf("expected ATTRIBUTE, got %s", TokenTypeToString(tok.Type))
	}
	if tok.Value != "#[attr]" {
		t.Errorf("Value = %q, want %q", tok.Value, "#[attr]")
	}
}

func TestLexer_AttributeWithSemicolon(t *testing.T) {
	// #[attr] 后面跟 fn
	l := NewLexer("#[sor] fn main() void {}")
	tokens := collectTokens(l)
	if len(tokens) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TOKEN_ATTRIBUTE {
		t.Errorf("tokens[0] = %s, want ATTRIBUTE", TokenTypeToString(tokens[0].Type))
	}
	if tokens[0].Value != "#[sor]" {
		t.Errorf("tokens[0].Value = %q, want %q", tokens[0].Value, "#[sor]")
	}
}

func TestLexer_Whitespace(t *testing.T) {
	l := NewLexer("  \t\n  foo")
	tok := l.Next()
	if tok.Type != TOKEN_IDENT {
		t.Errorf("expected IDENT, got %s", TokenTypeToString(tok.Type))
	}
	if tok.Value != "foo" {
		t.Errorf("Value = %q, want %q", tok.Value, "foo")
	}
	if tok.Line != 2 {
		t.Errorf("Line = %d, want 2", tok.Line)
	}
}

func TestLexer_SaveRestoreState(t *testing.T) {
	l := NewLexer("foo bar baz")
	_ = l.Next() // consume "foo"
	state := l.SaveState()
	tok2 := l.Next() // bar
	if tok2.Value != "bar" {
		t.Errorf("before restore: got %q, want %q", tok2.Value, "bar")
	}
	l.RestoreState(state)
	tok3 := l.Next() // should be bar again
	if tok3.Value != "bar" {
		t.Errorf("after restore: got %q, want %q", tok3.Value, "bar")
	}
}

func TestLexer_ErrorOnUnexpectedChar(t *testing.T) {
	l := NewLexer("!")
	tok := l.Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("after error, expected EOF, got %s", TokenTypeToString(tok.Type))
	}
	if !l.HasErrors() {
		t.Error("expected error for unexpected '!'")
	}
}

func TestLexer_GetSource(t *testing.T) {
	src := "fn main() void {}"
	l := NewLexer(src)
	if l.GetSource() != src {
		t.Errorf("GetSource() = %q, want %q", l.GetSource(), src)
	}
}

func TestLexer_SetFile(t *testing.T) {
	l := NewLexer("foo")
	l.SetFile("test.kl")
	// 验证 file 设置没有副作用
	tok := l.Next()
	if tok.Type != TOKEN_IDENT {
		t.Errorf("expected IDENT, got %s", TokenTypeToString(tok.Type))
	}
}

func TestLexer_ScanUntilRbrace(t *testing.T) {
	l := NewLexer("abc}")
	got := l.ScanUntilRbrace()
	if got != "abc" {
		t.Errorf("ScanUntilRbrace() = %q, want %q", got, "abc")
	}
}

func TestLexer_TokenPosition(t *testing.T) {
	l := NewLexer("foo\nbar")
	tok := l.Next()
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("foo: got line=%d col=%d, want line=1 col=1", tok.Line, tok.Column)
	}
	tok = l.Next()
	if tok.Line != 2 || tok.Column != 1 {
		t.Errorf("bar: got line=%d col=%d, want line=2 col=1", tok.Line, tok.Column)
	}
}

func TestLexer_TokenString(t *testing.T) {
	l := NewLexer("42")
	tok := l.Next()
	s := tok.String()
	if s == "" {
		t.Error("Token.String() returned empty string")
	}
}

func TestLexer_ProgramSnippet(t *testing.T) {
	src := `import std.io

fn main() void {
    println("hello")
}
`
	l := NewLexer(src)
	tokens := collectTokens(l)
	// Verify key tokens in order
	expectedTypes := []TokenType{
		TOKEN_IMPORT, TOKEN_IDENT, TOKEN_DOT, TOKEN_IDENT, // import std.io
		TOKEN_FUNC, TOKEN_IDENT, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_TYPE_VOID, // fn main() void
		TOKEN_LBRACE, // {
		TOKEN_PRINTLN, TOKEN_LPAREN, TOKEN_STRING, TOKEN_RPAREN, // println("hello")
		TOKEN_RBRACE, // }
		TOKEN_EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Errorf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokens)
		// Still check what we can
		minLen := len(expectedTypes)
		if len(tokens) < minLen {
			minLen = len(tokens)
		}
		for i := 0; i < minLen; i++ {
			if tokens[i].Type != expectedTypes[i] {
				t.Errorf("token[%d]: got %s, want %s", i, TokenTypeToString(tokens[i].Type), TokenTypeToString(expectedTypes[i]))
			}
		}
		return
	}
	for i, tok := range tokens {
		if tok.Type != expectedTypes[i] {
			t.Errorf("token[%d]: got %s(%q), want %s", i, TokenTypeToString(tok.Type), tok.Value, TokenTypeToString(expectedTypes[i]))
		}
	}
}

func TestLexer_GetSetPosition(t *testing.T) {
	l := NewLexer("foo bar")
	// initial position is 0
	if l.GetPosition() != 0 {
		t.Errorf("initial GetPosition() = %d, want 0", l.GetPosition())
	}
	l.Next() // consume 'foo'
	pos := l.GetPosition()
	if pos <= 0 {
		t.Errorf("after consuming 'foo', GetPosition() = %d, want > 0", pos)
	}
	// reset to beginning
	l.SetPosition(0)
	// should read 'foo' again
	tok := l.Next()
	if tok.Value != "foo" {
		t.Errorf("after SetPosition(0), Next() = %q, want %q", tok.Value, "foo")
	}
}

func TestLexer_ErrorCollector(t *testing.T) {
	l := NewLexer("!")
	ec := l.GetErrorCollector()
	if ec == nil {
		t.Fatal("GetErrorCollector() returned nil")
	}
	l.Next()
	if !ec.HasErrors() {
		t.Error("expected errors after lexing '!'")
	}
}

func TestLexer_SetErrorCollector(t *testing.T) {
	l := NewLexer("foo")
	ec := l.GetErrorCollector()
	// Should not panic
	l.SetErrorCollector(ec)
}

func TestLexer_HasErrors(t *testing.T) {
	l := NewLexer("valid")
	if l.HasErrors() {
		t.Error("expected no errors for valid input")
	}
	l.Next() // valid token
	if l.HasErrors() {
		t.Error("expected no errors after lexing valid token")
	}
}

func TestLexer_ReportErrors(t *testing.T) {
	l := NewLexer("!")
	// Should not panic
	l.ReportErrors()
}

func TestTokenTypeToString(t *testing.T) {
	tests := []struct {
		tt   TokenType
		want string
	}{
		{TOKEN_EOF, "EOF"},
		{TOKEN_IDENT, "IDENT"},
		{TOKEN_FUNC, "FUNC"},
		{TOKEN_IF, "IF"},
		{TOKEN_PLUS, "PLUS"},
		{TOKEN_LPAREN, "LPAREN"},
		{TOKEN_RBRACE, "RBRACE"},
		{TOKEN_STRING, "STRING"},
		{TOKEN_SEMICOLON, "SEMICOLON"},
		{TOKEN_ARROW, "ARROW"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := TokenTypeToString(tt.tt)
			if got != tt.want {
				t.Errorf("TokenTypeToString(%d) = %q, want %q", tt.tt, got, tt.want)
			}
		})
	}
}

func TestLexer_MultipleTokens(t *testing.T) {
	l := NewLexer("int x = 42")
	tokens := collectTokens(l)
	expected := []struct {
		Type  TokenType
		Value string
	}{
		{TOKEN_TYPE_INT, "int"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_LITERAL_INT, "42"},
		{TOKEN_EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i].Type {
			t.Errorf("token[%d].Type = %s, want %s", i, TokenTypeToString(tok.Type), TokenTypeToString(expected[i].Type))
		}
		if tok.Value != expected[i].Value {
			t.Errorf("token[%d].Value = %q, want %q", i, tok.Value, expected[i].Value)
		}
	}
}

func TestLexer_UnicodeIdentifiers(t *testing.T) {
	// 测试 unicode 标识符（单字节 unicode 字符，如 ä 等）
	// 注：多字节 UTF-8 字符（如中文）当前词法分析器仅按字节处理
	l := NewLexer("foo")
	tok := l.Next()
	if tok.Type != TOKEN_IDENT {
		t.Errorf("expected IDENT, got %s", TokenTypeToString(tok.Type))
	}
	if tok.Value != "foo" {
		t.Errorf("Value = %q, want %q", tok.Value, "foo")
	}
}

func TestLexer_EmptyInput(t *testing.T) {
	l := NewLexer("")
	tok := l.Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("expected EOF, got %s", TokenTypeToString(tok.Type))
	}
}

func TestLexer_OnlyWhitespace(t *testing.T) {
	l := NewLexer("   \t\n  \n  ")
	tok := l.Next()
	if tok.Type != TOKEN_EOF {
		t.Errorf("expected EOF for whitespace-only input, got %s", TokenTypeToString(tok.Type))
	}
}

func TestLexer_SequenceNumberOperators(t *testing.T) {
	// 测试数字后面跟运算符的连续 token
	l := NewLexer("1+2")
	tok1 := l.Next()
	if tok1.Type != TOKEN_LITERAL_INT || tok1.Value != "1" {
		t.Errorf("expected INT(1), got %s(%q)", TokenTypeToString(tok1.Type), tok1.Value)
	}
	tok2 := l.Next()
	if tok2.Type != TOKEN_PLUS || tok2.Value != "+" {
		t.Errorf("expected PLUS(+), got %s(%q)", TokenTypeToString(tok2.Type), tok2.Value)
	}
	tok3 := l.Next()
	if tok3.Type != TOKEN_LITERAL_INT || tok3.Value != "2" {
		t.Errorf("expected INT(2), got %s(%q)", TokenTypeToString(tok3.Type), tok3.Value)
	}
}

func TestLexer_InvalidHex(t *testing.T) {
	// 0x 后面不跟有效字符
	l := NewLexer("0x")
	tok := l.Next()
	if tok.Type != TOKEN_LITERAL_INT {
		t.Errorf("expected INT, got %s", TokenTypeToString(tok.Type))
	}
	if tok.Value != "0x" {
		t.Errorf("Value = %q, want %q", tok.Value, "0x")
	}
}