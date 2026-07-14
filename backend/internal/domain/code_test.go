package domain

import (
	"testing"
)

func TestNewTrailCode(t *testing.T) {
	t.Run("has TR- prefix", func(t *testing.T) {
		code := NewTrailCode()
		if len(code) < 3 || code[:3] != "TR-" {
			t.Errorf("expected prefix 'TR-', got '%s'", code[:3])
		}
	})

	t.Run("has total length of 7", func(t *testing.T) {
		code := NewTrailCode()
		if len(code) != 7 {
			t.Errorf("expected length 7, got %d for code '%s'", len(code), code)
		}
	})

	t.Run("only contains allowed characters after prefix", func(t *testing.T) {
		const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
		code := NewTrailCode()
		suffix := code[3:] // characters after "TR-"

		for _, ch := range suffix {
			if !containsRune(alphabet, ch) {
				t.Errorf("character '%c' not in allowed alphabet", ch)
			}
		}
	})

	t.Run("1000 generated codes are unique", func(t *testing.T) {
		codes := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			code := NewTrailCode()
			if codes[code] {
				t.Errorf("collision detected: code '%s' generated twice", code)
			}
			codes[code] = true
		}

		if len(codes) != 1000 {
			t.Errorf("expected 1000 unique codes, got %d", len(codes))
		}
	})
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}
