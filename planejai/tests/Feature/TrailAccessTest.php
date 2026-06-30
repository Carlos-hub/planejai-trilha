<?php

use App\Models\{User, BnccSkill, LessonPlan, StudyTrail};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

function fazerTrilha(bool $publicada): StudyTrail {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-DEMO']);
    $trail->topics()->create(['ordem' => 1, 'titulo' => 'Tópico 1', 'resumo' => 'resumo aqui']);
    if ($publicada) $trail->publicar();
    return $trail;
}

test('trilha publicada é acessível por código', function () {
    fazerTrilha(true);
    $this->get('/t/TR-DEMO')->assertOk()->assertSee('Tópico 1');
});

test('trilha não publicada retorna 404', function () {
    fazerTrilha(false);
    $this->get('/t/TR-DEMO')->assertNotFound();
});
