package errors

import (
	"strings"
	"testing"
)

func TestErrorCollector_AddAndRetrieve(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddSyntaxError("unexpected token", 1, 5, "test.kl", "")
	ec.AddSemanticError("undefined variable", 2, 3, "test.kl", "")
	ec.AddTypeError("type mismatch", 3, 1, "test.kl", "")

	if !ec.HasErrors() {
		t.Fatal("expected errors")
	}
	if len(ec.Errors()) != 3 {
		t.Errorf("error count = %d, want 3", len(ec.Errors()))
	}
}

func TestErrorCollector_CountByType(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddSyntaxError("err1", 1, 1, "", "")
	ec.AddSyntaxError("err2", 2, 1, "", "")
	ec.AddSemanticError("err3", 3, 1, "", "")
	ec.AddWarning("warn1", 4, 1, "", "")

	counts := ec.CountByType()
	if counts[ErrorSyntax] != 2 {
		t.Errorf("syntax count = %d, want 2", counts[ErrorSyntax])
	}
	if counts[ErrorSemantic] != 1 {
		t.Errorf("semantic count = %d, want 1", counts[ErrorSemantic])
	}
	if counts[ErrorWarning] != 1 {
		t.Errorf("warning count = %d, want 1", counts[ErrorWarning])
	}
}

func TestErrorCollector_GetWarnings(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddWarning("warn1", 1, 1, "", "")
	ec.AddSyntaxError("err1", 2, 1, "", "")

	if !ec.HasWarnings() {
		t.Error("expected warnings")
	}
	warnings := ec.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("warning count = %d, want 1", len(warnings))
	}
}

func TestErrorCollector_GetErrorsByType(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddTypeError("t1", 1, 1, "", "")
	ec.AddTypeError("t2", 2, 1, "", "")
	ec.AddSyntaxError("s1", 3, 1, "", "")

	typeErrors := ec.GetErrorsByType(ErrorTypeError)
	if len(typeErrors) != 2 {
		t.Errorf("type errors = %d, want 2", len(typeErrors))
	}
}

func TestErrorCollector_GetErrorTypes(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddSyntaxError("s", 1, 1, "", "")
	ec.AddTypeError("t", 2, 1, "", "")
	ec.AddSyntaxError("s2", 3, 1, "", "") // duplicate type

	types := ec.GetErrorTypes()
	if len(types) != 2 {
		t.Errorf("unique error types = %d, want 2", len(types))
	}
}

func TestErrorCollector_Clear(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddSyntaxError("err", 1, 1, "", "")
	ec.Clear()

	if ec.HasErrors() {
		t.Error("expected no errors after Clear()")
	}
}

func TestError_String(t *testing.T) {
	e := &Error{
		Type:    ErrorSyntax,
		Message: "unexpected token",
		Line:    10,
		Column:  5,
		File:    "main.kl",
	}
	s := e.String()
	if !strings.Contains(s, "Syntax Error") {
		t.Errorf("error string missing type: %s", s)
	}
	if !strings.Contains(s, "main.kl") {
		t.Errorf("error string missing file: %s", s)
	}
	if !strings.Contains(s, "unexpected token") {
		t.Errorf("error string missing message: %s", s)
	}
}

func TestError_StringWithSuggestion(t *testing.T) {
	e := &Error{
		Type:       ErrorSyntax,
		Message:    "unterminated string",
		Line:       1,
		Column:     1,
		Suggestion: "close the quote",
	}
	s := e.String()
	if !strings.Contains(s, "Suggestion") {
		t.Errorf("error string missing suggestion: %s", s)
	}
}

