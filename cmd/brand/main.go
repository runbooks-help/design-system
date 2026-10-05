// SPDX-License-Identifier: FSL-1.1-MIT

// Command brand generates the Runbooks brand assets into brand/: the logomark,
// the wordmark, and horizontal + stacked lockups, for light and dark
// backgrounds, as outlined SVG (text converted to paths, so the files need no
// font installed) plus PNG at 512 and 1024. It also emits the 1200x630 social
// card served as the link preview image.
//
// The palette is the design system's (see tokens.css); the glyphs
// come from the self-hosted fonts. Run from the repository root. Needs
// woff2_decompress, inkscape and rsvg-convert on PATH — generation is a local,
// manual task and the output is committed.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	outDir    = "brand"
	fontsDir  = "fonts"
	shadeFont = "Shade Mono"
	sansFont  = "Atkinson Hyperlegible Next"
	monoFont  = "Atkinson Hyperlegible Mono"
)

// Palette, copied from tokens.css. The logomark always uses the
// dark-theme brand greens; the wordmark follows the background it is drawn for.
const (
	markHead    = "#3d6756" // --accent (dark)
	markBright  = "#6aa88d" // --accent-strong (dark)
	onAccent    = "#eef4f0" // --on-accent (dark)
	inkOnDark   = "#d7d9e3" // --nav-text (dark theme)
	inkOnLight  = "#2e2d38" // --nav-text (light theme)
	runOnDark   = "#6aa88d" // --accent-strong (dark theme)
	runOnLight  = "#345c4c" // --accent-strong (light theme)
	markRune    = "\u2591"  // ░
	markRuneAlt = "\u2593"  // ▓

	cardBg   = "#17161d" // --bg (dark theme)
	cardText = "#e9eaf1" // --text (dark theme)
	cardLink = "#9fb0cf" // --link (dark theme)
	cardGlow = "#3d6756" // --accent (dark theme)
)

// Geometry of the social card. It is a fixed canvas rather than a drawing to be
// cropped, so these are the card's own points, not tokens.
const (
	cardWidth  = 1200 // the OG image width
	cardHeight = 630
	cardPadX   = 72
	cardInkY   = 566 // baseline of the bottom row: lockup and host
	// The card re-uses the horizontal lockup at 236px wide, placed by the
	// drawing coordinates the standalone lockup is cropped from, so its mark and
	// wordmark cannot drift from that file.
	lockupNaturalWidth = 1401.5649
	cardLockupWidth    = 236.0
)

// sources are the self-hosted woff2 files the outliner needs: the pinned shade
// subset for the mark, the sans for the wordmark and headlines, and the mono
// for the card's eyebrow and host.
var sources = []string{
	"noto-sans-mono-shades.woff2",
	"atkinson-hyperlegible-next-latin-wght-normal.woff2",
	"atkinson-hyperlegible-mono-latin-wght-normal.woff2",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "brand:", err)
		os.Exit(1)
	}
}

func run() error {
	for _, tool := range []string{"woff2_decompress", "inkscape", "rsvg-convert"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found on PATH (see docs / mise)", tool)
		}
	}

	tmp, err := os.MkdirTemp("", "runbooks-brand-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	fontEnv, err := fontEnvironment(tmp)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	for _, v := range variants() {
		svgPath := filepath.Join(outDir, v.name+".svg")
		if err := outline(tmp, v, fontEnv, svgPath); err != nil {
			return err
		}
		for _, size := range v.widths() {
			png := filepath.Join(outDir, fmt.Sprintf("%s-%d.png", v.name, size))
			if err := rasterise(svgPath, png, size); err != nil {
				return err
			}
		}
		fmt.Println("wrote", filepath.Join(outDir, v.name+".svg"))
	}

	// The app's favicon is the icon, so the served mark cannot drift from the
	// generated one.
	if err := copyFile(filepath.Join(outDir, "icon.svg"), "favicon.svg"); err != nil {
		return err
	}
	fmt.Println("wrote favicon.svg")
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// variant is one exported asset: a base filename, the source SVG body, and the
// PNG widths to rasterise it at. A fixed variant keeps its exact canvas (a link
// preview is 1200x630 and nothing else); a loose one is cropped to its drawing
// and rasterised at the usual two sizes.
type variant struct {
	name  string
	body  string
	fixed bool
	sizes []int
}

