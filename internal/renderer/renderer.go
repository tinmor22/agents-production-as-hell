package renderer

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Format constants for card dimensions.
const (
	FormatSquare   = "square"
	FormatPortrait = "portrait"

	squareW   = 1080
	squareH   = 1080
	portraitW = 1080
	portraitH = 1350
)

// ErrUnknownFormat is returned for unsupported format strings.
var ErrUnknownFormat = fmt.Errorf("formato desconocido: usa 'square' o 'portrait'")

// Renderer draws stat cards as PNG images.
type Renderer struct {
	regularFont *opentype.Font
	boldFont    *opentype.Font
}

// New creates a Renderer, loading fonts from the embedded FS.
func New(assets fs.FS) (*Renderer, error) {
	regularData, err := fs.ReadFile(assets, "fonts/Inter-Regular.ttf")
	if err != nil {
		return nil, fmt.Errorf("load regular font: %w", err)
	}
	boldData, err := fs.ReadFile(assets, "fonts/Inter-Bold.ttf")
	if err != nil {
		return nil, fmt.Errorf("load bold font: %w", err)
	}
	regularFont, err := opentype.Parse(regularData)
	if err != nil {
		return nil, fmt.Errorf("parse regular font: %w", err)
	}
	boldFont, err := opentype.Parse(boldData)
	if err != nil {
		return nil, fmt.Errorf("parse bold font: %w", err)
	}
	return &Renderer{
		regularFont: regularFont,
		boldFont:    boldFont,
	}, nil
}

// RenderCard generates PNG files for the given formats and returns the file paths.
func (r *Renderer) RenderCard(data CardData, formats []string, outputDir string) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	var paths []string
	for _, format := range formats {
		w, h, err := dimensionsFor(format)
		if err != nil {
			return nil, err
		}
		img := r.drawCard(data, w, h)
		path := cardFilename(data, w, h, outputDir)
		if err := savePNG(img, path); err != nil {
			return nil, fmt.Errorf("save png %s: %w", format, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func dimensionsFor(format string) (int, int, error) {
	switch format {
	case FormatSquare:
		return squareW, squareH, nil
	case FormatPortrait:
		return portraitW, portraitH, nil
	default:
		return 0, 0, ErrUnknownFormat
	}
}

func cardFilename(data CardData, w, h int, dir string) string {
	a := sanitize(data.EntityA.Name)
	b := sanitize(data.EntityB.Name)
	ts := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("statcard-%s-vs-%s-%s-%dx%d.png", a, b, ts, w, h)
	return filepath.Join(dir, name)
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Keep only alphanumeric, dash, and common accented chars
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' ||
			r >= 0x00C0 && r <= 0x024F { // Latin Extended
			b.WriteRune(r)
		}
	}
	return b.String()
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return nil
}

// drawCard renders the entire card onto an image.RGBA.
func (r *Renderer) drawCard(data CardData, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bg := color.RGBA{R: 26, G: 26, B: 46, A: 255}
	fillRect(img, 0, 0, w, h, bg)

	// Title
	titleFace := r.makeFace(r.boldFont, 44)
	drawCenteredText(img, titleFace, data.Title, w/2, 100, color.White)

	// Subtitle
	subFace := r.makeFace(r.regularFont, 26)
	drawCenteredText(img, subFace, data.Subtitle, w/2, 155, color.RGBA{R: 200, G: 200, B: 200, A: 255})

	// Entity names
	entityFace := r.makeFace(r.boldFont, 34)
	colorA := parseHexColor(data.EntityA.AccentColor)
	colorB := parseHexColor(data.EntityB.AccentColor)
	drawCenteredText(img, entityFace, data.EntityA.Name, w/4, 260, colorA)
	drawCenteredText(img, entityFace, data.EntityB.Name, 3*w/4, 260, colorB)

	// Divider line
	dividerY := 310
	fillRect(img, 0, dividerY, w, dividerY+2, color.RGBA{R: 60, G: 60, B: 60, A: 255})

	// Stats rows
	statLabelFace := r.makeFace(r.regularFont, 24)
	statValueFace := r.makeFace(r.boldFont, 28)
	rowHeight := 70
	startY := dividerY + 60

	for i, stat := range data.Stats {
		y := startY + i*rowHeight
		if y > h-100 {
			break
		}
		// Value A (left)
		drawCenteredText(img, statValueFace, stat.ValueA, w/4, y, colorA)
		// Label (center)
		drawCenteredText(img, statLabelFace, stat.Label, w/2, y, color.RGBA{R: 180, G: 180, B: 180, A: 255})
		// Value B (right)
		drawCenteredText(img, statValueFace, stat.ValueB, 3*w/4, y, colorB)

		// Row separator
		sepY := y + 30
		fillRect(img, 80, sepY, w-80, sepY+1, color.RGBA{R: 40, G: 40, B: 50, A: 255})
	}

	// Watermark
	wmFace := r.makeFace(r.regularFont, 16)
	drawCenteredText(img, wmFace, data.Watermark, w/2, h-30, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	return img
}

func (r *Renderer) makeFace(f *opentype.Font, size float64) font.Face {
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		// Should never happen with a valid parsed font
		panic(fmt.Sprintf("create font face: %v", err))
	}
	return face
}

func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			img.Set(x, y, c)
		}
	}
}

func drawCenteredText(img *image.RGBA, face font.Face, text string, cx, cy int, c color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
	}
	bounds, _ := d.BoundString(text)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	d.Dot = fixed.Point26_6{
		X: fixed.I(cx - textWidth/2),
		Y: fixed.I(cy),
	}
	d.DrawString(text)
}

func parseHexColor(hex string) color.RGBA {
	if len(hex) == 0 {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
