package templates

import (
	"testing"
)

func TestEmbed_CardSVG(t *testing.T) {
	data, err := FS.ReadFile("card.svg")
	if err != nil {
		t.Fatalf("ReadFile card.svg: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("card.svg is empty")
	}
}

func TestEmbed_FontRegular(t *testing.T) {
	data, err := FS.ReadFile("fonts/Inter-Regular.ttf")
	if err != nil {
		t.Fatalf("ReadFile Inter-Regular.ttf: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Inter-Regular.ttf is empty")
	}
}

func TestEmbed_FontBold(t *testing.T) {
	data, err := FS.ReadFile("fonts/Inter-Bold.ttf")
	if err != nil {
		t.Fatalf("ReadFile Inter-Bold.ttf: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Inter-Bold.ttf is empty")
	}
}
