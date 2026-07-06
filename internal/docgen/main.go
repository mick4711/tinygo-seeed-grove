// Command docgen splices source files into markdown documentation.
//
// Markdown files contain marker comments followed by a fenced code block:
//
//	<!-- code: examples/examples-grove-single/grove-buzzer/grove-buzzer.go -->
//	```go
//	```
//
// docgen replaces the contents of the fenced block with the referenced
// file (path relative to the repository root) and maintains a
// "[view source](...)" link after the block with a path relative to the
// markdown file's directory. The operation is idempotent: markers remain in
// place and rerunning docgen on up-to-date files produces no changes.
//
// Usage:
//
//	go run ./internal/docgen -w ./docs    # rewrite files in place
//	go run ./internal/docgen ./docs       # check mode: exit 1 if stale
//
// Future work: named snippet regions (docgen:start/docgen:end) for embedding
// parts of a file. Not needed while examples are small single files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	write := flag.Bool("w", false, "rewrite markdown files in place instead of checking")
	flag.Parse()
	dir := flag.Arg(0)
	if dir == "" {
		dir = "./docs"
	}
	stale, err := run(dir, ".", *write)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docgen:", err)
		os.Exit(1)
	}
	if !*write && len(stale) > 0 {
		fmt.Fprintln(os.Stderr, "docgen: stale files (run `go generate ./...` to fix):")
		for _, f := range stale {
			fmt.Fprintln(os.Stderr, "  "+f)
		}
		os.Exit(1)
	}
}

// run processes all markdown files under dir. Referenced code paths are
// resolved relative to root. In write mode changed files are rewritten;
// otherwise the paths of stale files are returned.
func run(dir, root string, write bool) (stale []string, err error) {
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out, err := process(src, p, root)
		if err != nil {
			return err
		}
		if bytes.Equal(src, out) {
			return nil
		}
		if write {
			return os.WriteFile(p, out, 0o644)
		}
		stale = append(stale, p)
		return nil
	})
	return stale, err
}

const (
	markerPrefix = "<!-- code: "
	markerSuffix = " -->"
	linkPrefix   = "[view source]("
)

// process returns mdSrc with every marker's fenced block refreshed from the
// referenced file. mdPath is the markdown file's path (used for error
// messages and for computing relative source links); root is the directory
// marker paths are relative to.
func process(mdSrc []byte, mdPath, root string) ([]byte, error) {
	lines := strings.SplitAfter(string(mdSrc), "\n")
	var out strings.Builder
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		out.WriteString(line)
		trimmed := strings.TrimRight(line, "\n\r \t")
		if !strings.HasPrefix(trimmed, markerPrefix) || !strings.HasSuffix(trimmed, markerSuffix) {
			continue
		}
		codePath := strings.TrimSuffix(strings.TrimPrefix(trimmed, markerPrefix), markerSuffix)
		code, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(codePath)))
		if err != nil {
			return nil, fmt.Errorf("%s:%d: marker references unreadable file: %w", mdPath, i+1, err)
		}
		if i+1 >= len(lines) || !strings.HasPrefix(lines[i+1], "```") {
			return nil, fmt.Errorf("%s:%d: marker must be immediately followed by a ``` fence", mdPath, i+1)
		}
		// Skip the old fenced block.
		j := i + 2
		for ; ; j++ {
			if j >= len(lines) {
				return nil, fmt.Errorf("%s:%d: unterminated ``` fence", mdPath, i+2)
			}
			if strings.TrimRight(lines[j], "\n\r \t") == "```" {
				break
			}
		}
		// Skip an existing source link (with optional preceding blank line).
		k := j + 1
		if k < len(lines) && strings.TrimSpace(lines[k]) == "" &&
			k+1 < len(lines) && strings.HasPrefix(lines[k+1], linkPrefix) {
			k += 2
		} else if k < len(lines) && strings.HasPrefix(lines[k], linkPrefix) {
			k++
		}
		rel, err := filepath.Rel(filepath.Dir(mdPath), filepath.Join(root, filepath.FromSlash(codePath)))
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", mdPath, i+1, err)
		}
		fmt.Fprintf(&out, "```%s\n%s\n```\n\n%s%s)\n",
			fenceLang(codePath), bytes.TrimRight(code, "\n"), linkPrefix, filepath.ToSlash(rel))
		i = k - 1
	}
	return []byte(out.String()), nil
}

func fenceLang(codePath string) string {
	switch path.Ext(codePath) {
	case ".go":
		return "go"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp":
		return "cpp"
	case ".py":
		return "python"
	case ".sh":
		return "sh"
	default:
		return ""
	}
}
