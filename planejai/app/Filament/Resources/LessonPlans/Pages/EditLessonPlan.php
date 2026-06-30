<?php

namespace App\Filament\Resources\LessonPlans\Pages;

use App\Actions\GenerateLessonAction;
use App\Filament\Resources\LessonPlans\LessonPlanResource;
use Filament\Actions\Action;
use Filament\Actions\DeleteAction;
use Filament\Notifications\Notification;
use Filament\Resources\Pages\EditRecord;

class EditLessonPlan extends EditRecord
{
    protected static string $resource = LessonPlanResource::class;

    protected function getHeaderActions(): array
    {
        return [
            Action::make('gerar')
                ->label('Gerar com IA')
                ->icon('heroicon-o-sparkles')
                ->color('primary')
                ->requiresConfirmation()
                ->action(function () {
                    app(GenerateLessonAction::class)->execute($this->record);
                    $this->record->refresh();
                    $this->fillForm();

                    $this->record->status === 'gerado'
                        ? Notification::make()->title('Plano gerado com sucesso')->success()->send()
                        : Notification::make()->title('Falha ao gerar o plano')->danger()->send();
                }),
            Action::make('publicar')
                ->label('Publicar trilha')
                ->icon('heroicon-o-globe-alt')
                ->color('success')
                ->visible(fn () => $this->record->trail && ! $this->record->trail->publicada_em)
                ->action(function () {
                    $this->record->trail->publicar();
                    Notification::make()->title('Trilha publicada')->success()->send();
                }),
            Action::make('turma')
                ->label('Ver turma')
                ->icon('heroicon-o-users')
                ->color('gray')
                ->visible(fn () => (bool) $this->record->trail)
                ->url(fn () => Turma::getUrl(['record' => $this->record])),
            DeleteAction::make(),
        ];
    }
}
