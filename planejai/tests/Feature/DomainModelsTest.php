<?php

use App\Models\{User, BnccSkill, LessonPlan, StudyTrail, TrailTopic, Quiz, QuizQuestion, StudentAttempt, AttemptAnswer};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('relações do domínio funcionam ponta a ponta', function () {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);

    $plan = LessonPlan::create([
        'user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50,
        'objetivos' => 'o', 'metodologia' => 'm', 'recursos' => 'r', 'avaliacao' => 'a',
        'atividade' => 'at', 'status' => 'gerado',
    ]);
    $trail = $plan->trail()->create(['codigo' => 'TR-AAA1']);
    $trail->topics()->create(['ordem' => 1, 'titulo' => 't', 'resumo' => 's']);
    $quiz = $trail->quiz()->create([]);
    $q = $quiz->questions()->create(['enunciado' => 'q', 'opcoes' => ['a', 'b'], 'correta' => 1]);

    $attempt = $trail->attempts()->create(['nome_aluno' => 'Ana', 'pontos' => 0]);
    $attempt->answers()->create(['quiz_question_id' => $q->id, 'escolhida' => 1, 'correta' => true]);

    expect($plan->trail->codigo)->toBe('TR-AAA1')
        ->and($trail->topics)->toHaveCount(1)
        ->and($quiz->questions->first()->opcoes)->toBe(['a', 'b'])
        ->and($attempt->answers->first()->correta)->toBeTrue();
});
