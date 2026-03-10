package renderer

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/statcard/statcard/templates"
)

func testData() CardData {
	return CardData{
		Title:    "Messi vs Cristiano",
		Subtitle: "Comparacion historica",
		EntityA: EntityInfo{
			Name:        "Lionel Messi",
			AccentColor: "#00A3E0",
		},
		EntityB: EntityInfo{
			Name:        "Cristiano Ronaldo",
			AccentColor: "#E4002B",
		},
		Stats: []StatRow{
			{Label: "Goles", ValueA: "821", ValueB: "899"},
			{Label: "Asistencias", ValueA: "361", ValueB: "233"},
			{Label: "Partidos", ValueA: "1050", ValueB: "1180"},
			{Label: "Ballon d'Or", ValueA: "8", ValueB: "5"},
		},
		Watermark:   "@futbol_stats",
		GeneratedAt: "2026-03-08",
	}
}

func TestRenderCard_TwoFormats(t *testing.T) {
	r, err := New(templates.FS)
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}

	outDir := t.TempDir()
	paths, err := r.RenderCard(testData(), []string{FormatSquare, FormatPortrait}, outDir)
	if err != nil {
		t.Fatalf("RenderCard: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			t.Fatalf("open %s: %v", p, err)
		}
		defer f.Close()

		img, err := png.Decode(f)
		if err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		bounds := img.Bounds()
		if bounds.Dx() < 1080 {
			t.Errorf("%s: width %d < 1080", p, bounds.Dx())
		}
	}
}

func TestRenderCard_Filename(t *testing.T) {
	r, err := New(templates.FS)
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}

	outDir := t.TempDir()
	paths, err := r.RenderCard(testData(), []string{FormatSquare}, outDir)
	if err != nil {
		t.Fatalf("RenderCard: %v", err)
	}

	name := filepath.Base(paths[0])
	if !strings.Contains(name, "messi") || !strings.Contains(name, "ronaldo") {
		t.Errorf("filename should contain entity names, got: %s", name)
	}
	if !strings.HasPrefix(name, "statcard-") {
		t.Errorf("filename should start with 'statcard-', got: %s", name)
	}
	if !strings.HasSuffix(name, ".png") {
		t.Errorf("filename should end with '.png', got: %s", name)
	}
	if !strings.Contains(name, "1080x1080") {
		t.Errorf("filename should contain dimensions, got: %s", name)
	}
}

func TestRenderCard_UnknownFormat(t *testing.T) {
	r, err := New(templates.FS)
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}

	_, err = r.RenderCard(testData(), []string{"banana"}, t.TempDir())
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}