func TestErrorTypeToString(t *testing.T) {
	tests := []struct {
		typ  ErrorType
		want string
	}{
		{ErrorSyntax, "Syntax"},
		{ErrorSemantic, "Semantic"},
		{ErrorTypeError, "Type"},
		{ErrorRuntime, "Runtime"},
		{ErrorWarning, "Warning"},
		{ErrorType(99), "Unknown"},
	}

	for _, tt := range tests {
		got := ErrorTypeToString(tt.typ)
		if got != tt.want {
			t.Errorf("ErrorTypeToString(%d) = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

func TestFormatErrorPosition(t *testing.T) {
	pos := FormatErrorPosition("main.kl", 10, 5)
	if pos != "main.kl:10:5" {
		t.Errorf("FormatErrorPosition = %q, want \"main.kl:10:5\"", pos)
	}

	posNoFile := FormatErrorPosition("", 10, 5)
	if posNoFile != "10:5" {
		t.Errorf("FormatErrorPosition (no file) = %q, want \"10:5\"", posNoFile)
	}
}

func TestGenerateSuggestion(t *testing.T) {
	s := GenerateSuggestion("unterminated string literal")
	if !strings.Contains(s, "close") && !strings.Contains(s, "quote") {
		t.Errorf("suggestion for unterminated string: %q", s)
	}

	s2 := GenerateSuggestion("completely unknown issue")
	if s2 == "" {
		t.Error("suggestion should not be empty even for unknown messages")
	}
}

func TestExtractSourceContext(t *testing.T) {
	source := "line1\nline2\nline3\nline4\nline5"

	context, sourceLine, lineNumStr := ExtractSourceContext(source, 3, 2)

	if sourceLine != "line3" {
		t.Errorf("sourceLine = %q, want \"line3\"", sourceLine)
	}
	if lineNumStr != "3" {
		t.Errorf("lineNumStr = %q, want \"3\"", lineNumStr)
	}
	if context == "" {
		t.Error("context should not be empty")
	}
	if !strings.Contains(context, ">") {
		t.Error("context should contain '>' marker for the error line")
	}
}

func TestExtractSourceContext_OutOfRange(t *testing.T) {
	context, sourceLine, _ := ExtractSourceContext("hello", 0, 1)
	if context != "" || sourceLine != "" {
		t.Error("out-of-range line should return empty strings")
	}

	context2, sourceLine2, _ := ExtractSourceContext("hello", 100, 1)
	if context2 != "" || sourceLine2 != "" {
		t.Error("out-of-range line should return empty strings")
	}
}

func TestFormatErrorWithContext(t *testing.T) {
	err := &Error{
		Type:       ErrorSemantic,
		Message:    "undefined variable 'x'",
		Line:       5,
		Column:     3,
		File:       "test.kl",
		Suggestion: "declare the variable first",
	}
	formatted := FormatErrorWithContext(err)
	if !strings.Contains(formatted, "test.kl:5:3") {
		t.Errorf("missing position: %s", formatted)
	}
	if !strings.Contains(formatted, "Semantic Error") {
		t.Errorf("missing error type: %s", formatted)
	}
	if !strings.Contains(formatted, "Suggestion") {
		t.Errorf("missing suggestion: %s", formatted)
	}
}

func TestErrorCollector_GetErrorSummary(t *testing.T) {
	ec := NewErrorCollector()
	summary := ec.GetErrorSummary()
	if !strings.Contains(summary, "No errors") {
		t.Errorf("empty summary = %q, should contain 'No errors'", summary)
	}

	ec.AddSyntaxError("err", 1, 1, "", "")
	summary = ec.GetErrorSummary()
	if !strings.Contains(summary, "1 error") {
		t.Errorf("summary with 1 error = %q", summary)
	}
}

func TestErrorCollector_AddErrorInstance(t *testing.T) {
	ec := NewErrorCollector()
	err := &Error{
		Type:    ErrorRuntime,
		Message: "runtime failure",
		Line:    1,
		Column:  1,
	}
	ec.AddErrorInstance(err)
	if len(ec.Errors()) != 1 {
		t.Errorf("expected 1 error, got %d", len(ec.Errors()))
	}
	if ec.Errors()[0].Type != ErrorRuntime {
		t.Errorf("expected runtime error type")
	}
}

func TestErrorCollector_AddSemanticWarning(t *testing.T) {
	ec := NewErrorCollector()
	ec.AddSemanticWarning("unused variable", 1, 1, "", "")
	if !ec.HasWarnings() {
		t.Error("expected warnings after AddSemanticWarning")
	}
}
