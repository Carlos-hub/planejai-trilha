<?php

namespace App\Filament\Resources\CurriculumUnits\Schemas;

use App\Models\CurriculumUnit;
use Filament\Forms\Components\Repeater;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Schema;

class CurriculumUnitForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Select::make('etapa')
                    ->label('Etapa')
                    ->options([
                        'Ensino Fundamental - Anos Iniciais' => 'Ensino Fundamental - Anos Iniciais',
                        'Ensino Fundamental - Anos Finais' => 'Ensino Fundamental - Anos Finais',
                        'Ensino Médio' => 'Ensino Médio',
                    ])
                    ->required(),
                Select::make('serie')
                    ->label('Série')
                    ->options(array_combine(
                        $series = ['1º ano', '2º ano', '3º ano', '4º ano', '5º ano', '6º ano', '7º ano', '8º ano', '9º ano', '1ª série', '2ª série', '3ª série'],
                        $series,
                    ))
                    ->required(),
                Select::make('componente')
                    ->label('Componente / Matéria')
                    ->options(fn () => CurriculumUnit::query()
                        ->distinct()
                        ->orderBy('componente')
                        ->pluck('componente', 'componente'))
                    ->searchable()
                    ->createOptionForm([
                        TextInput::make('componente')->label('Nova matéria')->required(),
                    ])
                    ->createOptionUsing(fn (array $data) => $data['componente'])
                    ->required(),
                Select::make('origem')
                    ->label('Origem')
                    ->options(['bncc' => 'BNCC', 'extra' => 'Extra (fora da BNCC)'])
                    ->default('extra')
                    ->required(),
                Select::make('trimestre')
                    ->label('Trimestre')
                    ->options([1 => '1º trimestre', 2 => '2º trimestre', 3 => '3º trimestre'])
                    ->required(),
                Repeater::make('objetos')
                    ->label('Objetos de conhecimento / conteúdos')
                    ->simple(
                        TextInput::make('item')->required(),
                    )
                    ->minItems(1)
                    ->defaultItems(1)
                    ->columnSpanFull(),
            ]);
    }
}
