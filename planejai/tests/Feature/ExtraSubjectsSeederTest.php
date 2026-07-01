<?php

use App\Models\CurriculumUnit;
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('seeder carrega matérias extras (não-BNCC) marcadas como origem=extra', function () {
    $this->seed(\Database\Seeders\ExtraSubjectsSeeder::class);

    // todas as linhas inseridas aqui são extras
    expect(CurriculumUnit::where('origem', 'extra')->count())->toBeGreaterThan(0)
        ->and(CurriculumUnit::where('origem', 'bncc')->count())->toBe(0);

    // Educação Financeira presente nas 12 séries
    expect(CurriculumUnit::where('componente', 'Educação Financeira')->distinct('serie')->count('serie'))->toBe(12);

    // exemplo: Finanças no 6º ano traz orçamento
    $fin6 = CurriculumUnit::where('serie', '6º ano')
        ->where('componente', 'Educação Financeira')
        ->where('trimestre', 1)
        ->first();

    expect($fin6)->not->toBeNull()
        ->and($fin6->origem)->toBe('extra')
        ->and(implode(' ', $fin6->objetos))->toContain('Orçamento');

    // demais extras existem
    foreach (['Projeto de Vida', 'Empreendedorismo', 'Educação Digital e Tecnologia', 'Cidadania, Ética e Direitos Humanos'] as $comp) {
        expect(CurriculumUnit::where('componente', $comp)->count())->toBeGreaterThan(0);
    }
});
