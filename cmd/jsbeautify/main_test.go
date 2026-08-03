package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name      string
		arguments []string
		input     string
		want      string
		wantErr   bool
	}{
		{
			name:  "formats standard input",
			input: "function x(){return 1;}",
			want:  "function x() {\n  return 1;\n}\n",
		},
		{
			name:      "rejects multiple files",
			arguments: []string{"one.js", "two.js"},
			wantErr:   true,
		},
		{
			name:      "rejects unknown flags",
			arguments: []string{"--unknown"},
			wantErr:   true,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := run(test.arguments, strings.NewReader(test.input), &output)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if output.String() != test.want {
				t.Fatalf("want %q, got %q", test.want, output.String())
			}
		})
	}
}

func TestRunWritesOutputFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "formatted.js")
	if err := run([]string{"-o", path}, strings.NewReader("const a={b:1};"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "const a = {\n  b: 1\n};\n"
	if string(actual) != want {
		t.Fatalf("want %q, got %q", want, actual)
	}
}

func TestReadInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.js")
	if err := os.WriteFile(path, []byte("const value=1;"), 0644); err != nil {
		t.Fatal(err)
	}
	actual, err := readInput([]string{path}, strings.NewReader("unused"))
	if err != nil {
		t.Fatal(err)
	}
	if actual != "const value=1;" {
		t.Fatalf("want file input, got %q", actual)
	}
	if _, err := readInput([]string{"missing.js"}, strings.NewReader("unused")); err == nil {
		t.Fatal("expected a missing-file error")
	}
}

func TestRunReportsOutputError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "formatted.js")
	err := run([]string{"-o", path}, strings.NewReader("const value=1;"), &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected an output error")
	}
}
