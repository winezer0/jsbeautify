package jsbeautify

import "testing"

func TestScan(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []token
	}{
		{
			name:  "recognizes regular expressions and division",
			input: "const re=/a[\\/]b/g;const ratio=a/b;",
			want: []token{
				{tokenKeyword, "const"}, {tokenWord, "re"}, {tokenOperator, "="}, {tokenRegex, "/a[\\/]b/g"}, {tokenPunctuation, ";"},
				{tokenKeyword, "const"}, {tokenWord, "ratio"}, {tokenOperator, "="}, {tokenWord, "a"}, {tokenOperator, "/"}, {tokenWord, "b"}, {tokenPunctuation, ";"},
			},
		},
		{
			name:  "recognizes modern operators",
			input: "a?.b??c,...items",
			want: []token{
				{tokenWord, "a"}, {tokenOperator, "?."}, {tokenWord, "b"}, {tokenOperator, "??"}, {tokenWord, "c"}, {tokenPunctuation, ","}, {tokenOperator, "..."}, {tokenWord, "items"},
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			actual, err := scan(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(actual) != len(test.want) {
				t.Fatalf("want %d tokens, got %d", len(test.want), len(actual))
			}
			for index := range test.want {
				if actual[index] != test.want[index] {
					t.Fatalf("token %d: want %#v, got %#v", index, test.want[index], actual[index])
				}
			}
		})
	}
}
