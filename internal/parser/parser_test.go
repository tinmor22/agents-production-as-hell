package parser

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantA     string
		wantB     string
		wantCtx   string
		wantType  string
		wantErr   bool
		errTarget error
	}{
		{
			name:     "VS separator",
			input:    "Messi vs Cristiano",
			wantA:    "Messi",
			wantB:    "Cristiano",
			wantCtx:  "",
			wantType: QueryTypeH2H,
		},
		{
			name:     "contra separator with context",
			input:    "messi contra mbappé en Champions League",
			wantA:    "messi",
			wantB:    "mbappé",
			wantCtx:  "Champions League",
			wantType: QueryTypeH2H,
		},
		{
			name:      "no separator",
			input:     "solo messi",
			wantErr:   true,
			errTarget: ErrNoSeparator,
		},
		{
			name:     "accented chars preserved",
			input:    "Mbappé vs Vinícius Jr",
			wantA:    "Mbappé",
			wantB:    "Vinícius Jr",
			wantCtx:  "",
			wantType: QueryTypeH2H,
		},
		{
			name:     "VS uppercase",
			input:    "Real Madrid VS Barcelona",
			wantA:    "Real Madrid",
			wantB:    "Barcelona",
			wantCtx:  "",
			wantType: QueryTypeH2H,
		},
		{
			name:      "empty input",
			input:     "",
			wantErr:   true,
			errTarget: ErrNoSeparator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
					t.Errorf("expected error %v, got: %v", tt.errTarget, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.EntityA != tt.wantA {
				t.Errorf("EntityA: got %q, want %q", got.EntityA, tt.wantA)
			}
			if got.EntityB != tt.wantB {
				t.Errorf("EntityB: got %q, want %q", got.EntityB, tt.wantB)
			}
			if got.Context != tt.wantCtx {
				t.Errorf("Context: got %q, want %q", got.Context, tt.wantCtx)
			}
			if got.QueryType != tt.wantType {
				t.Errorf("QueryType: got %q, want %q", got.QueryType, tt.wantType)
			}
		})
	}
}
