// Command genfont converts a TTF/OTF font into a Go source file holding a
// fixed-height bitmap font for the drivers/fonts package (page-major layout,
// see that package's documentation). Outline parsing and rasterization are
// done by github.com/soypat/lefevre; this tool only thresholds the coverage
// to 1 bit per pixel and packs the SSD1306-style pages.
//
// Example:
//
//	go run ./cmd/genfont -ttf DejaVuSans.ttf -size 12 -name FontSans12 \
//	    -out drivers/fonts/fontsans12.go
//
// Pass -preview to render sample text to the terminal instead of writing
// the file, to judge size and threshold before committing to an output:
//
//	go run ./cmd/genfont -ttf DejaVuSans.ttf -size 12 -name F -preview "Hello 123"
package main

import (
	"flag"
	"fmt"
	"go/format"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/soypat/lefevre"
	"github.com/soypat/lefevre/raster"
	"github.com/soypat/seeed-grove/drivers/fonts"
)

const asciiChars = ` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_` + "`" + `abcdefghijklmnopqrstuvwxyz{|}~`

func main() {
	var (
		ttfPath = flag.String("ttf", "", "path to TTF/OTF file (required)")
		size    = flag.Int("size", 12, "font height in pixels (ascent+descent)")
		name    = flag.String("name", "", "Go variable name, e.g. FontSans12 (required)")
		pkg     = flag.String("pkg", "fonts", "package name of the generated file")
		out     = flag.String("out", "", "output file (default: lowercased name + .go)")
		chars   = flag.String("chars", asciiChars, "characters to include")
		thresh  = flag.Int("threshold", 128, "coverage threshold 1..255 for a lit pixel")
		prev    = flag.String("preview", "", "render this text to stdout instead of writing the file")
	)
	flag.Parse()
	if *ttfPath == "" || *name == "" || *size < 1 || *size > 64 ||
		*thresh < 1 || *thresh > 255 || len(*chars) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	if *out == "" {
		*out = strings.ToLower(*name) + ".go"
	}
	if err := run(*ttfPath, *size, *name, *pkg, *out, *chars, uint8(*thresh), *prev); err != nil {
		fmt.Fprintln(os.Stderr, "genfont:", err)
		os.Exit(1)
	}
}

