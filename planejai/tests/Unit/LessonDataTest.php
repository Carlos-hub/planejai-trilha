<?php

use App\Services\LessonData;

test('LessonData::fromArray mapeia o JSON da IA', function () {
    $raw = json_decode(file_get_contents(dirname(__DIR__) . '/fixtures/lesson_response.json'), true);

    $data = LessonData::fromArray($raw);

    expect($data)->toBeInstanceOf(LessonData::class)
        ->and($data->plano['objetivos'])->toBe('Compreender números racionais decimais.')
        ->and($data->atividade)->toContain('5 exercícios')
        ->and($data->topicos)->toHaveCount(2)
        ->and($data->topicos[0]['titulo'])->toBe('O que é número decimal')
        ->and($data->questoes[0]['correta'])->toBe(0)
        ->and($data->questoes[0]['opcoes'])->toBe(['0,5', '0,45']);
});
