<?php

namespace App\Filament\Resources\LessonPlans\Pages;

use App\Filament\Resources\LessonPlans\LessonPlanResource;
use Filament\Resources\Pages\CreateRecord;

class CreateLessonPlan extends CreateRecord
{
    protected static string $resource = LessonPlanResource::class;
}
