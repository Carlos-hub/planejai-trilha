<?php

namespace App\Contracts;

use App\Models\BnccSkill;
use App\Services\LessonData;

interface LessonGenerator
{
    public function generate(BnccSkill $skill, int $duracaoMin): LessonData;
}
