package store

import (
	"fmt"
	"strings"
)

// AnoLabel renders a BNCC skill's school-year span as human-readable text
// (e.g. "6º e 7º ano", "1º ano"). Ensino Médio skills span all three séries and
// carry no per-year data, so they collapse to a single label.
func (s BnccSkill) AnoLabel() string {
	if s.Etapa == "EM" {
		return "Ensino Médio"
	}
	if len(s.Anos) == 0 {
		return ""
	}
	parts := make([]string, len(s.Anos))
	for i, a := range s.Anos {
		parts[i] = fmt.Sprintf("%dº ano", a)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " e " + parts[len(parts)-1]
}
