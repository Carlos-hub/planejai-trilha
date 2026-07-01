<?php

namespace App\Filament\Resources\CurriculumUnits\Tables;

use App\Filament\Resources\LessonPlans\Pages\EditLessonPlan;
use App\Models\CurriculumUnit;
use App\Models\LessonPlan;
use Filament\Actions\Action;
use Filament\Actions\ViewAction;
use Filament\Notifications\Notification;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;
use Illuminate\Support\Facades\Auth;

class CurriculumUnitsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->defaultSort('serie')
            ->columns([
                TextColumn::make('serie')->label('Série')->sortable()->searchable(),
                TextColumn::make('componente')->label('Matéria')->sortable()->searchable(),
                TextColumn::make('origem')->label('Origem')->badge()->color(fn (string $state) => $state === 'bncc' ? 'info' : 'warning')
                    ->formatStateUsing(fn (string $state) => $state === 'bncc' ? 'BNCC' : 'Extra'),
                TextColumn::make('trimestre')->label('Trim.')->badge()->sortable(),
                TextColumn::make('objetos')->label('Objetos de conhecimento')
                    ->formatStateUsing(fn ($state) => is_array($state) ? implode(' · ', $state) : $state)
                    ->wrap()->limit(120),
                TextColumn::make('etapa')->label('Etapa')->toggleable(isToggledHiddenByDefault: true),
            ])
            ->filters([
                SelectFilter::make('origem')->label('Origem')
                    ->options(['bncc' => 'BNCC', 'extra' => 'Extra']),
                SelectFilter::make('etapa')->label('Etapa')
                    ->options(fn () => CurriculumUnit::query()->distinct()->pluck('etapa', 'etapa')),
                SelectFilter::make('serie')->label('Série')
                    ->options(fn () => CurriculumUnit::query()->distinct()->orderBy('serie')->pluck('serie', 'serie')),
                SelectFilter::make('componente')->label('Matéria')
                    ->options(fn () => CurriculumUnit::query()->distinct()->orderBy('componente')->pluck('componente', 'componente')),
            ])
            ->recordActions([
                ViewAction::make()
                    ->label('Ver'),
                Action::make('duplicar')
                    ->label('Duplicar para plano')
                    ->icon('heroicon-o-document-duplicate')
                    ->color('primary')
                    ->requiresConfirmation()
                    ->modalHeading('Duplicar para plano de aula')
                    ->modalDescription('Cria um novo plano de aula a partir desta unidade do currículo. O currículo permanece intacto.')
                    ->action(function (CurriculumUnit $record) {
                        $plan = LessonPlan::create([
                            'user_id' => Auth::id(),
                            'curriculum_unit_id' => $record->id,
                            'titulo' => "{$record->componente} — {$record->serie} ({$record->trimestre}º tri)",
                            'duracao_min' => 50,
                            'objetivos' => 'Conteúdos do currículo: ' . implode('; ', $record->objetos),
                            'status' => 'rascunho',
                        ]);

                        Notification::make()
                            ->title('Plano de aula criado a partir do currículo')
                            ->success()
                            ->send();

                        return redirect(EditLessonPlan::getUrl(['record' => $plan]));
                    }),
            ]);
        // sem toolbarActions/bulk delete: currículo é somente leitura
    }
}
