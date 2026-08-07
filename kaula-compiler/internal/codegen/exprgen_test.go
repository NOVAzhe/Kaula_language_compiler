package codegen

import (
	"kaula-compiler/internal/ast"
	"testing"
)

// ---- kaulaOpToCOp ----

func TestKaulaOpToCOp(t *testing.T) {
	tests := []struct {
		kaula string
		want  string
	}{
		{"+", "+"},
		{"PLUS", "+"},
		{"-", "-"},
		{"MINUS", "-"},
		{"*", "*"},
		{"MULTIPLY", "*"},
		{"/", "/"},
		{"DIVIDE", "/"},
		{"%", "%"},
		{"MOD", "%"},
		{"==", "=="},
		{"EQ", "=="},
		{"!=", "!="},
		{"NE", "!="},
		{"<", "<"},
		{"LT", "<"},
		{">", ">"},
		{"GT", ">"},
		{"<=", "<="},
		{"LE", "<="},
		{">=", ">="},
		{"GE", ">="},
		{"<<", "<<"},
		{"LSHIFT", "<<"},
		{"SHIFT_LEFT", "<<"},
		{">>", ">>"},
		{"RSHIFT", ">>"},
		{"SHIFT_RIGHT", ">>"},
		{"&&", "&&"},
		{"AND", "&&"},
		{"||", "||"},
		{"OR", "||"},
		{"&", "&"},
		{"BITWISE_AND", "&"},
		{"AMPERSAND", "&"},
		{"|", "|"},
		{"BITWISE_OR", "|"},
		{"PIPE", "|"},
		{"^", "^"},
		{"BITWISE_XOR", "^"},
		{"CARET", "^"},
		{"XOR", "^"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.kaula, func(t *testing.T) {
			got := kaulaOpToCOp(tt.kaula)
			if got != tt.want {
				t.Errorf("kaulaOpToCOp(%q) = %q, want %q", tt.kaula, got, tt.want)
			}
		})
	}
}

// ---- escapeCString ----

