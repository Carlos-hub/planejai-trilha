package lesson

import (
	"context"
	"os"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// TestLangChainGeneratorGenerate exercises the real Anthropic provider via
// LangChainGo. It is skipped unless ANTHROPIC_API_KEY is set, since it makes
// a real network call.
func TestLangChainGeneratorGenerate(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set; skipping LangChainGo integration test")
	}

	gen, err := NewLangChainGenerator()
	if err != nil {
		t.Fatalf("NewLangChainGenerator failed: %v", err)
	}

	skill := store.BnccSkill{
		ID:         1,
		Code:       "EF01LP01",
		Disciplina: "Português",
		Etapa:      "EF",
		Anos:       []int32{1},
		Descricao:  "Reconhecer a função social de textos que circulam em campos da vida social dos quais participa cotidianamente",
	}

	ld, err := gen.Generate(context.Background(), skill, 45)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(ld.Trilha.Topicos) < 1 {
		t.Errorf("expected at least 1 topico, got %d", len(ld.Trilha.Topicos))
	}
	if len(ld.Trilha.Quiz.Questoes) < 1 {
		t.Errorf("expected at least 1 questao, got %d", len(ld.Trilha.Quiz.Questoes))
	}
	if ld.Plano.Objetivos == "" {
		t.Errorf("expected non-empty objetivos")
	}
}

// TestExtractJSON is a non-network unit test covering markdown-fenced and
// prose-wrapped LLM responses.
func TestExtractJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "fenced with json tag",
			in:   "```json\n{\"a\": 1}\n```",
			want: `{"a": 1}`,
		},
		{
			name: "fenced without tag",
			in:   "```\n{\"a\": 1}\n```",
			want: `{"a": 1}`,
		},
		{
			name: "prose wrapped",
			in:   "Aqui está o plano de aula:\n{\"a\": 1}\nEspero que ajude!",
			want: `{"a": 1}`,
		},
		{
			name: "plain json",
			in:   `{"a": 1}`,
			want: `{"a": 1}`,
		},
		{
			name: "nested braces",
			in:   "texto antes {\"a\": {\"b\": 2}} texto depois",
			want: `{"a": {"b": 2}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractJSON(tc.in)
			if string(got) != tc.want {
				t.Errorf("extractJSON(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractJSONNoBraces(t *testing.T) {
	got := extractJSON("no json here")
	if got != nil {
		t.Errorf("expected nil, got %q", got)
	}
}
