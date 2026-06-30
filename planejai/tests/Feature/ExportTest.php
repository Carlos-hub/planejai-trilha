<?php

use App\Models\{User, BnccSkill, LessonPlan};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('export PDF retorna application/pdf', function () {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-PDF']);
    $trail->topics()->create(['ordem' => 1, 'titulo' => 'T1', 'resumo' => 'r1']);
    $trail->publicar();

    $resp = $this->get('/t/TR-PDF/pdf');
    $resp->assertOk();
    expect($resp->headers->get('content-type'))->toContain('application/pdf');
});
