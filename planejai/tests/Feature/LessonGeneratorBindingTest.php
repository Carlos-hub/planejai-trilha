<?php

use App\Contracts\LessonGenerator;
use App\Services\PrismLessonGenerator;

test('LessonGenerator resolve para o adapter configurado', function () {
    config()->set('llm.provider', 'anthropic');
    config()->set('llm.model', 'claude-opus-4-8');

    $service = app(LessonGenerator::class);

    expect($service)->toBeInstanceOf(PrismLessonGenerator::class);
});
