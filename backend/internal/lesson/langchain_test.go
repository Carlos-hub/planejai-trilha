package lesson

import (
	"testing"
)

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
