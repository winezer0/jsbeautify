package jsbeautify

import "testing"

func TestFormat(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:  "minified function and object",
			input: "function run(a,b){const item={a:a,b:b};return item;}",
			output: "function run(a, b) {\n" +
				"  const item = {\n" +
				"    a: a,\n" +
				"    b: b\n" +
				"  };\n" +
				"  return item;\n" +
				"}\n",
		},
		{
			name:  "for loop and operators",
			input: "for(let i=0;i<3;i++){total+=i;}",
			output: "for (let i = 0; i < 3; i++) {\n" +
				"  total += i;\n" +
				"}\n",
		},
		{
			name:  "comments strings regex and templates remain intact",
			input: "const value=/a\\/b/g;const text=`value:${value}`;// trailing\ncall(value,text);",
			output: "const value = /a\\/b/g;\n" +
				"const text = `value:${value}`;\n" +
				"// trailing\n" +
				"call(value, text);\n",
		},
		{
			name:  "modern operators",
			input: "const name=user?.profile?.name??'anonymous';const copy={...user};",
			output: "const name = user?.profile?.name ?? 'anonymous';\n" +
				"const copy = {\n" +
				"  ...user\n" +
				"};\n",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Format(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if actual != test.output {
				t.Fatalf("formatted output mismatch\nwant:\n%s\ngot:\n%s", test.output, actual)
			}
		})
	}
}

func TestFormatErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "unterminated string", input: "const text = 'value"},
		{name: "unterminated comment", input: "/* value"},
		{name: "unterminated regular expression", input: "const matcher = /value"},
		{name: "unmatched closing brace", input: "}"},
		{name: "unclosed opening brace", input: "{"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Format(test.input); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestFormatWithOptionsRejectsInvalidOptions(t *testing.T) {
	cases := []struct {
		name    string
		options Options
	}{
		{name: "zero indent", options: Options{IndentSize: 0}},
		{name: "large indent", options: Options{IndentSize: 9}},
		{name: "negative line length", options: Options{IndentSize: 2, MaxLineLength: -1}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := FormatWithOptions("const value = 1;", test.options); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestFormatWithOptions(t *testing.T) {
	options := DefaultOptions()
	options.IndentSize = 4
	options.EndWithNewline = false
	actual, err := FormatWithOptions("if(a){b();}", options)
	if err != nil {
		t.Fatal(err)
	}
	want := "if (a) {\n    b();\n}"
	if actual != want {
		t.Fatalf("want %q, got %q", want, actual)
	}
}