func (v variant) widths() []int {
	if len(v.sizes) > 0 {
		return v.sizes
	}
	return []int{1024, 512}
}

// canvas wraps an SVG body. A loose variant gets a generous viewBox and lets
// inkscape crop to the drawing; a fixed one is drawn at its exact export size.
func canvas(v variant) string {
	w, h := 1400, 700
	if v.fixed {
		w, h = cardWidth, cardHeight
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">%s</svg>`, w, h, w, h, v.body)
}

// mark draws the ░▓ logomark; its glyphs come from the pinned shade subset.
func mark(x, y, size int, anchor string) string {
	anchorAttr := ""
	if anchor != "" {
		anchorAttr = fmt.Sprintf(` text-anchor="%s"`, anchor)
	}
	return fmt.Sprintf(`<text x="%d" y="%d" font-family="%s" font-weight="400" font-size="%d"%s><tspan fill="%s">%s</tspan><tspan fill="%s">%s</tspan></text>`,
		x, y, shadeFont, size, anchorAttr, markHead, markRune, markBright, markRuneAlt)
}

// word draws the wordmark: "run" in the caller's green, "books" in its ink.
func word(x, y, size int, run, ink, anchor string) string {
	anchorAttr := ""
	if anchor != "" {
		anchorAttr = fmt.Sprintf(` text-anchor="%s"`, anchor)
	}
	return fmt.Sprintf(`<text x="%d" y="%d" font-family="%s" font-weight="700" font-size="%d"%s><tspan fill="%s">run</tspan><tspan fill="%s">books</tspan></text>`,
		x, y, sansFont, size, anchorAttr, run, ink)
}

// lockupHorizontal is the mark and the wordmark side by side, in the drawing
// coordinates the standalone lockup asset is cropped from.
func lockupHorizontal(run, ink string) string {
	return mark(420, 460, 300, "") + word(880, 460, 220, run, ink, "")
}

// socialCard draws the 1200x630 link preview: the marketing statement, the
// horizontal lockup, and the site's host. It is dark-only, because a preview
// crawler never themes an OG image.
func socialCard() string {
	scale := cardLockupWidth / lockupNaturalWidth
	lockup := fmt.Sprintf(`<g transform="translate(%.2f %.2f) scale(%.5f)">%s</g>`,
		cardPadX-420*scale, cardInkY-460*scale, scale, lockupHorizontal(runOnDark, inkOnDark))

	return `<defs>` +
		`<radialGradient id="glow" gradientUnits="userSpaceOnUse" cx="0" cy="0" r="1" gradientTransform="translate(936 50) scale(900 520)">` +
		`<stop offset="0" stop-color="` + cardGlow + `" stop-opacity="0.42"/>` +
		`<stop offset="0.66" stop-color="` + cardGlow + `" stop-opacity="0"/>` +
		`</radialGradient>` +
		`</defs>` +
		// Opaque: a crawler composites a transparent PNG unpredictably.
		fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, cardWidth, cardHeight, cardBg) +
		fmt.Sprintf(`<rect width="%d" height="%d" fill="url(#glow)"/>`, cardWidth, cardHeight) +
		fmt.Sprintf(`<text x="%d" y="84" font-family="%s" font-weight="450" font-size="14" letter-spacing="2.24" fill="%s">SELF-HOSTED · SOURCE-AVAILABLE · FSL-1.1-MIT</text>`, cardPadX, monoFont, runOnDark) +
		fmt.Sprintf(`<text x="%d" y="189" font-family="%s" font-weight="700" font-size="96" letter-spacing="-3.36" fill="%s">Prod broke?</text>`, cardPadX, sansFont, cardText) +
		fmt.Sprintf(`<text x="%d" y="285" font-family="%s" font-weight="700" font-size="96" letter-spacing="-3.36" fill="%s">Runbooks help.</text>`, cardPadX, sansFont, runOnDark) +
		fmt.Sprintf(`<text x="%d" y="351" font-family="%s" font-size="21" fill="%s" fill-opacity="0.78">Markdown in, step-by-step procedures out. No database, no</text>`, cardPadX, sansFont, cardText) +
		fmt.Sprintf(`<text x="%d" y="380" font-family="%s" font-size="21" fill="%s" fill-opacity="0.78">CMS, no authoring UI.</text>`, cardPadX, sansFont, cardText) +
		lockup +
		fmt.Sprintf(`<text x="%d" y="%d" text-anchor="end" font-family="%s" font-weight="450" font-size="15" fill="%s">runbooks.help</text>`,
			cardWidth-cardPadX, cardInkY, monoFont, cardLink)
}

