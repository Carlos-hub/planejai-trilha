package seed

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBnccJSONParses(t *testing.T) {
	b, err := os.ReadFile("../../seed/bncc.json")
	if err != nil {
		t.Fatal(err)
	}
	var items []struct {
		Code       string  `json:"code"`
		Etapa      string  `json:"etapa"`
		Disciplina string  `json:"disciplina"`
		Anos       []int32 `json:"anos"`
		Descricao  string  `json:"descricao"`
	}
	if err := json.Unmarshal(b, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) < 10 {
		t.Fatalf("want >=10 skills got %d", len(items))
	}
	for _, it := range items {
		if it.Code == "" || it.Etapa == "" || it.Disciplina == "" || len(it.Anos) == 0 || it.Descricao == "" {
			t.Fatalf("incomplete field in %+v", it)
		}
	}
}
