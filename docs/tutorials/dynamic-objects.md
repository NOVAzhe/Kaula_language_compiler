# 动态对象字面量

Kaula 提供动态对象字面量，用于在语言层面模拟 JS/Python 风格的动态对象。字段可读可写可动态增删，方法可存为函数字段。

## 创建动态对象

使用 `object { field: value, ... }` 语法创建动态对象：

```kaula
import std.io

fn main() {
    // 创建动态对象
    auto person = object {
        name: "Alice",
        age: 30,
        greet: func() { println("Hello!") }
    }
}
```

空对象使用 `object()` 创建：

```kaula
auto empty = object()
```

## 访问字段

通过 `.` 运算符访问动态对象字段：

```kaula
fn main() {
    auto obj = object { x: 10, y: 20 }

    println("x = ", obj.x)     // 输出: x = 10
    println("y = ", obj.y)     // 输出: y = 20
}
```

## 修改字段

动态对象字段可写，直接赋值即可：

```kaula
fn main() {
    auto obj = object { count: 0 }
    println("before: ", obj.count)  // 输出: before: 0

    obj.count = 42
    println("after: ", obj.count)   // 输出: after: 42
}
```

## 动态增删字段

```kaula
fn main() {
    auto obj = object { }
    obj.name = "dynamic"  // 新增字段
    println(obj.name)
}
```

## 方法字段

函数可以作为字段值存储在动态对象中：

```kaula
fn main() {
    auto calc = object {
        add: func(int a, int b) int { return a + b },
        mul: func(int a, int b) int { return a * b }
    }

    int sum = calc.add(3, 4)     // 7
    int prod = calc.mul(3, 4)    // 12
    println("3 + 4 = ", sum)
    println("3 * 4 = ", prod)
}
```

## 字段重名检查

编译器会检测动态对象中的字段重名并报错：

```kaula
// 编译错误：字段 x 重复定义
auto obj = object {
    x: 1,
    x: 2
}
```

## 完整示例

参见 [examples/dynamic_object.kl](examples/dynamic_object.kl)。