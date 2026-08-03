package jsbeautify

import (
	"fmt"
	"strings"
)

type frame struct {
	open string
}

type printer struct {
	options  Options
	output   strings.Builder
	stack    []frame
	previous token
	hasPrev  bool
	indent   int
	lineSize int
	pending  bool
	ternary  int
}

// printTokens renders a token stream without changing token text or order.
func printTokens(tokens []token, options Options) (string, error) {
	p := printer{options: options}
	for index, current := range tokens {
		var next token
		hasNext := index+1 < len(tokens)
		if hasNext {
			next = tokens[index+1]
		}
		if err := p.write(current, next, hasNext); err != nil {
			return "", err
		}
	}
	if len(p.stack) != 0 {
		return "", fmt.Errorf("unclosed %q", p.stack[len(p.stack)-1].open)
	}
	result := strings.TrimRight(p.output.String(), " \t\r\n")
	if options.EndWithNewline && result != "" {
		result += "\n"
	}
	return result, nil
}

// write emits one token and updates printer context.
func (p *printer) write(current, next token, hasNext bool) error {
	var err error
	switch current.kind {
	case tokenLineComment:
		p.space()
		p.raw(current.text)
		p.newline()
	case tokenBlockComment:
		p.space()
		p.raw(current.text)
		p.space()
	case tokenPunctuation:
		err = p.writePunctuation(current, next, hasNext)
	case tokenOperator:
		p.writeOperator(current)
	default:
		p.writeValue(current)
	}
	if err != nil {
		return err
	}
	p.previous = current
	p.hasPrev = true
	return nil
}

// writePunctuation applies delimiter-specific spacing and indentation rules.
func (p *printer) writePunctuation(current, next token, hasNext bool) error {
	switch current.text {
	case "{":
		if p.needsSpaceBeforeBrace() {
			p.space()
		}
		p.raw(current.text)
		p.stack = append(p.stack, frame{open: current.text})
		p.indent++
		if !hasNext || next.text != "}" {
			p.newline()
		}
	case "}":
		if err := p.close("{"); err != nil {
			return err
		}
		p.indent--
		if p.lineSize > 0 {
			p.newline()
		}
		p.raw(current.text)
		if hasNext && isBlockFollower(next.text) {
			p.space()
		} else if hasNext && !isTightAfterBlock(next.text) {
			p.newline()
		}
	case "(":
		if p.isControlWord() {
			p.space()
		}
		p.raw(current.text)
		p.stack = append(p.stack, frame{open: current.text})
	case ")":
		if err := p.close("("); err != nil {
			return err
		}
		p.raw(current.text)
	case "[":
		if p.isKeyword("return") || p.isKeyword("throw") || p.isKeyword("yield") {
			p.space()
		}
		p.raw(current.text)
		p.stack = append(p.stack, frame{open: current.text})
	case "]":
		if err := p.close("["); err != nil {
			return err
		}
		p.raw(current.text)
	case ";":
		p.raw(current.text)
		if p.insideParen() {
			p.space()
		} else {
			p.newline()
		}
	case ",":
		p.raw(current.text)
		if p.topIs("{") || p.shouldWrap() {
			p.newline()
		} else {
			p.space()
		}
	case ":":
		if p.ternary > 0 {
			p.ternary--
			p.space()
			p.raw(current.text)
			p.space()
		} else {
			p.raw(current.text)
			p.space()
		}
	case ".", "@":
		if current.text == "." && p.options.BreakChainedMethod && p.shouldWrap() {
			p.newline()
		}
		p.raw(current.text)
	default:
		p.raw(current.text)
	}
	return nil
}

// writeOperator applies unary, binary, and ternary operator spacing.
func (p *printer) writeOperator(current token) {
	switch current.text {
	case "?.", "...":
		if current.text == "?." && p.options.BreakChainedMethod && p.shouldWrap() {
			p.newline()
		}
		p.raw(current.text)
	case "!", "~":
		p.raw(current.text)
	case "++", "--":
		if p.hasPrev && isPostfix(p.previous) {
			p.raw(current.text)
			return
		}
		p.raw(current.text)
	case "?":
		p.ternary++
		p.space()
		p.raw(current.text)
		p.space()
	case "+", "-":
		if p.isUnaryPosition() {
			p.raw(current.text)
			return
		}
		p.binaryOperator(current.text)
	default:
		p.binaryOperator(current.text)
	}
}

// binaryOperator emits an operator with space on both sides.
func (p *printer) binaryOperator(text string) {
	if p.shouldWrap() {
		p.newline()
	} else {
		p.space()
	}
	p.raw(text)
	p.space()
}

