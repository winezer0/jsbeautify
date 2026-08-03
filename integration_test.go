package jsbeautify

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestTestdataSyntaxBeforeAndAfterFormat verifies real minified assets before and after formatting.
func TestTestdataSyntaxBeforeAndAfterFormat(t *testing.T) {
	files := []string{
		"jquery@2.2.3.min.js",
		"jquery@3.5.1.min.js",
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			inputPath := filepath.Join("testdata", file)
			runNodeCheck(t, inputPath)

			source, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatal(err)
			}
			formatted, err := Format(string(source))
			if err != nil {
				t.Fatal(err)
			}

			outputPath := filepath.Join(t.TempDir(), file)
			if err := os.WriteFile(outputPath, []byte(formatted), 0644); err != nil {
				t.Fatal(err)
			}
			runNodeCheck(t, outputPath)
		})
	}
}

// runNodeCheck fails the current test when Node.js cannot parse a JavaScript file.
func runNodeCheck(t *testing.T, path string) {
	t.Helper()
	output, err := exec.Command("node", "--check", path).CombinedOutput()
	if err != nil {
		t.Fatalf("node --check %s: %v\n%s", path, err, output)
	}
}