func TestEscapeCString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"line1\nline2", "line1\\nline2"},
		{"tab\there", "tab\\there"},
		{"carriage\rreturn", "carriage\\rreturn"},
		{"null\x00byte", "null\\0byte"},
		{"", ""},
		{"already safe", "already safe"},
		{"multi\nline\r\nend", "multi\\nline\\r\\nend"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeCString(tt.input)
			if got != tt.want {
				t.Errorf("escapeCString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---- escapeCIdentifier ----

func TestEscapeCIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal", "normal"},
		{"with_underscore", "with_underscore"},
		{"with123", "with123"},
		{"123starting_with_digit", "_123starting_with_digit"},
		{"special!@#$chars", "specialchars"},
		{"", "_invalid"},
		{"a", "a"},
		{" mixed ", "mixed"},
		{"dash-ed-name", "dashedname"},
		{"unicodeλ", "unicode"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeCIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("escapeCIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---- isIntegerLiteral ----

func TestIsIntegerLiteral(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"123", true},
		{"0", true},
		{"-5", true},
		{"+10", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"3.14", false},
		// Note: code treats "-" and "+" as valid integer literals
		{"-", true},
		{"+", true},
		{"  123", false}, // spaces not allowed
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := isIntegerLiteral(tt.s)
			if got != tt.want {
				t.Errorf("isIntegerLiteral(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// ---- wrapIfNeeded ----

func TestWrapIfNeeded_NoWrap(t *testing.T) {
	// Simple identifier should not be wrapped
	got := wrapIfNeeded("x", "+", "left")
	if got != "x" {
		t.Errorf("expected 'x', got %q", got)
	}
}

func TestWrapIfNeeded_LeftWrap(t *testing.T) {
	// a + b as left operand of * should be wrapped: (a + b) * c
	got := wrapIfNeeded("a + b", "*", "left")
	if got != "(a + b)" {
		t.Errorf("expected '(a + b)', got %q", got)
	}
}

func TestWrapIfNeeded_RightWrap(t *testing.T) {
	// a * b as right operand of + should be wrapped: c + (a * b)... actually no,
	// multiplication has higher precedence, so a * b + c would be fine without parens
	// But for right side: a + b as right operand of * should be wrapped: c * (a + b)
	got := wrapIfNeeded("a + b", "*", "right")
	if got != "(a + b)" {
		t.Errorf("expected '(a + b)', got %q", got)
	}
}

func TestWrapIfNeeded_NoWrapHigherPrec(t *testing.T) {
	// a * b as left operand of + should NOT be wrapped: a * b + c
	got := wrapIfNeeded("a * b", "+", "left")
	if got != "a * b" {
		t.Errorf("expected 'a * b', got %q", got)
	}
}

func TestWrapIfNeeded_SemicolonEnd(t *testing.T) {
	// Expressions ending with ; should not be wrapped
	got := wrapIfNeeded("x;", "+", "left")
	if got != "x;" {
		t.Errorf("expected 'x;', got %q", got)
	}
}

func TestWrapIfNeeded_RightBitwiseSpecial(t *testing.T) {
	// Right operand of & with ~ should be wrapped
	got := wrapIfNeeded("~x", "&", "right")
	if got != "(~x)" {
		t.Errorf("expected '(~x)', got %q", got)
	}
}

func TestWrapIfNeeded_RightShiftSpecial(t *testing.T) {
	// Right operand of & with << should be wrapped
	got := wrapIfNeeded("1 << n", "&", "right")
	if got != "(1 << n)" {
		t.Errorf("expected '(1 << n)', got %q", got)
	}
}

func TestWrapIfNeeded_Empty(t *testing.T) {
	got := wrapIfNeeded("", "+", "left")
	if got != "" {
		t.Errorf("expected '', got %q", got)
	}
}

// ---- cOperatorPrecedence ----

func TestCOperatorPrecedence_Values(t *testing.T) {
	// Verify that precedence ordering is correct
	if cOperatorPrecedence["*"] <= cOperatorPrecedence["+"] {
		t.Error("* should have higher precedence than +")
	}
	if cOperatorPrecedence["+"] <= cOperatorPrecedence["=="] {
		t.Error("+ should have higher precedence than ==")
	}
	if cOperatorPrecedence["=="] <= cOperatorPrecedence["&&"] {
		t.Error("== should have higher precedence than &&")
	}
	if cOperatorPrecedence["&&"] <= cOperatorPrecedence["||"] {
		t.Error("&& should have higher precedence than ||")
	}
	if cOperatorPrecedence["<<"] >= cOperatorPrecedence["+"] {
		t.Error("<< should have lower precedence than +")
	}
	if cOperatorPrecedence["*"] != cOperatorPrecedence["/"] {
		t.Errorf("* (%d) and / (%d) should have equal precedence", cOperatorPrecedence["*"], cOperatorPrecedence["/"])
	}
}

// ---- astBinaryNeedsParens ----

func TestAstBinaryNeedsParens_NotBinary(t *testing.T) {
	// Non-binary child should not need parens
	ident := &ast.Identifier{Name: "x"}
	if astBinaryNeedsParens(ident, "+", "left") {
		t.Error("identifier should not need parens")
	}
}

func TestAstBinaryNeedsParens_Nil(t *testing.T) {
	if astBinaryNeedsParens(nil, "+", "left") {
		t.Error("nil should not need parens")
	}
}

func TestAstBinaryNeedsParens_LeftNeedsWrapping(t *testing.T) {
	// (a + b) * c: a+b as left operand of * should need parens
	inner := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "a"},
		Operator: "+",
		Right:    &ast.Identifier{Name: "b"},
	}
	if !astBinaryNeedsParens(inner, "*", "left") {
		t.Error("a+b as left operand of * should need parens")
	}
}

func TestAstBinaryNeedsParens_RightNeedsWrapping(t *testing.T) {
	// a * (b + c): b+c as right operand of * should need parens
	inner := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "b"},
		Operator: "+",
		Right:    &ast.Identifier{Name: "c"},
	}
	if !astBinaryNeedsParens(inner, "*", "right") {
		t.Error("b+c as right operand of * should need parens")
	}
}

func TestAstBinaryNeedsParens_NoWrapHigherPrec(t *testing.T) {
	// a * b + c: a*b as left operand of + should NOT need parens
	inner := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "a"},
		Operator: "*",
		Right:    &ast.Identifier{Name: "b"},
	}
	if astBinaryNeedsParens(inner, "+", "left") {
		t.Error("a*b as left operand of + should not need parens")
	}
}

func TestAstBinaryNeedsParens_UnknownOp(t *testing.T) {
	// Unknown operator should not force parens
	inner := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "a"},
		Operator: "UNKNOWN",
		Right:    &ast.Identifier{Name: "b"},
	}
	if astBinaryNeedsParens(inner, "+", "left") {
		t.Error("unknown operator should not need parens")
	}
}

// ---- sortedOps order ----

func TestSortedOps_LongBeforeShort(t *testing.T) {
	// Verify that == comes before = in sortedOps
	eqIdx := -1
	ltIdx := -1
	for i, op := range sortedOps {
		if op == "==" {
			eqIdx = i
		}
		if op == "<" {
			ltIdx = i
		}
	}
	if eqIdx < 0 {
		t.Fatal("sortedOps missing ==")
	}
	if ltIdx < 0 {
		t.Fatal("sortedOps missing <")
	}
	if eqIdx >= ltIdx {
		t.Error("== should come before < in sortedOps (longer match first)")
	}
}

func TestSortedOps_ContainsAll(t *testing.T) {
	requiredOps := []string{"==", "!=", "<=", ">=", "<<", ">>", "&&", "||",
		"+", "-", "*", "/", "%", "|", "^", "&", "<", ">"}
	for _, op := range requiredOps {
		found := false
		for _, s := range sortedOps {
			if s == op {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("sortedOps missing required operator: %q", op)
		}
	}
}