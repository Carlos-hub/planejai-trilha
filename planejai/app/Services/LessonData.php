<?php

namespace App\Services;

class LessonData
{
    public function __construct(
        public array $plano,
        public string $atividade,
        public array $topicos,
        public array $questoes,
    ) {}

    public static function fromArray(array $d): self
    {
        return new self(
            plano: $d['plano'],
            atividade: $d['atividade'],
            topicos: $d['trilha']['topicos'],
            questoes: $d['trilha']['quiz']['questoes'],
        );
    }
}
