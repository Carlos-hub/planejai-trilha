<?php

namespace App\Filament\Resources\LessonPlans\Pages;

use App\Filament\Resources\LessonPlans\LessonPlanResource;
use Filament\Resources\Pages\Page;
use Filament\Resources\Pages\Concerns\InteractsWithRecord;

class Turma extends Page
{
    use InteractsWithRecord;

    protected static string $resource = LessonPlanResource::class;

    protected string $view = 'filament.resources.lesson-plans.pages.turma';

    protected static ?string $title = 'Turma';

    public function mount(int|string $record): void
    {
        $this->record = $this->resolveRecord($record);
    }

    public function getStats(): array
    {
        $trail = $this->record->trail;

        return $trail
            ? $trail->estatisticas()
            : ['total_alunos' => 0, 'media_pontos' => 0.0, 'concluidos' => 0];
    }

    public function getAttempts()
    {
        return $this->record->trail
            ? $this->record->trail->attempts()->latest()->get()
            : collect();
    }
}
