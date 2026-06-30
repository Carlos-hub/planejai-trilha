<?php

namespace App\Filament\Resources\LessonPlans\Schemas;

use App\Models\BnccSkill;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Schema;

class LessonPlanForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Select::make('bncc_skill_id')
                    ->label('Habilidade BNCC')
                    ->options(fn () => BnccSkill::all()
                        ->mapWithKeys(fn ($s) => [$s->id => "{$s->code} — {$s->disciplina} ({$s->ano})"]))
                    ->searchable()
                    ->required(),
                TextInput::make('duracao_min')
                    ->label('Duração (min)')
                    ->numeric()
                    ->minValue(10)
                    ->maxValue(300)
                    ->required(),
                TextInput::make('status')
                    ->label('Status')
                    ->disabled()
                    ->dehydrated(false),
                Textarea::make('objetivos')->label('Objetivos')->rows(3)->columnSpanFull(),
                Textarea::make('metodologia')->label('Metodologia')->rows(3)->columnSpanFull(),
                Textarea::make('recursos')->label('Recursos')->rows(2)->columnSpanFull(),
                Textarea::make('avaliacao')->label('Avaliação')->rows(2)->columnSpanFull(),
                Textarea::make('atividade')->label('Atividade da aula')->rows(3)->columnSpanFull(),
            ]);
    }
}
