package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree creates a fake repo root with a code file and returns its path.
func writeTree(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()
	code := "package main\n\nfunc main() {\n\tprintln(\"beep\")\n}\n"
	dir := filepath.Join(root, "examples", "grove-buzzer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	err := os.WriteFile(filepath.Join(dir, "buzzer.go"), []byte(code), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs", "devices"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

const wantEmbed = `# Buzzer

<!-- code: examples/grove-buzzer/buzzer.go -->
` + "```go" + `
package main

func main() {
	println("beep")
}
` + "```" + `

[view source](../../examples/grove-buzzer/buzzer.go)

More text.
`

func TestProcess(t *testing.T) {
	root := writeTree(t)
	mdPath := filepath.Join(root, "docs", "devices", "buzzer.md")
	tests := []struct {
		name, in string
	}{
		{
			name: "empty fence first authoring",
			in: `# Buzzer

<!-- code: examples/grove-buzzer/buzzer.go -->
` + "```go\n```" + `

More text.
`,
		},
		{
			name: "stale block with old link refreshed",
			in: `# Buzzer

<!-- code: examples/grove-buzzer/buzzer.go -->
` + "```go" + `
old stale content
` + "```" + `

[view source](wrong/path.go)

More text.
`,
		},
		{
			name: "already up to date",
			in:   wantEmbed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := process([]byte(tt.in), mdPath, root)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != wantEmbed {
				t.Errorf("got:\n%s\nwant:\n%s", got, wantEmbed)
			}
			// Idempotency: processing the output again changes nothing.
			again, err := process(got, mdPath, root)
			if err != nil {
				t.Fatal(err)
			}
			if string(again) != string(got) {
				t.Errorf("not idempotent:\n%s", again)
			}
		})
	}
}

func TestProcessErrors(t *testing.T) {
	root := writeTree(t)
	mdPath := filepath.Join(root, "docs", "buzzer.md")
	tests := []struct {
		name, in, wantErr string
	}{
		{
			name:    "missing file",
			in:      "<!-- code: examples/nope.go -->\n```go\n```\n",
			wantErr: "unreadable file",
		},
		{
			name:    "no fence after marker",
			in:      "<!-- code: examples/grove-buzzer/buzzer.go -->\n\ntext\n",
			wantErr: "immediately followed",
		},
		{
			name:    "unterminated fence",
			in:      "<!-- code: examples/grove-buzzer/buzzer.go -->\n```go\nabc\n",
			wantErr: "unterminated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := process([]byte(tt.in), mdPath, root)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got err %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRunCheckAndWrite(t *testing.T) {
	root := writeTree(t)
	mdPath := filepath.Join(root, "docs", "devices", "buzzer.md")
	stalein := "<!-- code: examples/grove-buzzer/buzzer.go -->\n```go\n```\n"
	if err := os.WriteFile(mdPath, []byte(stalein), 0o644); err != nil {
		t.Fatal(err)
	}
	stale, err := run(filepath.Join(root, "docs"), root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 1 {
		t.Fatalf("want 1 stale file, got %v", stale)
	}
	if _, err := run(filepath.Join(root, "docs"), root, true); err != nil {
		t.Fatal(err)
	}
	stale, err = run(filepath.Join(root, "docs"), root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 0 {
		t.Fatalf("want no stale files after write, got %v", stale)
	}
}
