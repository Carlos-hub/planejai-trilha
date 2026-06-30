<?php

use App\Actions\GenerateLessonAction;
use App\Contracts\LessonGenerator;
use App\Models\{User, BnccSkill, LessonPlan};
use App\Services\LessonData;
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('ação gera plano, trilha, tópicos e quiz', function () {
    $data = new LessonData(
        plano: ['objetivos' => 'o', 'metodologia' => 'm', 'recursos' => 'r', 'avaliacao' => 'a'],
        atividade: 'atividade x',
        topicos: [['titulo' => 't1', 'resumo' => 's1'], ['titulo' => 't2', 'resumo' => 's2']],
        questoes: [['enunciado' => 'q', 'opcoes' => ['a', 'b'], 'correta' => 1]],
    );
    $ai = Mockery::mock(LessonGenerator::class);
    $ai->shouldReceive('generate')->once()->andReturn($data);
    app()->instance(LessonGenerator::class, $ai);

    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create([
        'user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'rascunho',
    ]);

    app(GenerateLessonAction::class)->execute($plan);
    $plan->refresh();

    expect($plan->status)->toBe('gerado')
        ->and($plan->objetivos)->toBe('o')
        ->and($plan->atividade)->toBe('atividade x')
        ->and($plan->trail)->not->toBeNull()
        ->and($plan->trail->codigo)->toStartWith('TR-')
        ->and($plan->trail->topics)->toHaveCount(2)
        ->and($plan->trail->quiz->questions)->toHaveCount(1)
        ->and($plan->trail->quiz->questions->first()->correta)->toBe(1);
});