func run(ttfPath string, size int, name, pkg, out, chars string, thresh uint8, prev string) error {
	data, err := os.ReadFile(ttfPath)
	if err != nil {
		return err
	}
	fnt, err := lefevre.FontFromMemory(data, 0)
	if err != nil {
		return err
	}
	info := fnt.Info()
	if info.Ascent <= info.Descent {
		return fmt.Errorf("bad font metrics: ascent=%d descent=%d", info.Ascent, info.Descent)
	}
	// Scale so ascent+descent spans the requested pixel height; the baseline
	// sits ascentPx rows below the glyph box top.
	scale := float32(size) / float32(int(info.Ascent)-int(info.Descent))
	ascentPx := int(math.Round(float64(float32(info.Ascent) * scale)))
	pages := (size + 7) / 8

	runes := []rune(chars)
	slices.Sort(runes)
	runes = slices.Compact(runes)

	var (
		bitmaps  []byte
		glyphs   []byte
		rast     raster.ScanlineRasterizer
		segs     []lefevre.Segment
		alpha    []byte
		hasNotes []string
	)
	for _, r := range runes {
		id := fnt.GlyphID(r)
		if id == 0 && r != 0 {
			hasNotes = append(hasNotes, fmt.Sprintf("missing %q", r))
		}
		adv := int(math.Round(float64(float32(fnt.GlyphAdvance(id)) * scale)))
		gw, gh, xoff, yoff := raster.GlyphBox(fnt, id, scale)
		col0 := max(xoff, 0) // clip left overhang; the format has no negative offsets.
		cellW := max(adv, col0+gw)
		if cellW > 255 || adv > 255 {
			return fmt.Errorf("glyph %q wider than 255px", r)
		}
		cell := make([]byte, pages*cellW)
		if gw > 0 && gh > 0 {
			// Rasterize the tight glyph box; same call convention as
			// raster.PackConfig.BakeAtlas (Y-flipped segments, offsets from
			// the font-unit bounds).
			segs = fnt.GlyphOutline(segs[:0], id)
			for i := range segs {
				segs[i].Y = -segs[i].Y
				segs[i].Cy = -segs[i].Cy
			}
			xMin, _, _, yMax := fnt.GlyphBounds(id)
			alpha = append(alpha[:0], make([]byte, gw*gh)...)
			rast.Rasterize(alpha, gw, gh, gw, segs, scale, -float32(xMin)*scale, float32(yMax)*scale)
			for j := 0; j < gh; j++ {
				row := ascentPx + yoff + j
				if row < 0 || row >= size {
					continue // clip rows outside the font box.
				}
				for i := 0; i < gw; i++ {
					col := col0 + i
					if col >= cellW || alpha[j*gw+i] < thresh {
						continue
					}
					cell[(row/8)*cellW+col] |= 1 << (row % 8)
				}
			}
		}
		off := len(bitmaps)
		if off > math.MaxUint16 {
			return fmt.Errorf("font exceeds 64KB bitmap limit at %q", r)
		}
		bitmaps = append(bitmaps, cell...)
		glyphs = append(glyphs, byte(off>>8), byte(off), byte(cellW), byte(adv))
	}

	// Group consecutive runes into ranges.
	var ranges []fonts.GlyphRange
	for i := 0; i < len(runes); {
		j := i
		for j+1 < len(runes) && runes[j+1] == runes[j]+1 {
			j++
		}
		ranges = append(ranges, fonts.GlyphRange{FirstRune: runes[i], GlyphStart: uint16(i), Count: uint16(j - i + 1)})
		i = j + 1
	}

	if prev != "" {
		font := fonts.NewFont(uint8(size), string(bitmaps), string(glyphs), ranges)
		fmt.Printf("%s %dpx (%q %s): %d glyphs, %d bytes bitmap+index\n",
			name, size, info.Family, info.Subfamily, len(glyphs)/4, len(bitmaps)+len(glyphs))
		for _, note := range hasNotes {
			fmt.Printf("note: %s in source font, mapped to .notdef\n", note)
		}
		preview(os.Stdout, &font, prev)
		return nil // visualize only, no file written.
	}

	var rangeLits []string
	for _, rg := range ranges {
		rangeLits = append(rangeLits, fmt.Sprintf("{FirstRune: %q, GlyphStart: %d, Count: %d}", rg.FirstRune, rg.GlyphStart, rg.Count))
	}

	prefix := strings.ToLower(name)
	var b strings.Builder
	fmt.Fprintf(&b, `// Code generated by cmd/genfont -ttf %s -size %d -threshold %d; DO NOT EDIT.

package %s

const (
	%sbitmap = %q
	%sglyphs = %q
)

// %s is %q (%s) rendered at %d pixel height.
`, ttfPath, size, thresh, pkg, prefix, string(bitmaps), prefix, string(glyphs), name, info.Family, info.Subfamily, size)
	for _, n := range hasNotes {
		fmt.Fprintf(&b, "// Note: %s in source font, mapped to .notdef.\n", n)
	}
	fmt.Fprintf(&b, "var %s = NewFont(%d, %sbitmap, %sglyphs, []GlyphRange{%s})\n",
		name, size, prefix, prefix, strings.Join(rangeLits, ", "))

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("generated code does not format: %w", err)
	}
	return os.WriteFile(out, src, 0o644)
}

// preview renders text with the freshly built font as one terminal character
// per pixel, so glyph quality can be judged before writing the file.
func preview(w io.Writer, f *fonts.Font, text string) {
	h := f.Height()
	for _, line := range strings.Split(text, "\n") {
		rows := make([][]byte, h)
		for _, r := range line {
			bm, gw, adv := f.Glyph(r)
			for j := 0; j < h; j++ {
				for i := 0; i < adv; i++ {
					c := byte(' ')
					if i < gw && bm[(j/8)*gw+i]&(1<<(j%8)) != 0 {
						c = '#'
					}
					rows[j] = append(rows[j], c)
				}
			}
		}
		for _, row := range rows {
			fmt.Fprintf(w, "|%s|\n", row)
		}
	}
}
