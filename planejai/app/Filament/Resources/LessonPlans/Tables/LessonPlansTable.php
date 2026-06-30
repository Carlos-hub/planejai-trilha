<?php

namespace App\Filament\Resources\LessonPlans\Tables;

use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;

class LessonPlansTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('bnccSkill.code')->label('BNCC')->searchable(),
                TextColumn::make('bnccSkill.disciplina')->label('Disciplina'),
                TextColumn::make('duracao_min')->label('Duração'),
                TextColumn::make('status')->label('Status')->badge()->color(fn (string $state) => match ($state) {
                    'gerado' => 'success',
                    'falha' => 'danger',
                    default => 'gray',
                }),
                TextColumn::make('trail.codigo')->label('Trilha')->placeholder('—'),
            ])
            ->filters([
                //
            ])
            ->recordActions([
                EditAction::make(),
            ])
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                ]),
            ]);
    }
}
