<?php

use App\Models\CurriculumUnit;
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('seeder carrega o currículo BNCC por série, componente e trimestre', function () {
    $this->seed(\Database\Seeders\CurriculumSeeder::class);

    // 12 séries, todas com 3 trimestres
    expect(CurriculumUnit::distinct('serie')->count('serie'))->toBe(12)
        ->and(CurriculumUnit::where('trimestre', '>', 3)->count())->toBe(0)
        ->and(CurriculumUnit::where('trimestre', '<', 1)->count())->toBe(0);

    // exemplo do enunciado: Matemática 1º ano traz contagem/agrupamento/estimativa
    $mat1 = CurriculumUnit::where('serie', '1º ano')
        ->where('componente', 'Matemática')
        ->where('trimestre', 1)
        ->first();

    expect($mat1)->not->toBeNull()
        ->and($mat1->etapa)->toBe('Ensino Fundamental - Anos Iniciais')
        ->and(implode(' ', $mat1->objetos))->toContain('contagem');

    // Ensino Médio inclui componentes específicos (Física só no EM)
    expect(CurriculumUnit::where('componente', 'Física')->where('etapa', 'Ensino Médio')->count())->toBeGreaterThan(0)
        ->and(CurriculumUnit::where('serie', '3ª série')->distinct('componente')->count('componente'))->toBe(12);
});
