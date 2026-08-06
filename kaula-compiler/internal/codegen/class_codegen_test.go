package codegen

import (
	"kaula-compiler/internal/ast"
	"strings"
	"testing"
)

// newTestCodegen 创建测试用的 TypeGenerator 和 CodeGenerator
func newTestCodegen() (*TypeGenerator, *CodeGenerator) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}
	cg := &CodeGenerator{
		typeGenerator:       tg,
		functionGenerator:   NewFunctionGenerator(nil),
		expressionGenerator: NewExpressionGenerator(nil),
		statementGenerator:  NewStatementGenerator(nil),
		pluginManager:       NewPluginManager(),
		indent:              0,
	}
	tg.codegen = cg
	cg.functionGenerator.codegen = cg
	cg.expressionGenerator.codegen = cg
	cg.statementGenerator.codegen = cg
	return tg, cg
}

// TestGenerateClassStatement_Basic 验证基本类声明的代码生成
func TestGenerateClassStatement_Basic(t *testing.T) {
	_, codegen := newTestCodegen()

	classStmt := &ast.ClassStatement{
		Name: "Person",
		Fields: []*ast.FieldDeclaration{
			{Name: "name", Type: "string"},
			{Name: "age", Type: "int"},
		},
		Methods: []*ast.MethodStatement{
			{
				Name:       "greet",
				Params:     []*ast.Param{{Name: "greeting", Type: "string"}},
				ReturnType: "string",
				Body:       []ast.Statement{},
			},
		},
		Constructors: []*ast.ConstructorStatement{
			{
				Params: []*ast.Param{{Name: "name", Type: "string"}, {Name: "age", Type: "int"}},
				Body:   []ast.Statement{},
			},
		},
		Implements: []string{},
		Generic:    false,
	}

	code := codegen.typeGenerator.GenerateClassStatement(classStmt)

	// 验证结构体定义
	if !strings.Contains(code, "typedef struct K_Person") {
		t.Error("生成的代码应包含结构体定义 typedef struct K_Person")
	}
	// 验证字段
	if !strings.Contains(code, "String name;") {
		t.Error("生成的代码应包含字段 name (String)")
	}
	if !strings.Contains(code, "int64_t age;") {
		t.Error("生成的代码应包含字段 age (int64_t)")
	}
	// 验证方法声明
	if !strings.Contains(code, "Person_greet") {
		t.Error("生成的代码应包含方法声明 Person_greet")
	}
	// 验证构造函数
	if !strings.Contains(code, "Person_new") {
		t.Error("生成的代码应包含构造函数 Person_new")
	}
	// 验证 alloc
	if !strings.Contains(code, "KMM_V4_ALLOC_ZERO") {
		t.Error("构造函数应包含内存分配 KMM_V4_ALLOC_ZERO")
	}
}

// TestGenerateClassStatement_WithInterface 验证实现了接口的类代码生成
func TestGenerateClassStatement_WithInterface(t *testing.T) {
	_, codegen := newTestCodegen()

	// 注册接口方法，供 buildInterfaceMethodCast 查找
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InterfaceStatement{
				Name: "Speaker",
				Methods: []*ast.MethodStatement{
					{Name: "speak", ReturnType: "string"},
				},
			},
		},
	}
	codegen.program = program

	classStmt := &ast.ClassStatement{
		Name: "Dog",
		Fields: []*ast.FieldDeclaration{
			{Name: "name", Type: "string"},
		},
		Methods: []*ast.MethodStatement{
			{
				Name:       "speak",
				ReturnType: "string",
				Body:       []ast.Statement{},
			},
		},
		Constructors: []*ast.ConstructorStatement{
			{
				Params: []*ast.Param{{Name: "name", Type: "string"}},
				Body:   []ast.Statement{},
			},
		},
		Implements: []string{"Speaker"},
		Generic:    false,
	}

	code := codegen.typeGenerator.GenerateClassStatement(classStmt)

	// 验证接口方法组嵌入
	if !strings.Contains(code, "K_Speaker_MethodGroup") {
		t.Error("实现了接口的类应包含接口方法组类型")
	}
	// 验证接口方法赋值
	if !strings.Contains(code, "self->Speaker.speak = ") {
		t.Error("构造函数应包含接口方法赋值 self->Speaker.speak")
	}
}

