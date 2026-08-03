package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/winezer0/jsbeautify"
)

// main runs the command-line formatter.
func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		exit(err)
	}
}

// run parses command-line arguments and writes formatted JavaScript.
func run(arguments []string, stdin io.Reader, stdout io.Writer) error {
	options := jsbeautify.DefaultOptions()
	flags := flag.NewFlagSet("jsbeautify", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	output := flags.String("o", "", "write formatted source to file")
	flags.IntVar(&options.IndentSize, "indent", options.IndentSize, "spaces per indentation level")
	flags.IntVar(&options.MaxLineLength, "width", options.MaxLineLength, "maximum preferred line length, 0 disables wrapping")
	flags.BoolVar(&options.BreakChainedMethod, "break-chains", false, "wrap long method chains")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	source, err := readInput(flags.Args(), stdin)
	if err != nil {
		return err
	}
	formatted, err := jsbeautify.FormatWithOptions(source, options)
	if err != nil {
		return err
	}
	if *output == "" {
		_, err = fmt.Fprint(stdout, formatted)
	} else {
		err = os.WriteFile(*output, []byte(formatted), 0644)
	}
	return err
}

// readInput reads one file argument or standard input.
func readInput(arguments []string, stdin io.Reader) (string, error) {
	if len(arguments) > 1 {
		return "", fmt.Errorf("only one input file is supported")
	}
	if len(arguments) == 1 {
		data, err := os.ReadFile(arguments[0])
		return string(data), err
	}
	data, err := io.ReadAll(stdin)
	return string(data), err
}

// exit prints an error and terminates the process.
func exit(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
