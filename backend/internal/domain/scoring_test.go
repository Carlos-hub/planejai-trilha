package domain

import "testing"

func TestScore_AllCorrect(t *testing.T) {
	correct := map[int64]int{1: 0, 2: 1, 3: 2}
	answers := map[int64]int{1: 0, 2: 1, 3: 2}

	pontos, acertos, total := Score(answers, correct)

	if total != 3 {
		t.Fatalf("total want 3 got %d", total)
	}
	if acertos != 3 {
		t.Fatalf("acertos want 3 got %d", acertos)
	}
	if pontos != 30 {
		t.Fatalf("pontos want 30 got %d", pontos)
	}
}

func TestScore_NoneCorrect(t *testing.T) {
	correct := map[int64]int{1: 0, 2: 1, 3: 2}
	answers := map[int64]int{1: 1, 2: 2, 3: 0}

	pontos, acertos, total := Score(answers, correct)

	if total != 3 {
		t.Fatalf("total want 3 got %d", total)
	}
	if acertos != 0 {
		t.Fatalf("acertos want 0 got %d", acertos)
	}
	if pontos != 0 {
		t.Fatalf("pontos want 0 got %d", pontos)
	}
}

func TestScore_Partial(t *testing.T) {
	correct := map[int64]int{1: 0, 2: 1, 3: 2, 4: 0}
	answers := map[int64]int{1: 0, 2: 2, 3: 2, 4: 1}

	pontos, acertos, total := Score(answers, correct)

	if total != 4 {
		t.Fatalf("total want 4 got %d", total)
	}
	if acertos != 2 {
		t.Fatalf("acertos want 2 got %d", acertos)
	}
	if pontos != 20 {
		t.Fatalf("pontos want 20 got %d", pontos)
	}
}

func TestScore_StrayAndMissingAnswerIDs(t *testing.T) {
	correct := map[int64]int{1: 0, 2: 1}
	// question 3 is a stray id not present in the answer key; question 2 is
	// unanswered entirely.
	answers := map[int64]int{1: 0, 3: 5}

	pontos, acertos, total := Score(answers, correct)

	if total != 2 {
		t.Fatalf("total want 2 got %d", total)
	}
	if acertos != 1 {
		t.Fatalf("acertos want 1 got %d", acertos)
	}
	if pontos != 10 {
		t.Fatalf("pontos want 10 got %d", pontos)
	}
}
