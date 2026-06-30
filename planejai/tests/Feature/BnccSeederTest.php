<?php

use App\Models\BnccSkill;
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('seeder carrega habilidades BNCC do JSON', function () {
    $this->seed(\Database\Seeders\BnccSeeder::class);

    expect(BnccSkill::count())->toBeGreaterThan(0);

    $skill = BnccSkill::where('code', 'EF06MA01')->first();
    expect($skill)->not->toBeNull()
        ->and($skill->disciplina)->toBe('Matemática')
        ->and($skill->ano)->toBe('6º ano');
});
