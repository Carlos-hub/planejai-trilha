<?php

use App\Models\{User, BnccSkill, LessonPlan};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('estatísticas da turma somam tentativas', function () {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-TURMA']);
    $trail->attempts()->create(['nome_aluno' => 'A', 'pontos' => 20, 'concluido_em' => now()]);
    $trail->attempts()->create(['nome_aluno' => 'B', 'pontos' => 10, 'concluido_em' => now()]);

    $stats = $trail->estatisticas();

    expect($stats['total_alunos'])->toBe(2)
        ->and($stats['media_pontos'])->toBe(15.0)
        ->and($stats['concluidos'])->toBe(2);
});
