# jsbeautify

[Chinese documentation](README.zh-CN.md)

A pure Go JavaScript readability formatter. It expands long minified lines into consistently indented source without building an AST.

## Features

- Formats without Node.js or third-party Go dependencies.
- Preserves strings, template literals, regular expressions, comments, and token order.
- Supports configurable indentation, preferred line width, and chained method wrapping.
- Detects unterminated literals and mismatched delimiters.

## Install

Requires Go 1.22 or later.

```bash
go install github.com/winezer0/jsbeautify/cmd/jsbeautify@latest
```

Run directly from the repository:

```bash
go run ./cmd/jsbeautify input.min.js
```

## CLI Usage

Format a file and write to standard output:

```bash
jsbeautify input.min.js
```

Read from standard input:

```bash
cat input.min.js | jsbeautify
```

Write the result to a file:

```bash
jsbeautify -o output.js input.min.js
```

Configure indentation and preferred line width:

```bash
jsbeautify -indent 4 -width 100 -o output.js input.min.js
```

| Flag | Default | Description |
| --- | --- | --- |
| `-o <path>` | stdout | Write formatted source to a file. |
| `-indent <1-8>` | `2` | Spaces per indentation level. |
| `-width <n>` | `120` | Preferred line width; `0` disables soft wrapping. |
| `-break-chains` | `false` | Wrap long method chains before `.` or `?.`. |

## Go Library

```go
formatted, err := jsbeautify.Format("function run(a,b){return a+b;}")
```

Use custom options when needed:

```go
options := jsbeautify.DefaultOptions()
options.IndentSize = 4
options.MaxLineLength = 100
options.BreakChainedMethod = true

formatted, err := jsbeautify.FormatWithOptions(source, options)
```

## How It Works

```text
source -> tokens -> delimiter/context stack -> formatted source
```

The scanner preserves literals, regular expressions, and comments as opaque tokens. The printer uses delimiter context, semicolons, commas, and operators to add whitespace and line breaks.

## Validation

```bash
go test -cover ./...
go vet ./...
```

Integration tests run `node --check` on the real minified jQuery assets in `testdata`, both before and after formatting, to detect truncation and syntax damage.

## Limitations

This tool is designed for readability recovery, not complete JavaScript parsing or semantic verification. Template literal interpolation is kept unchanged. `node --check` validates syntax only; it does not validate browser DOM behavior or application-level equivalence.
