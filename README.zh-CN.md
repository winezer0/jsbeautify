# jsbeautify

[English documentation](README.md)

一个纯 Go 的 JavaScript 可读性格式化工具，用于将压缩后的长行 JavaScript 拆分为缩进清晰的多行代码，不构建 AST。

## 特性

- 格式化不依赖 Node.js 或第三方 Go 包。
- 保留字符串、模板字符串、正则表达式、注释与 token 顺序。
- 支持配置缩进、首选行宽和链式调用换行。
- 检测未闭合字面量与不匹配的括号。

## 安装

需要 Go 1.22 或更高版本。

```bash
go install github.com/winezer0/jsbeautify/cmd/jsbeautify@latest
```

在仓库目录中直接运行：

```bash
go run ./cmd/jsbeautify input.min.js
```

## 命令行用法

格式化文件并输出到标准输出：

```bash
jsbeautify input.min.js
```

从标准输入读取：

```bash
cat input.min.js | jsbeautify
```

将结果写入文件：

```bash
jsbeautify -o output.js input.min.js
```

自定义缩进和首选行宽：

```bash
jsbeautify -indent 4 -width 100 -o output.js input.min.js
```

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-o <path>` | stdout | 将格式化结果写入指定文件。 |
| `-indent <1-8>` | `2` | 每级缩进的空格数。 |
| `-width <n>` | `120` | 首选最大行宽；`0` 禁用软换行。 |
| `-break-chains` | `false` | 当链式调用超宽时，在 `.` 或 `?.` 前换行。 |

## Go 库用法

```go
formatted, err := jsbeautify.Format("function run(a,b){return a+b;}")
```

需要自定义行为时：

```go
options := jsbeautify.DefaultOptions()
options.IndentSize = 4
options.MaxLineLength = 100
options.BreakChainedMethod = true

formatted, err := jsbeautify.FormatWithOptions(source, options)
```

## 工作方式

```text
source -> tokens -> delimiter/context stack -> formatted source
```

扫描器将字面量、正则和注释作为原样 token 保留。Printer 根据分隔符上下文、分号、逗号和运算符插入空格、缩进和换行。

## 验证

```bash
go test -cover ./...
go vet ./...
```

集成测试会对 `testdata` 中的真实压缩 jQuery 文件执行格式化前后的 `node --check`，用于检测输出截断和 JavaScript 语法损坏。

## 限制

本工具用于恢复压缩 JavaScript 的可读性，不是完整 JavaScript parser 或语义验证器。模板字符串插值保持原样。`node --check` 只验证语法可解析，不验证浏览器 DOM 行为或业务逻辑等价性。
