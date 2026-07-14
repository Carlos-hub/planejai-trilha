package domain

// Score computes the server-side quiz result for a set of student answers.
//
// answers maps quiz_question_id -> the option index the student chose.
// correct maps quiz_question_id -> the correct option index (server-side
// answer key; must never be sourced from client input).
//
// total is the number of questions in the answer key (correct). Only
// question ids present in correct are scored; stray or missing answer ids
// are ignored. Each correct answer is worth 10 points.
func Score(answers map[int64]int, correct map[int64]int) (pontos, acertos, total int) {
	total = len(correct)
	for id, want := range correct {
		if got, ok := answers[id]; ok && got == want {
			acertos++
		}
	}
	pontos = acertos * 10
	return pontos, acertos, total
}
