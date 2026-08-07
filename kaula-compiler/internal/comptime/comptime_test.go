package comptime

import (
	"kaula-compiler/internal/ast"
	"testing"
)

// ---- Value basics ----

func TestValue_String(t *testing.T) {
	tests := []struct {
		val  *Value
		want string
	}{
		{&Value{Kind: KindInt, IntVal: 42}, "42"},
		{&Value{Kind: KindFloat, FloatVal: 3.14}, "3.140000"},
		{&Value{Kind: KindBool, BoolVal: true}, "true"},
		{&Value{Kind: KindBool, BoolVal: false}, "false"},
		{&Value{Kind: KindString, StringVal: "hello"}, "hello"},
		{&Value{Kind: KindNull}, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.val.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValue_TypeName(t *testing.T) {
	tests := []struct {
		val  *Value
		want string
	}{
		{&Value{Kind: KindInt}, "i64"},
		{&Value{Kind: KindFloat}, "f64"},
		{&Value{Kind: KindBool}, "bool"},
		{&Value{Kind: KindString}, "string"},
		{&Value{Kind: KindNull}, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.val.TypeName()
			if got != tt.want {
				t.Errorf("TypeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---- integer literals ----

func TestEval_IntegerLiteral(t *testing.T) {
	e := NewEvaluator()
	val, err := e.Eval(&ast.IntegerLiteral{Value: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindInt || val.IntVal != 42 {
		t.Errorf("got %+v, want KindInt(42)", val)
	}
}

// ---- float literals ----

func TestEval_FloatLiteral(t *testing.T) {
	e := NewEvaluator()
	val, err := e.Eval(&ast.FloatLiteral{Value: 3.14})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindFloat || val.FloatVal != 3.14 {
		t.Errorf("got %+v, want KindFloat(3.14)", val)
	}
}

// ---- boolean literals ----

func TestEval_BooleanLiteral(t *testing.T) {
	e := NewEvaluator()
	val, err := e.Eval(&ast.BooleanLiteral{Value: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != true {
		t.Errorf("got %+v, want KindBool(true)", val)
	}
}

// ---- string literals ----

func TestEval_StringLiteral(t *testing.T) {
	e := NewEvaluator()
	val, err := e.Eval(&ast.StringLiteral{Value: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindString || val.StringVal != "hello" {
		t.Errorf("got %+v, want KindString(hello)", val)
	}
}

// ---- nil expression ----

func TestEval_NilExpr(t *testing.T) {
	e := NewEvaluator()
	_, err := e.Eval(nil)
	if err == nil {
		t.Fatal("expected error for nil expression")
	}
}

// ---- identifiers ----

func TestEval_Ident(t *testing.T) {
	e := NewEvaluator()
	e.DefineVar("x", &Value{Kind: KindInt, IntVal: 10})
	val, err := e.Eval(&ast.Identifier{Name: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.IntVal != 10 {
		t.Errorf("got %d, want 10", val.IntVal)
	}
}

func TestEval_Ident_Builtin(t *testing.T) {
	e := NewEvaluator()
	val, err := e.Eval(&ast.Identifier{Name: "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != true {
		t.Errorf("got %+v, want KindBool(true)", val)
	}
}

func TestEval_Ident_Null(t *testing.T) {
	e := NewEvaluator()
	val, err := e.Eval(&ast.Identifier{Name: "null"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindNull {
		t.Errorf("got %+v, want KindNull", val)
	}
}

func TestEval_Ident_Undefined(t *testing.T) {
	e := NewEvaluator()
	_, err := e.Eval(&ast.Identifier{Name: "undefined"})
	if err == nil {
		t.Fatal("expected error for undefined identifier")
	}
}

// ---- binary integer operations ----

func TestEval_BinaryInt(t *testing.T) {
	tests := []struct {
		op    string
		l, r  uint64
		want  uint64
		isInt bool
	}{
		{"+", 10, 20, 30, true},
		{"-", 20, 10, 10, true},
		{"*", 5, 6, 30, true},
		{"/", 30, 5, 6, true},
		{"%", 17, 5, 2, true},
		{"==", 10, 10, 1, false},
		{"!=", 10, 20, 1, false},
		{"<", 10, 20, 1, false},
		{">", 20, 10, 1, false},
		{"<=", 10, 10, 1, false},
		{">=", 10, 10, 1, false},
		{"<<", 1, 4, 16, true},
		{">>", 16, 2, 4, true},
		{"&", 6, 3, 2, true},
		{"|", 6, 3, 7, true},
		{"^", 6, 3, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			e := NewEvaluator()
			expr := &ast.BinaryExpression{
				Left:     &ast.IntegerLiteral{Value: tt.l},
				Operator: tt.op,
				Right:    &ast.IntegerLiteral{Value: tt.r},
			}
			val, err := e.Eval(expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.isInt {
				if val.Kind != KindInt || val.IntVal != tt.want {
					t.Errorf("got %+v, want KindInt(%d)", val, tt.want)
				}
			} else {
				wantBool := tt.want != 0
				if val.Kind != KindBool || val.BoolVal != wantBool {
					t.Errorf("got %+v, want KindBool(%v)", val, wantBool)
				}
			}
		})
	}
}

func TestEval_BinaryInt_DivByZero(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.IntegerLiteral{Value: 10},
		Operator: "/",
		Right:    &ast.IntegerLiteral{Value: 0},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

func TestEval_BinaryInt_ModByZero(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.IntegerLiteral{Value: 10},
		Operator: "%",
		Right:    &ast.IntegerLiteral{Value: 0},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected modulo by zero error")
	}
}

// ---- binary float operations ----

func TestEval_BinaryFloat(t *testing.T) {
	tests := []struct {
		op       string
		l, r     float64
		wantVal  float64
		wantBool bool
		isFloat  bool
	}{
		{"+", 1.5, 2.5, 4.0, false, true},
		{"-", 5.0, 2.0, 3.0, false, true},
		{"*", 3.0, 4.0, 12.0, false, true},
		{"/", 10.0, 2.5, 4.0, false, true},
		{"==", 3.0, 3.0, 0, true, false},
		{"!=", 3.0, 4.0, 0, true, false},
		{"<", 2.0, 3.0, 0, true, false},
		{">", 3.0, 2.0, 0, true, false},
		{"<=", 3.0, 3.0, 0, true, false},
		{">=", 3.0, 3.0, 0, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			e := NewEvaluator()
			expr := &ast.BinaryExpression{
				Left:     &ast.FloatLiteral{Value: tt.l},
				Operator: tt.op,
				Right:    &ast.FloatLiteral{Value: tt.r},
			}
			val, err := e.Eval(expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.isFloat {
				if val.Kind != KindFloat || val.FloatVal != tt.wantVal {
					t.Errorf("got %+v, want KindFloat(%f)", val, tt.wantVal)
				}
			} else {
				if val.Kind != KindBool || val.BoolVal != tt.wantBool {
					t.Errorf("got %+v, want KindBool(%v)", val, tt.wantBool)
				}
			}
		})
	}
}

func TestEval_BinaryFloat_DivByZero(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.FloatLiteral{Value: 1.0},
		Operator: "/",
		Right:    &ast.FloatLiteral{Value: 0},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

// ---- mixed int/float operations ----

func TestEval_BinaryMixed(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.IntegerLiteral{Value: 10},
		Operator: "+",
		Right:    &ast.FloatLiteral{Value: 3.5},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindFloat || val.FloatVal != 13.5 {
		t.Errorf("got %+v, want KindFloat(13.5)", val)
	}
}

// ---- binary bool operations ----

func TestEval_BinaryBool(t *testing.T) {
	tests := []struct {
		op   string
		l, r bool
		want bool
	}{
		{"&&", true, true, true},
		{"&&", true, false, false},
		{"||", false, true, true},
		{"||", false, false, false},
		{"==", true, true, true},
		{"==", true, false, false},
		{"!=", true, false, true},
		{"!=", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			e := NewEvaluator()
			expr := &ast.BinaryExpression{
				Left:     &ast.BooleanLiteral{Value: tt.l},
				Operator: tt.op,
				Right:    &ast.BooleanLiteral{Value: tt.r},
			}
			val, err := e.Eval(expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val.Kind != KindBool || val.BoolVal != tt.want {
				t.Errorf("got %+v, want KindBool(%v)", val, tt.want)
			}
		})
	}
}

// ---- binary string operations ----

func TestEval_BinaryString_Concat(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.StringLiteral{Value: "hello "},
		Operator: "+",
		Right:    &ast.StringLiteral{Value: "world"},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindString || val.StringVal != "hello world" {
		t.Errorf("got %+v, want KindString(hello world)", val)
	}
}

func TestEval_BinaryString_Equal(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.StringLiteral{Value: "abc"},
		Operator: "==",
		Right:    &ast.StringLiteral{Value: "abc"},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != true {
		t.Errorf("got %+v, want KindBool(true)", val)
	}
}

func TestEval_BinaryString_NotEqual(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.StringLiteral{Value: "abc"},
		Operator: "!=",
		Right:    &ast.StringLiteral{Value: "def"},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != true {
		t.Errorf("got %+v, want KindBool(true)", val)
	}
}

// ---- unary operations ----

func TestEval_Unary_NegateInt(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "-",
		Right:    &ast.IntegerLiteral{Value: 42},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := ^uint64(42) + 1 // two's complement of 42
	if val.Kind != KindInt || val.IntVal != want {
		t.Errorf("got %d, want %d", val.IntVal, want)
	}
}

func TestEval_Unary_NegateFloat(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "-",
		Right:    &ast.FloatLiteral{Value: 3.14},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindFloat || val.FloatVal != -3.14 {
		t.Errorf("got %f, want %f", val.FloatVal, -3.14)
	}
}

func TestEval_Unary_Plus(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "+",
		Right:    &ast.IntegerLiteral{Value: 42},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindInt || val.IntVal != 42 {
		t.Errorf("got %d, want 42", val.IntVal)
	}
}

func TestEval_Unary_NotBool(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "!",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != false {
		t.Errorf("got %+v, want KindBool(false)", val)
	}
}

func TestEval_Unary_NotInt(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "!",
		Right:    &ast.IntegerLiteral{Value: 0},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != true {
		t.Errorf("got %+v, want KindBool(true) for !0", val)
	}
}

func TestEval_Unary_BitwiseNot(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "~",
		Right:    &ast.IntegerLiteral{Value: 0xFF},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindInt {
		t.Errorf("got %+v, want KindInt", val)
	}
}

// ---- paren expression ----

func TestEval_Paren(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.ParenExpression{
		Inner: &ast.IntegerLiteral{Value: 42},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.IntVal != 42 {
		t.Errorf("got %d, want 42", val.IntVal)
	}
}

// ---- conditional expression ----

func TestEval_Conditional_True(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.ConditionalExpression{
		Condition: &ast.BooleanLiteral{Value: true},
		TrueExpr:  &ast.IntegerLiteral{Value: 100},
		FalseExpr: &ast.IntegerLiteral{Value: 200},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.IntVal != 100 {
		t.Errorf("got %d, want 100", val.IntVal)
	}
}

func TestEval_Conditional_False(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.ConditionalExpression{
		Condition: &ast.BooleanLiteral{Value: false},
		TrueExpr:  &ast.IntegerLiteral{Value: 100},
		FalseExpr: &ast.IntegerLiteral{Value: 200},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.IntVal != 200 {
		t.Errorf("got %d, want 200", val.IntVal)
	}
}

// ---- sizeof / alignof ----

func TestEval_SizeOf(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.SizeOfExpression{TargetType: "i64"}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindInt || val.IntVal != 8 {
		t.Errorf("got %d, want 8", val.IntVal)
	}
}

func TestEval_SizeOf_Unknown(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.SizeOfExpression{TargetType: "unknown"}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestEval_AlignOf(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.AlignOfExpression{TargetType: "i32"}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindInt || val.IntVal != 4 {
		t.Errorf("got %d, want 4", val.IntVal)
	}
}

// ---- comptime expression (passthrough) ----

func TestEval_ComptimeExpression(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.ComptimeExpression{
		Inner: &ast.IntegerLiteral{Value: 99},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.IntVal != 99 {
		t.Errorf("got %d, want 99", val.IntVal)
	}
}

// ---- unsupported expression ----

func TestEval_Unsupported(t *testing.T) {
	e := NewEvaluator()
	// CharLiteral is an Expression type not handled by the comptime evaluator
	_, err := e.Eval(&ast.CharLiteral{Value: "a"})
	if err == nil {
		t.Fatal("expected error for unsupported expression type")
	}
}

// ---- isTruthy ----

func TestIsTruthy(t *testing.T) {
	e := NewEvaluator()
	tests := []struct {
		val  *Value
		want bool
	}{
		{&Value{Kind: KindBool, BoolVal: true}, true},
		{&Value{Kind: KindBool, BoolVal: false}, false},
		{&Value{Kind: KindInt, IntVal: 1}, true},
		{&Value{Kind: KindInt, IntVal: 0}, false},
		{&Value{Kind: KindFloat, FloatVal: 1.0}, true},
		{&Value{Kind: KindFloat, FloatVal: 0.0}, false},
		{&Value{Kind: KindString, StringVal: "hello"}, true},
		{&Value{Kind: KindString, StringVal: ""}, false},
		{&Value{Kind: KindNull}, false},
	}

	for _, tt := range tests {
		t.Run(tt.val.String(), func(t *testing.T) {
			got := e.isTruthy(tt.val)
			if got != tt.want {
				t.Errorf("isTruthy(%+v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

// ---- toFloat ----

func TestToFloat(t *testing.T) {
	e := NewEvaluator()
	tests := []struct {
		val  *Value
		want float64
	}{
		{&Value{Kind: KindFloat, FloatVal: 3.14}, 3.14},
		{&Value{Kind: KindInt, IntVal: 42}, 42.0},
		{&Value{Kind: KindNull}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.val.String(), func(t *testing.T) {
			got := e.toFloat(tt.val)
			if got != tt.want {
				t.Errorf("toFloat(%+v) = %f, want %f", tt.val, got, tt.want)
			}
		})
	}
}

// ---- toString ----

func TestToString(t *testing.T) {
	e := NewEvaluator()
	tests := []struct {
		val  *Value
		want string
	}{
		{&Value{Kind: KindString, StringVal: "hello"}, "hello"},
		{&Value{Kind: KindInt, IntVal: 42}, "42"},
		{&Value{Kind: KindBool, BoolVal: true}, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := e.toString(tt.val)
			if got != tt.want {
				t.Errorf("toString(%+v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

// ---- DefineVar / GetVar ----

func TestDefineVar_GetVar(t *testing.T) {
	e := NewEvaluator()
	e.DefineVar("x", &Value{Kind: KindInt, IntVal: 100})
	val, ok := e.GetVar("x")
	if !ok {
		t.Fatal("GetVar(x) should return ok=true")
	}
	if val.IntVal != 100 {
		t.Errorf("got %d, want 100", val.IntVal)
	}
}

func TestGetVar_NotFound(t *testing.T) {
	e := NewEvaluator()
	_, ok := e.GetVar("nonexistent")
	if ok {
		t.Fatal("GetVar(nonexistent) should return ok=false")
	}
}

// ---- typeSize / typeAlign ----

func TestTypeSize(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"i8", 1}, {"u8", 1}, {"char", 1}, {"bool", 1},
		{"i16", 2}, {"u16", 2},
		{"i32", 4}, {"u32", 4}, {"int", 4}, {"float", 4}, {"f32", 4},
		{"i64", 8}, {"u64", 8}, {"double", 8}, {"f64", 8},
		{"unknown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeSize(tt.name)
			if got != tt.want {
				t.Errorf("typeSize(%q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestTypeAlign(t *testing.T) {
	if got := typeAlign("i32"); got != 4 {
		t.Errorf("typeAlign(i32) = %d, want 4", got)
	}
}

// ---- ToLiteral ----

func TestToLiteral_Int(t *testing.T) {
	val := &Value{Kind: KindInt, IntVal: 42}
	lit := ToLiteral(val)
	intLit, ok := lit.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected *ast.IntegerLiteral, got %T", lit)
	}
	if intLit.Value != 42 {
		t.Errorf("got %d, want 42", intLit.Value)
	}
}

func TestToLiteral_Float(t *testing.T) {
	val := &Value{Kind: KindFloat, FloatVal: 3.14}
	lit := ToLiteral(val)
	floatLit, ok := lit.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("expected *ast.FloatLiteral, got %T", lit)
	}
	if floatLit.Value != 3.14 {
		t.Errorf("got %f, want 3.14", floatLit.Value)
	}
}

func TestToLiteral_Null(t *testing.T) {
	val := &Value{Kind: KindNull}
	lit := ToLiteral(val)
	ident, ok := lit.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected *ast.Identifier, got %T", lit)
	}
	if ident.Name != "null" {
		t.Errorf("got Name=%q, want %q", ident.Name, "null")
	}
}

// ---- logical AND/OR on integers ----

func TestEval_BinaryInt_LogicalAnd(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.IntegerLiteral{Value: 1},
		Operator: "&&",
		Right:    &ast.IntegerLiteral{Value: 0},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != false {
		t.Errorf("got %+v, want KindBool(false)", val)
	}
}

func TestEval_BinaryInt_LogicalOr(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.IntegerLiteral{Value: 0},
		Operator: "||",
		Right:    &ast.IntegerLiteral{Value: 42},
	}
	val, err := e.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Kind != KindBool || val.BoolVal != true {
		t.Errorf("got %+v, want KindBool(true)", val)
	}
}

// ---- unsupported binary operation ----

func TestEval_Binary_Unsupported(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.IntegerLiteral{Value: 1},
		Operator: "**",
		Right:    &ast.IntegerLiteral{Value: 2},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected error for unsupported operator")
	}
}

// ---- unsupported bool operator ----

func TestEval_BinaryBool_Unsupported(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "+",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected error for unsupported bool operator")
	}
}

// ---- unsupported float operator ----

func TestEval_BinaryFloat_Unsupported(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.FloatLiteral{Value: 1.0},
		Operator: "%",
		Right:    &ast.FloatLiteral{Value: 2.0},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected error for unsupported float operator")
	}
}

// ---- unary unsupported ----

func TestEval_Unary_UnsupportedOp(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.UnaryExpression{
		Operator: "?",
		Right:    &ast.IntegerLiteral{Value: 1},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected error for unsupported unary operator")
	}
}

// ---- binary string unsupported operator ----

func TestEval_BinaryString_Unsupported(t *testing.T) {
	e := NewEvaluator()
	expr := &ast.BinaryExpression{
		Left:     &ast.StringLiteral{Value: "a"},
		Operator: "-",
		Right:    &ast.StringLiteral{Value: "b"},
	}
	_, err := e.Eval(expr)
	if err == nil {
		t.Fatal("expected error for unsupported string operator")
	}
}