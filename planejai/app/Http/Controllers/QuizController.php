<?php

namespace App\Http\Controllers;

use App\Models\StudyTrail;
use Illuminate\Http\Request;

class QuizController extends Controller
{
    public function show(string $codigo)
    {
        $trail = StudyTrail::published()->where('codigo', $codigo)
            ->with('quiz.questions')->firstOrFail();

        return view('quiz.show', ['trail' => $trail]);
    }

    public function submit(Request $request, string $codigo)
    {
        $trail = StudyTrail::published()->where('codigo', $codigo)
            ->with('quiz.questions')->firstOrFail();

        $validated = $request->validate([
            'nome_aluno' => 'required|string|max:120',
            'respostas' => 'required|array',
        ]);

        $attempt = $trail->attempts()->create([
            'nome_aluno' => $validated['nome_aluno'],
            'pontos' => 0,
        ]);

        $acertos = 0;
        foreach ($trail->quiz->questions as $q) {
            $escolhida = (int) ($validated['respostas'][$q->id] ?? -1);
            $correta = $escolhida === (int) $q->correta;
            if ($correta) {
                $acertos++;
            }
            $attempt->answers()->create([
                'quiz_question_id' => $q->id,
                'escolhida' => max(0, $escolhida),
                'correta' => $correta,
            ]);
        }

        $total = $trail->quiz->questions->count();
        $attempt->update(['pontos' => $acertos * 10, 'concluido_em' => now()]);

        return view('quiz.result', [
            'trail' => $trail,
            'attempt' => $attempt->fresh(),
            'acertos' => $acertos,
            'total' => $total,
        ]);
    }
}
