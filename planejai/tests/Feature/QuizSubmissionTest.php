<?php

use App\Models\{User, BnccSkill, LessonPlan, StudentAttempt};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

function trilhaComQuiz(): string {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-QUIZ']);
    $trail->publicar();
    $quiz = $trail->quiz()->create([]);
    $quiz->questions()->create(['enunciado' => 'Q1', 'opcoes' => ['a', 'b'], 'correta' => 0]);
    $quiz->questions()->create(['enunciado' => 'Q2', 'opcoes' => ['x', 'y'], 'correta' => 1]);
    return 'TR-QUIZ';
}

test('submissão corrige respostas e pontua', function () {
    $codigo = trilhaComQuiz();
    $perguntas = \App\Models\StudyTrail::where('codigo', $codigo)->first()->quiz->questions;

    $resp = $this->post("/t/{$codigo}/quiz", [
        'nome_aluno' => 'Ana',
        'respostas' => [
            $perguntas[0]->id => 0, // correta
            $perguntas[1]->id => 0, // errada
        ],
    ]);

    $resp->assertOk()->assertSee('10'); // 1 acerto = 10 pontos

    $attempt = StudentAttempt::where('nome_aluno', 'Ana')->first();
    expect($attempt->pontos)->toBe(10)
        ->and($attempt->concluido_em)->not->toBeNull()
        ->and($attempt->answers)->toHaveCount(2)
        ->and($attempt->answers->where('correta', true))->toHaveCount(1);
});
