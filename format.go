// Package jsbeautify formats JavaScript source without building an AST.
package jsbeautify

import "fmt"

// Options controls whitespace emitted by the formatter.
type Options struct {
	IndentSize         int
	MaxLineLength      int
	BreakChainedMethod bool
	EndWithNewline     bool
}

// DefaultOptions returns the default formatter settings.
func DefaultOptions() Options {
	return Options{
		IndentSize:         2,
		MaxLineLength:      120,
		BreakChainedMethod: false,
		EndWithNewline:     true,
	}
}

// Format formats JavaScript source with the default options.
func Format(source string) (string, error) {
	return FormatWithOptions(source, DefaultOptions())
}

// FormatWithOptions formats JavaScript source with caller-provided options.
func FormatWithOptions(source string, options Options) (string, error) {
	if options.IndentSize < 1 || options.IndentSize > 8 {
		return "", fmt.Errorf("indent size must be between 1 and 8")
	}
	if options.MaxLineLength < 0 {
		return "", fmt.Errorf("maximum line length must not be negative")
	}

	tokens, err := scan(source)
	if err != nil {
		return "", err
	}
	return printTokens(tokens, options)
}
