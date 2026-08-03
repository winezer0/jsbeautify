package jsbeautify

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenKind uint8

const (
	tokenWord tokenKind = iota
	tokenKeyword
	tokenNumber
	tokenString
	tokenTemplate
	tokenRegex
	tokenLineComment
	tokenBlockComment
	tokenOperator
	tokenPunctuation
)

type token struct {
	kind tokenKind
	text string
}

var keywords = map[string]struct{}{
	"async": {}, "await": {}, "break": {}, "case": {}, "catch": {}, "class": {},
	"const": {}, "continue": {}, "debugger": {}, "default": {}, "delete": {}, "do": {},
	"else": {}, "export": {}, "extends": {}, "finally": {}, "for": {}, "from": {},
	"function": {}, "if": {}, "import": {}, "in": {}, "instanceof": {}, "let": {},
	"new": {}, "of": {}, "return": {}, "static": {}, "super": {}, "switch": {},
	"throw": {}, "try": {}, "typeof": {}, "var": {}, "void": {}, "while": {},
	"with": {}, "yield": {},
}

var operators = []string{
	">>>=", "===", "!==", "**=", "&&=", "||=", "??=", "...", "=>", "?.",
	">>>", "<<=", ">>=", "==", "!=", "<=", ">=", "++", "--", "&&", "||",
	"??", "**", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<", ">>",
}

// scan tokenizes source while preserving every non-whitespace token verbatim.
func scan(source string) ([]token, error) {
	var tokens []token
	for offset := 0; offset < len(source); {
		if isSpace(source[offset]) {
			offset++
			continue
		}

		start := offset
		if hasPrefix(source, offset, "//") {
			offset = scanLineComment(source, offset)
			tokens = append(tokens, token{tokenLineComment, source[start:offset]})
			continue
		}
		if hasPrefix(source, offset, "/*") {
			end, err := scanBlockComment(source, offset)
			if err != nil {
				return nil, err
			}
			offset = end
			tokens = append(tokens, token{tokenBlockComment, source[start:offset]})
			continue
		}
		if source[offset] == '\'' || source[offset] == '"' {
			end, err := scanQuoted(source, offset, source[offset])
			if err != nil {
				return nil, err
			}
			offset = end
			tokens = append(tokens, token{tokenString, source[start:offset]})
			continue
		}
		if source[offset] == '`' {
			end, err := scanQuoted(source, offset, '`')
			if err != nil {
				return nil, err
			}
			offset = end
			tokens = append(tokens, token{tokenTemplate, source[start:offset]})
			continue
		}
		if source[offset] == '/' && startsRegex(tokens) {
			end, err := scanRegex(source, offset)
			if err != nil {
				return nil, err
			}
			offset = end
			tokens = append(tokens, token{tokenRegex, source[start:offset]})
			continue
		}
		if isNumberStart(source, offset) {
			offset = scanNumber(source, offset)
			tokens = append(tokens, token{tokenNumber, source[start:offset]})
			continue
		}
		if isIdentifierStart(source, offset) || source[offset] == '#' {
			offset = scanIdentifier(source, offset)
			word := source[start:offset]
			kind := tokenWord
			if _, ok := keywords[word]; ok {
				kind = tokenKeyword
			}
			tokens = append(tokens, token{kind, word})
			continue
		}
		if operator := matchingOperator(source, offset); operator != "" {
			offset += len(operator)
			tokens = append(tokens, token{tokenOperator, operator})
			continue
		}
		if strings.ContainsRune("{}()[],:;.@", rune(source[offset])) {
			offset++
			tokens = append(tokens, token{tokenPunctuation, source[start:offset]})
			continue
		}
		offset++
		tokens = append(tokens, token{tokenOperator, source[start:offset]})
	}
	return tokens, nil
}

// isSpace reports whether value is JavaScript whitespace handled by this scanner.
func isSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

// hasPrefix reports whether source has prefix at offset.
func hasPrefix(source string, offset int, prefix string) bool {
	return len(source)-offset >= len(prefix) && source[offset:offset+len(prefix)] == prefix
}

