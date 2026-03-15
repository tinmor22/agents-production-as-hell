package renderer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io/fs"
	"math"
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
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' ||
			r >= 0x00C0 && r <= 0x024F {
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

	colorA := parseHexColor(data.EntityA.AccentColor)
	colorB := parseHexColor(data.EntityB.AccentColor)

	// 1. Dark base background
	fillRect(img, 0, 0, w, h, color.RGBA{R: 10, G: 10, B: 20, A: 255})

	// 2. Player photos as background (each on their half), drawn before gradients
	if len(data.EntityA.PhotoData) > 0 {
		drawPhotoBackground(img, data.EntityA.PhotoData, 0, 0, w/2, h, true)
	}
	if len(data.EntityB.PhotoData) > 0 {
		drawPhotoBackground(img, data.EntityB.PhotoData, w/2, 0, w, h, false)
	}

	// 3. Color gradient overlay on each half (left = colorA, right = colorB)
	drawHalfGradient(img, 0, 0, w/2, h, colorA, 0.45)
	drawHalfGradient(img, w/2, 0, w, h, colorB, 0.45)

	// 4. Dark vignette top + bottom bars for text readability
	drawVerticalGradient(img, 0, 0, w, 220, color.RGBA{R: 5, G: 5, B: 10, A: 255}, 0.92, 0.0)
	drawVerticalGradient(img, 0, h-180, w, h, color.RGBA{R: 5, G: 5, B: 10, A: 255}, 0.0, 0.95)

	// 5. Central dark strip for the stats table
	statsTop := h/2 - 170
	statsBot := h/2 + 170
	if h > 1080 {
		statsTop = h/2 - 220
		statsBot = h/2 + 220
	}
	drawCentralStrip(img, statsTop, statsBot, w)

	// 6. Diagonal center divider
	drawDiagonalDivider(img, w, h)

	// 7. Top section: entity names
	nameFace := r.makeFace(r.boldFont, 48)
	nameY := 130
	drawCenteredText(img, nameFace, strings.ToUpper(data.EntityA.Name), w/4, nameY, color.White)
	drawCenteredText(img, nameFace, strings.ToUpper(data.EntityB.Name), 3*w/4, nameY, color.White)

	// Subtitle below names
	teamFace := r.makeFace(r.regularFont, 22)
	teamColor := color.RGBA{R: 200, G: 200, B: 210, A: 220}
	drawCenteredText(img, teamFace, data.Subtitle, w/2, 175, teamColor)

	// 8. VS circle in center
	drawVSCircle(img, w/2, (nameY+statsTop)/2+10, 44, r)

	// 9. Stats table inside the central strip
	statValueFace := r.makeFace(r.boldFont, 52)
	statLabelFace := r.makeFace(r.regularFont, 20)

	rowSpacing := (statsBot - statsTop) / (len(data.Stats) + 1)
	for i, stat := range data.Stats {
		y := statsTop + (i+1)*rowSpacing

		// Highlight winner with a subtle glow behind the number
		winA, winB := compareValues(stat.ValueA, stat.ValueB)
		valColorA := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		valColorB := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		if winA {
			valColorA = colorA
			drawGlowRect(img, w/4-70, y-45, w/4+70, y+10, colorA, 0.15)
		}
		if winB {
			valColorB = colorB
			drawGlowRect(img, 3*w/4-70, y-45, 3*w/4+70, y+10, colorB, 0.15)
		}

		drawCenteredText(img, statValueFace, stat.ValueA, w/4, y, valColorA)
		drawCenteredText(img, statValueFace, stat.ValueB, 3*w/4, y, valColorB)
		drawCenteredText(img, statLabelFace, strings.ToUpper(stat.Label), w/2, y-10, color.RGBA{R: 170, G: 170, B: 185, A: 255})

		// Thin separator
		if i < len(data.Stats)-1 {
			sepY := y + 22
			fillRect(img, 80, sepY, w-80, sepY+1, color.RGBA{R: 60, G: 60, B: 80, A: 120})
		}
	}

	// 10. Bottom: subtitle and watermark
	subtitleFace := r.makeFace(r.regularFont, 20)
	drawCenteredText(img, subtitleFace, data.Subtitle, w/2, h-90, color.RGBA{R: 160, G: 160, B: 175, A: 230})

	wmFace := r.makeFace(r.boldFont, 18)
	drawCenteredText(img, wmFace, data.Watermark, w/2, h-55, color.RGBA{R: 130, G: 130, B: 145, A: 200})

	genFace := r.makeFace(r.regularFont, 14)
	drawCenteredText(img, genFace, data.GeneratedAt, w/2, h-25, color.RGBA{R: 90, G: 90, B: 105, A: 180})

	return img
}