// writeValue emits an identifier or literal token.
func (p *printer) writeValue(current token) {
	if p.requiresSpace(current) {
		p.space()
	}
	p.raw(current.text)
}

// requiresSpace reports whether a value must be separated from its predecessor.
func (p *printer) requiresSpace(current token) bool {
	if !p.hasPrev || p.lineSize == 0 {
		return false
	}
	if p.previous.kind == tokenOperator {
		return p.previous.text != "!" && p.previous.text != "~" && p.previous.text != "?." && p.previous.text != "..."
	}
	return isValue(p.previous) && isValue(current)
}

// needsSpaceBeforeBrace reports whether an opening brace follows a block header.
func (p *printer) needsSpaceBeforeBrace() bool {
	if !p.hasPrev || p.lineSize == 0 {
		return false
	}
	switch p.previous.text {
	case "(", "[", ".", "?.", ",", ":", "=", "=>", "?":
		return false
	default:
		return p.previous.kind != tokenOperator
	}
}

// isControlWord reports whether the previous token requires a space before parenthesis.
func (p *printer) isControlWord() bool {
	if !p.hasPrev || p.previous.kind != tokenKeyword {
		return false
	}
	switch p.previous.text {
	case "if", "for", "while", "switch", "catch", "with":
		return true
	default:
		return false
	}
}

// isKeyword reports whether the previous token is word.
func (p *printer) isKeyword(word string) bool {
	return p.hasPrev && p.previous.kind == tokenKeyword && p.previous.text == word
}

// isUnaryPosition reports whether the next plus or minus is unary.
func (p *printer) isUnaryPosition() bool {
	if !p.hasPrev {
		return true
	}
	if p.previous.kind == tokenOperator {
		return true
	}
	switch p.previous.text {
	case "(", "[", "{", ",", ":", ";":
		return true
	default:
		return false
	}
}

// close validates and removes the expected opening delimiter.
func (p *printer) close(want string) error {
	if len(p.stack) == 0 || p.stack[len(p.stack)-1].open != want {
		return fmt.Errorf("unmatched closing delimiter")
	}
	p.stack = p.stack[:len(p.stack)-1]
	return nil
}

// topIs reports whether the active delimiter is open.
func (p *printer) topIs(open string) bool {
	return len(p.stack) > 0 && p.stack[len(p.stack)-1].open == open
}

// insideParen reports whether the current context is inside parentheses.
func (p *printer) insideParen() bool {
	for index := len(p.stack) - 1; index >= 0; index-- {
		if p.stack[index].open == "(" {
			return true
		}
	}
	return false
}

// shouldWrap reports whether the current line reached the configured width.
func (p *printer) shouldWrap() bool {
	return p.options.MaxLineLength > 0 && p.lineSize >= p.options.MaxLineLength
}

// space schedules a space before the next token.
func (p *printer) space() {
	if p.lineSize > 0 {
		p.pending = true
	}
}

// raw writes token text and applies pending whitespace and indentation.
func (p *printer) raw(text string) {
	if p.pending && p.lineSize > 0 {
		p.output.WriteByte(' ')
		p.lineSize++
	}
	p.pending = false
	if p.lineSize == 0 {
		for count := 0; count < p.indent*p.options.IndentSize; count++ {
			p.output.WriteByte(' ')
		}
		p.lineSize = p.indent * p.options.IndentSize
	}
	p.output.WriteString(text)
	if newline := strings.LastIndexByte(text, '\n'); newline >= 0 {
		p.lineSize = len(text) - newline - 1
	} else {
		p.lineSize += len(text)
	}
}

// newline terminates the current non-empty output line.
func (p *printer) newline() {
	p.pending = false
	if p.lineSize == 0 {
		return
	}
	p.output.WriteByte('\n')
	p.lineSize = 0
}

// isValue reports whether a token is a word-like value.
func isValue(value token) bool {
	switch value.kind {
	case tokenWord, tokenKeyword, tokenNumber, tokenString, tokenTemplate, tokenRegex:
		return true
	default:
		return false
	}
}

// isPostfix reports whether an increment or decrement can follow value.
func isPostfix(value token) bool {
	return isValue(value) || value.text == ")" || value.text == "]"
}

// isBlockFollower reports whether text continues a closing block on the same line.
func isBlockFollower(text string) bool {
	return text == "else" || text == "catch" || text == "finally"
}

// isTightAfterBlock reports whether text must follow a closing brace without a newline.
func isTightAfterBlock(text string) bool {
	switch text {
	case ")", "]", ",", ";", ".", "?.", "(", "[", ":", "?":
		return true
	default:
		return false
	}
}
