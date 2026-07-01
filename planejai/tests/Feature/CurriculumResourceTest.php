<?php

use App\Models\{CurriculumUnit, User};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('professor autenticado acessa a listagem do currículo', function () {
    $user = User::factory()->create();
    CurriculumUnit::create([
        'etapa' => 'Ensino Médio', 'serie' => '1ª série', 'componente' => 'Educação Financeira',
        'origem' => 'extra', 'trimestre' => 1, 'objetos' => ['Orçamento pessoal', 'Juros'],
    ]);

    $this->actingAs($user)
        ->get('/admin/curriculum-units')
        ->assertOk()
        ->assertSee('Educação Financeira');
});