// drawPhotoBackground draws a player photo stretched to fill half the card at low opacity.
func drawPhotoBackground(img *image.RGBA, photoData []byte, x0, y0, x1, y1 int, mirrorX bool) {
	src, _, err := image.Decode(bytes.NewReader(photoData))
	if err != nil {
		return
	}
	bounds := src.Bounds()
	sw := bounds.Dx()
	sh := bounds.Dy()
	dw := x1 - x0
	dh := y1 - y0

	for dy := 0; dy < dh; dy++ {
		for dx := 0; dx < dw; dx++ {
			sx := dx * sw / dw
			sy := dy * sh / dh
			if mirrorX {
				sx = sw - 1 - sx
			}
			if sx < 0 {
				sx = 0
			}
			if sx >= sw {
				sx = sw - 1
			}
			srcC := src.At(bounds.Min.X+sx, bounds.Min.Y+sy)
			r, g, b, _ := srcC.RGBA()
			// Draw at ~55% opacity
			blendPixel(img, x0+dx, y0+dy, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: 255,
			}, 0.55)
		}
	}
}

// drawHalfGradient overlays a semi-transparent color on a region.
func drawHalfGradient(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, opacity float64) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			blendPixel(img, x, y, color.RGBA{R: c.R / 3, G: c.G / 3, B: c.B / 3, A: 255}, opacity)
		}
	}
}

// drawVerticalGradient draws a vertical fade from startOpacity (top) to endOpacity (bottom).
func drawVerticalGradient(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, startOp, endOp float64) {
	h := y1 - y0
	if h <= 0 {
		return
	}
	for y := y0; y < y1; y++ {
		t := float64(y-y0) / float64(h)
		op := startOp + (endOp-startOp)*t
		for x := x0; x < x1; x++ {
			blendPixel(img, x, y, c, op)
		}
	}
}

// drawCentralStrip draws a dark translucent band for the stats.
func drawCentralStrip(img *image.RGBA, y0, y1, w int) {
	stripColor := color.RGBA{R: 8, G: 8, B: 18, A: 255}
	fade := 30
	for y := y0; y < y1; y++ {
		var op float64
		if y < y0+fade {
			op = float64(y-y0) / float64(fade) * 0.85
		} else if y > y1-fade {
			op = float64(y1-y) / float64(fade) * 0.85
		} else {
			op = 0.85
		}
		for x := 0; x < w; x++ {
			blendPixel(img, x, y, stripColor, op)
		}
	}
}

// drawDiagonalDivider draws a thin diagonal line in the center.
func drawDiagonalDivider(img *image.RGBA, w, h int) {
	cx := w / 2
	skew := 30 // pixels of diagonal lean
	lineColor := color.RGBA{R: 255, G: 255, B: 255, A: 60}
	for y := 0; y < h; y++ {
		t := float64(y) / float64(h)
		x := cx + int(float64(skew)*(t-0.5)*2)
		for dx := -1; dx <= 1; dx++ {
			px := x + dx
			if px >= 0 && px < w {
				alpha := float64(60-20*abs(dx)) / 255.0
				blendPixel(img, px, y, lineColor, alpha)
			}
		}
	}
}

// drawVSCircle draws a dark circle with "VS" text at the center.
func drawVSCircle(img *image.RGBA, cx, cy int, radius int, r *Renderer) {
	circleColor := color.RGBA{R: 12, G: 12, B: 22, A: 255}
	borderColor := color.RGBA{R: 200, G: 200, B: 210, A: 180}
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(radius) {
				x := cx + dx
				y := cy + dy
				if dist >= float64(radius)-2 {
					blendPixel(img, x, y, borderColor, 0.6)
				} else {
					img.Set(x, y, circleColor)
				}
			}
		}
	}
	vsFace := r.makeFace(r.boldFont, 28)
	drawCenteredText(img, vsFace, "VS", cx, cy+10, color.White)
}

// drawGlowRect draws a soft glowing rectangle behind a winning value.
func drawGlowRect(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, opacity float64) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			blendPixel(img, x, y, c, opacity)
		}
	}
}

func (r *Renderer) makeFace(f *opentype.Font, size float64) font.Face {
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
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

// blendPixel blends color c over the existing pixel at (x,y) with the given opacity.
func blendPixel(img *image.RGBA, x, y int, c color.RGBA, opacity float64) {
	if x < 0 || y < 0 || x >= img.Bounds().Max.X || y >= img.Bounds().Max.Y {
		return
	}
	existing := img.RGBAAt(x, y)
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(existing.R)*(1-opacity) + float64(c.R)*opacity),
		G: uint8(float64(existing.G)*(1-opacity) + float64(c.G)*opacity),
		B: uint8(float64(existing.B)*(1-opacity) + float64(c.B)*opacity),
		A: 255,
	})
}

func drawCenteredText(img draw.Image, face font.Face, text string, cx, cy int, c color.Color) {
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

// compareValues returns (winnerIsA, winnerIsB) by parsing numeric strings.
func compareValues(a, b string) (bool, bool) {
	var va, vb float64
	fmt.Sscanf(a, "%f", &va)
	fmt.Sscanf(b, "%f", &vb)
	if va > vb {
		return true, false
	}
	if vb > va {
		return false, true
	}
	return false, false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