// TestGenerateConstructorStatement 验证构造函数代码生成
func TestGenerateConstructorStatement(t *testing.T) {
	_, codegen := newTestCodegen()

	constructor := &ast.ConstructorStatement{
		Params: []*ast.Param{{Name: "name", Type: "string"}},
		Body:   []ast.Statement{},
	}

	code := codegen.typeGenerator.GenerateConstructorStatementWithInterfaceInit("Cat", nil, nil, constructor)

	// 验证构造函数签名
	expectedSignature := "K_Cat* Cat_new("
	if !strings.Contains(code, expectedSignature) {
		t.Errorf("构造函数签名应包含 %q", expectedSignature)
	}
	// 验证参数类型转换
	if !strings.Contains(code, "String name") {
		t.Error("构造函数参数应包含 String name")
	}
	// 验证内存分配
	if !strings.Contains(code, "KMM_V4_ALLOC_ZERO") {
		t.Error("构造函数应包含 KMM_V4_ALLOC_ZERO 内存分配")
	}
	// 验证空指针检查
	if !strings.Contains(code, "if (self == NULL) { return NULL; }") {
		t.Error("构造函数应包含空指针检查")
	}
	// 验证返回值
	if !strings.Contains(code, "return self;") {
		t.Error("构造函数应包含 return self")
	}
}

// TestGenerateMethodStatement 验证方法代码生成
func TestGenerateMethodStatement(t *testing.T) {
	tg, _ := newTestCodegen()

	method := &ast.MethodStatement{
		Name:       "bark",
		ReturnType: "string",
		Params:     []*ast.Param{},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.StringLiteral{Value: "Woof!"},
			},
		},
	}

	code := tg.GenerateMethodStatement("Dog", method)

	// 验证方法签名（string 类型映射为 C 的 String）
	if !strings.Contains(code, "static inline String Dog_bark(K_Dog* self") {
		t.Error("方法签名应包含 'static inline String Dog_bark(K_Dog* self'")
	}
	// 验证方法名为 static inline
	if !strings.HasPrefix(code, "static inline") {
		t.Error("方法应为 static inline 函数")
	}
}

// TestGenerateMethodStatement_VoidReturn 验证无返回值方法生成
func TestGenerateMethodStatement_VoidReturn(t *testing.T) {
	tg, _ := newTestCodegen()

	method := &ast.MethodStatement{
		Name:       "setName",
		ReturnType: "",
		Params:     []*ast.Param{{Name: "newName", Type: "string"}},
		Body:       []ast.Statement{},
	}

	code := tg.GenerateMethodStatement("Dog", method)

	// 无返回类型的方法应生成 void 返回
	if !strings.Contains(code, "static inline void Dog_setName") {
		t.Error("无显式返回类型的方法应生成 void 返回类型")
	}
}

// TestGenerateMethodStatement_ImplicitReturn 验证非 void 方法缺少 return 时生成默认返回
func TestGenerateMethodStatement_ImplicitReturn(t *testing.T) {
	tg, _ := newTestCodegen()

	method := &ast.MethodStatement{
		Name:       "getName",
		ReturnType: "string",
		Params:     []*ast.Param{},
		Body:       []ast.Statement{}, // 空方法体，无 return
	}

	code := tg.GenerateMethodStatement("Dog", method)

	// 非 void 方法缺少 return 时应生成默认 return NULL
	if !strings.Contains(code, "return NULL;") {
		t.Error("非 void 方法缺少 return 语句时应生成默认 'return NULL;'")
	}
}

