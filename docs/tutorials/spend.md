# 强制消费流 (spend)

`spend` 语句实现强制消费流（Forced Consumption Flow），用于在编译期证明数组或枚举的所有元素/变体已被消费。这是 Kaula 编译器提供的一种静态安全检查机制。

## 数组模式

按索引消费数组元素，编译器在编译期检查每个索引是否都被覆盖：

```kaula
import std.io

fn main() {
    int[3] arr = [10, 20, 30]

    // 编译期证明：arr 长度为 3，所有元素被消费
    spend(arr) {
        call(1) {
            println("element 1: ", arr[1])
        }
        call(2) {
            println("element 2: ", arr[2])
        }
        call(3) {
            println("element 3: ", arr[3])
        }
    }
}
```

### 使用 default 兜底

数组模式支持 `call(default)` 覆盖剩余未消费的元素：

```kaula
spend(arr) {
    call(1) {
        println("first element")
    }
    call(default) {
        println("remaining elements")
    }
}
```

## 枚举模式

按变体名穷尽消费枚举的所有变体，编译器在编译期验证每个变体都被覆盖：

```kaula
import std.io

enum Color {
    Red,
    Green,
    Blue
}

fn describe(Color color) {
    spend(color) {
        call(Red)   { println("red") }
        call(Green) { println("green") }
        call(Blue)  { println("blue") }
    }
}
```

### 未穷尽的错误

如果枚举模式未覆盖所有变体，编译器会报错：

```kaula
enum Color { Red, Green, Blue }

spend(color) {
    call(Red)   { ... }
    call(Green) { ... }
    // 编译错误：未穷尽枚举 'Color'，变体 Blue 未被消费
}
```

### 枚举模式不允许 call(default)

枚举模式必须显式穷尽所有变体，不允许使用 `call(default)` 代替：

```kaula
enum Color { Red, Green, Blue }

spend(color) {
    call(default) { ... }
    // 编译错误：枚举消费模式必须穷尽所有变体，不允许 call(default)
}
```

## 约束

- 强制消费流禁止在 `call` 子句内使用 `return`/`break`/`continue` 提前退出，否则会跳过剩余元素消费
- 带数据的枚举变体暂不支持，需使用 `match` 表达式

## 完整示例

参见 [examples/spend.kl](examples/spend.kl)。