func variants() []variant {
	iconMark := fmt.Sprintf(`<text x="256" y="362" text-anchor="middle" font-family="%s" font-weight="400" font-size="300"><tspan fill="%s" fill-opacity="0.7">%s</tspan><tspan fill="%s">%s</tspan></text>`,
		shadeFont, onAccent, markRune, onAccent, markRuneAlt)
	icon := `<g transform="translate(444 94)">` +
		`<rect width="512" height="512" rx="112" fill="` + markHead + `"/>` +
		iconMark +
		`</g>`

	return []variant{
		{name: "logomark", body: mark(500, 460, 320, "")},
		{name: "wordmark-on-dark", body: word(500, 460, 300, runOnDark, inkOnDark, "")},
		{name: "wordmark-on-light", body: word(500, 460, 300, runOnLight, inkOnLight, "")},
		{name: "lockup-horizontal-on-dark", body: lockupHorizontal(runOnDark, inkOnDark)},
		{name: "lockup-horizontal-on-light", body: lockupHorizontal(runOnLight, inkOnLight)},
		{name: "lockup-stacked-on-dark", body: mark(700, 420, 300, "middle") + word(700, 620, 200, runOnDark, inkOnDark, "middle")},
		{name: "lockup-stacked-on-light", body: mark(700, 420, 300, "middle") + word(700, 620, 200, runOnLight, inkOnLight, "middle")},
		{name: "icon", body: icon},
		{name: "social-card-on-dark", body: socialCard(), fixed: true, sizes: []int{cardWidth}},
	}
}

// fontEnvironment decompresses the self-hosted woff2 files and writes a
// fontconfig file pointing only at them, so the outliner uses the pinned fonts
// regardless of what is installed system-wide. It returns the environment for
// the inkscape calls.
func fontEnvironment(tmp string) ([]string, error) {
	for _, name := range sources {
		data, err := os.ReadFile(filepath.Join(fontsDir, name))
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0o644); err != nil {
			return nil, err
		}
		cmd := exec.Command("woff2_decompress", name)
		cmd.Dir = tmp
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("woff2_decompress %s: %v: %s", name, err, out)
		}
	}

	conf := fmt.Sprintf(`<?xml version="1.0"?><!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig><dir>%s</dir><cachedir>%s</cachedir></fontconfig>
`, tmp, filepath.Join(tmp, "cache"))
	confPath := filepath.Join(tmp, "fonts.conf")
	if err := os.WriteFile(confPath, []byte(conf), 0o644); err != nil {
		return nil, err
	}
	return append(os.Environ(), "FONTCONFIG_FILE="+confPath), nil
}

// outline converts the source SVG's text to paths and crops to the drawing, so
// the exported file is portable and tightly bounded. A fixed variant keeps its
// canvas instead.
func outline(tmp string, v variant, env []string, dst string) error {
	src := filepath.Join(tmp, v.name+".src.svg")
	if err := os.WriteFile(src, []byte(canvas(v)), 0o644); err != nil {
		return err
	}
	args := []string{src, "--export-text-to-path", "--export-plain-svg", "--export-filename=" + dst}
	if !v.fixed {
		args = append(args, "--export-area-drawing", "--export-margin=4")
	}
	cmd := exec.Command("inkscape", args...)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("inkscape %s: %v: %s", v.name, err, out)
	}
	return nil
}

func rasterise(svg, png string, width int) error {
	cmd := exec.Command("rsvg-convert", "-w", fmt.Sprint(width), svg, "-o", png)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rsvg-convert %s: %v: %s", png, err, out)
	}
	return nil
}