// TestGenerateInterfaceStatement 验证接口代码生成
func TestGenerateInterfaceStatement(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	ifaceStmt := &ast.InterfaceStatement{
		Name: "Speaker",
		Methods: []*ast.MethodStatement{
			{
				Name:       "speak",
				ReturnType: "string",
				Params:     []*ast.Param{{Name: "msg", Type: "string"}},
			},
			{
				Name:       "getVolume",
				ReturnType: "int",
				Params:     []*ast.Param{},
			},
		},
	}

	code := tg.GenerateInterfaceStatement(ifaceStmt)

	// 验证方法组结构体定义
	if !strings.Contains(code, "typedef struct K_Speaker_MethodGroup") {
		t.Error("接口应生成方法组结构体 typedef struct K_Speaker_MethodGroup")
	}
	// 验证方法签名
	if !strings.Contains(code, "(*speak)") && !strings.Contains(code, "speak (*)") {
		t.Error("接口方法组应包含 speak 函数指针")
	}
	if !strings.Contains(code, "(*getVolume)") && !strings.Contains(code, "getVolume (*)") {
		t.Error("接口方法组应包含 getVolume 函数指针")
	}
	// 验证方法参数含 void* self
	if !strings.Contains(code, "void* self") {
		t.Error("接口方法组的函数指针应包含 void* self 参数")
	}
}

// TestBuildInterfaceMethodCast 验证接口方法强转代码生成
func TestBuildInterfaceMethodCast(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	method := &ast.MethodStatement{
		Name:       "speak",
		ReturnType: "string",
		Params:     []*ast.Param{{Name: "msg", Type: "string"}},
	}

	code := tg.buildInterfaceMethodCast("Dog", "Speaker", method)

	// 验证方法赋值格式（注意末尾有换行）
	expected := "self->Speaker.speak = (String(*)(void*, String))Dog_speak;\n"
	if code != expected {
		t.Errorf("接口方法强转生成错误:\n  期望: %q\n  实际: %q", expected, code)
	}
}

// TestBuildInterfaceMethodCast_NoParams 验证无参数方法强转代码生成
func TestBuildInterfaceMethodCast_NoParams(t *testing.T) {
	tg := &TypeGenerator{
		clibTypeMap:   make(map[string]string),
		structTypes:   make(map[string]bool),
		activeTypeMap: nil,
	}

	method := &ast.MethodStatement{
		Name:       "getValue",
		ReturnType: "int",
		Params:     []*ast.Param{},
	}

	code := tg.buildInterfaceMethodCast("Box", "Container", method)

	// 验证无参数方法强转（注意末尾有换行）
	expected := "self->Container.getValue = (int64_t(*)(void*))Box_getValue;\n"
	if code != expected {
		t.Errorf("无参数方法接口强转生成错误:\n  期望: %q\n  实际: %q", expected, code)
	}
}

// TestGenerateClassStatement_Generic 验证泛型类代码生成
func TestGenerateClassStatement_Generic(t *testing.T) {
	_, codegen := newTestCodegen()

	classStmt := &ast.ClassStatement{
		Name: "Box",
		TypeParams: []*ast.TypeParameter{
			{Name: "T"},
		},
		Fields:       []*ast.FieldDeclaration{},
		Methods:      []*ast.MethodStatement{},
		Constructors: []*ast.ConstructorStatement{},
		Generic:      true,
	}

	code := codegen.typeGenerator.GenerateClassStatement(classStmt)

	// 泛型类应生成占位注释，而非实际结构体定义
	if !strings.Contains(code, "Generic Class: Box") {
		t.Error("泛型类应生成注释 'Generic Class: Box'")
	}
	if !strings.Contains(code, "<T>") {
		t.Error("泛型类应包含类型参数 <T>")
	}
	if strings.Contains(code, "typedef struct") {
		t.Error("泛型类不应生成实际结构体定义")
	}
}