// scanLineComment returns the offset immediately after a line comment.
func scanLineComment(source string, offset int) int {
	for offset < len(source) && source[offset] != '\n' && source[offset] != '\r' {
		offset++
	}
	return offset
}

// scanBlockComment returns the offset immediately after a block comment.
func scanBlockComment(source string, offset int) (int, error) {
	end := strings.Index(source[offset+2:], "*/")
	if end < 0 {
		return 0, fmt.Errorf("unterminated block comment")
	}
	return offset + end + 4, nil
}

// scanQuoted returns the offset immediately after a quoted literal.
func scanQuoted(source string, offset int, quote byte) (int, error) {
	for offset++; offset < len(source); offset++ {
		if source[offset] == '\\' {
			offset++
			continue
		}
		if source[offset] == quote {
			return offset + 1, nil
		}
		if quote != '`' && (source[offset] == '\n' || source[offset] == '\r') {
			return 0, fmt.Errorf("unterminated string literal")
		}
	}
	return 0, fmt.Errorf("unterminated literal")
}

// scanRegex returns the offset immediately after a regular expression literal.
func scanRegex(source string, offset int) (int, error) {
	inClass := false
	for offset++; offset < len(source); offset++ {
		if source[offset] == '\\' {
			offset++
			continue
		}
		if source[offset] == '[' {
			inClass = true
		}
		if source[offset] == ']' {
			inClass = false
		}
		if source[offset] == '/' && !inClass {
			offset++
			for offset < len(source) && isIdentifierPart(source, offset) {
				offset = nextRune(source, offset)
			}
			return offset, nil
		}
		if source[offset] == '\n' || source[offset] == '\r' {
			break
		}
	}
	return 0, fmt.Errorf("unterminated regular expression")
}

// startsRegex uses the preceding token to distinguish a regex from division.
func startsRegex(tokens []token) bool {
	if len(tokens) == 0 {
		return true
	}
	previous := tokens[len(tokens)-1]
	if previous.kind == tokenOperator {
		return previous.text != "++" && previous.text != "--"
	}
	if previous.kind == tokenKeyword {
		switch previous.text {
		case "return", "throw", "case", "delete", "typeof", "void", "yield", "await", "else", "do", "in", "of":
			return true
		}
	}
	switch previous.text {
	case "(", "[", "{", ",", ":", ";":
		return true
	default:
		return false
	}
}

// isNumberStart reports whether an offset begins a numeric literal.
func isNumberStart(source string, offset int) bool {
	if source[offset] >= '0' && source[offset] <= '9' {
		return true
	}
	return source[offset] == '.' && offset+1 < len(source) && source[offset+1] >= '0' && source[offset+1] <= '9'
}

// scanNumber returns the offset immediately after a numeric literal.
func scanNumber(source string, offset int) int {
	for offset < len(source) {
		value := source[offset]
		if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F' || strings.ContainsRune("._xXoObBeEn", rune(value))) {
			break
		}
		offset++
	}
	return offset
}

// isIdentifierStart reports whether an offset begins an identifier.
func isIdentifierStart(source string, offset int) bool {
	value, _ := utf8.DecodeRuneInString(source[offset:])
	return value == '_' || value == '$' || unicode.IsLetter(value)
}

// isIdentifierPart reports whether an offset continues an identifier.
func isIdentifierPart(source string, offset int) bool {
	value, _ := utf8.DecodeRuneInString(source[offset:])
	return value == '_' || value == '$' || unicode.IsLetter(value) || unicode.IsDigit(value)
}

// scanIdentifier returns the offset immediately after an identifier.
func scanIdentifier(source string, offset int) int {
	if source[offset] == '#' {
		offset++
	}
	for offset < len(source) && isIdentifierPart(source, offset) {
		offset = nextRune(source, offset)
	}
	return offset
}

// nextRune advances offset by one UTF-8 rune.
func nextRune(source string, offset int) int {
	_, width := utf8.DecodeRuneInString(source[offset:])
	return offset + width
}

// matchingOperator returns the longest known operator at offset.
func matchingOperator(source string, offset int) string {
	for _, operator := range operators {
		if hasPrefix(source, offset, operator) {
			return operator
		}
	}
	return ""
}
