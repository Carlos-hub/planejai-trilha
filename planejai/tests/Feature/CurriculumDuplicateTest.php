<?php

use App\Filament\Resources\CurriculumUnits\CurriculumUnitResource;
use App\Filament\Resources\CurriculumUnits\Pages\ListCurriculumUnits;
use App\Models\{CurriculumUnit, LessonPlan, User};
use Illuminate\Foundation\Testing\RefreshDatabase;
use Livewire\Livewire;

uses(RefreshDatabase::class);

it('não permite editar, criar ou excluir o currículo', function () {
    expect(CurriculumUnitResource::canCreate())->toBeFalse()
        ->and(CurriculumUnitResource::canEdit(new CurriculumUnit()))->toBeFalse()
        ->and(CurriculumUnitResource::canDelete(new CurriculumUnit()))->toBeFalse();
});

it('duplica uma unidade do currículo para um plano de aula, mantendo o currículo intacto', function () {
    $user = User::factory()->create();
    $unit = CurriculumUnit::create([
        'etapa' => 'Ensino Fundamental - Anos Finais', 'serie' => '6º ano', 'componente' => 'Educação Financeira',
        'origem' => 'extra', 'trimestre' => 1, 'objetos' => ['Orçamento pessoal', 'Juros e inflação'],
    ]);

    $this->actingAs($user);

    Livewire::test(ListCurriculumUnits::class)
        ->callTableAction('duplicar', $unit);

    // plano criado a partir do currículo
    $plan = LessonPlan::where('curriculum_unit_id', $unit->id)->first();
    expect($plan)->not->toBeNull()
        ->and($plan->user_id)->toBe($user->id)
        ->and($plan->status)->toBe('rascunho')
        ->and($plan->bncc_skill_id)->toBeNull()
        ->and($plan->objetivos)->toContain('Orçamento pessoal');

    // currículo permanece intacto (nada alterado, nada removido)
    $unit->refresh();
    expect(CurriculumUnit::count())->toBe(1)
        ->and($unit->componente)->toBe('Educação Financeira')
        ->and($unit->objetos)->toBe(['Orçamento pessoal', 'Juros e inflação']);
});