// TestGenerateClassStatement_EmptyClass 验证空类代码生成
func TestGenerateClassStatement_EmptyClass(t *testing.T) {
	_, codegen := newTestCodegen()

	classStmt := &ast.ClassStatement{
		Name:         "Empty",
		Fields:       []*ast.FieldDeclaration{},
		Methods:      []*ast.MethodStatement{},
		Constructors: []*ast.ConstructorStatement{},
		Generic:      false,
	}

	code := codegen.typeGenerator.GenerateClassStatement(classStmt)

	// 空类应有结构体定义
	if !strings.Contains(code, "typedef struct K_Empty") {
		t.Error("空类应生成结构体定义 typedef struct K_Empty")
	}
	// 空类不应有构造函数
	if strings.Contains(code, "Empty_new") {
		t.Error("无构造函数的空类不应生成构造函数")
	}
}

// TestGenerateClassStatement_MultipleInterface 验证实现多个接口的类代码生成
func TestGenerateClassStatement_MultipleInterface(t *testing.T) {
	_, codegen := newTestCodegen()

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InterfaceStatement{
				Name: "Reader",
				Methods: []*ast.MethodStatement{
					{Name: "read", ReturnType: "string"},
				},
			},
			&ast.InterfaceStatement{
				Name: "Writer",
				Methods: []*ast.MethodStatement{
					{Name: "write", ReturnType: "void"},
				},
			},
		},
	}
	codegen.program = program

	classStmt := &ast.ClassStatement{
		Name: "FileIO",
		Fields: []*ast.FieldDeclaration{
			{Name: "path", Type: "string"},
		},
		Methods: []*ast.MethodStatement{
			{Name: "read", ReturnType: "string", Body: []ast.Statement{}},
			{Name: "write", ReturnType: "", Body: []ast.Statement{}},
		},
		Constructors: []*ast.ConstructorStatement{
			{Params: []*ast.Param{{Name: "path", Type: "string"}}, Body: []ast.Statement{}},
		},
		Implements: []string{"Reader", "Writer"},
		Generic:    false,
	}

	code := codegen.typeGenerator.GenerateClassStatement(classStmt)

	// 验证两个接口的方法组都嵌入
	if !strings.Contains(code, "K_Reader_MethodGroup") {
		t.Error("实现多个接口的类应包含 Reader 方法组")
	}
	if !strings.Contains(code, "K_Writer_MethodGroup") {
		t.Error("实现多个接口的类应包含 Writer 方法组")
	}
	// 验证两个接口的方法赋值
	if !strings.Contains(code, "self->Reader.read") {
		t.Error("构造函数应包含 Reader.read 方法赋值")
	}
	if !strings.Contains(code, "self->Writer.write") {
		t.Error("构造函数应包含 Writer.write 方法赋值")
	}
}

// TestGenerateClassStatement_ConstructorWithInterfaceInit 验证带接口初始化的构造函数
func TestGenerateClassStatement_ConstructorWithInterfaceInit(t *testing.T) {
	_, codegen := newTestCodegen()

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InterfaceStatement{
				Name: "Speaker",
				Methods: []*ast.MethodStatement{
					{Name: "speak", ReturnType: "string"},
				},
			},
		},
	}
	codegen.program = program

	constructor := &ast.ConstructorStatement{
		Params: []*ast.Param{},
		Body:   []ast.Statement{},
	}

	code := codegen.typeGenerator.GenerateConstructorStatementWithInterfaceInit("Robot", []string{"Speaker"}, []*ast.MethodStatement{
		{Name: "speak", ReturnType: "string"},
	}, constructor)

	// 验证接口方法组初始化
	if !strings.Contains(code, "Initialize interface method groups") {
		t.Error("构造函数应包含接口方法组初始化")
	}
	// 验证方法赋值
	if !strings.Contains(code, "self->Speaker.speak") {
		t.Error("构造函数应包含 self->Speaker.speak 方法赋值")
	}
}