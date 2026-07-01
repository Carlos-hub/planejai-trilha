<?php

namespace App\Filament\Resources\CurriculumUnits\Pages;

use App\Filament\Resources\CurriculumUnits\CurriculumUnitResource;
use Filament\Resources\Pages\ListRecords;
use Filament\Schemas\Components\Tabs\Tab;
use Illuminate\Database\Eloquent\Builder;

class ListCurriculumUnits extends ListRecords
{
    protected static string $resource = CurriculumUnitResource::class;

    protected function getHeaderActions(): array
    {
        // currículo é somente leitura: nenhuma ação de criação
        return [];
    }

    public function getTabs(): array
    {
        return [
            'fundamental' => Tab::make('Ensino Fundamental')
                ->modifyQueryUsing(fn (Builder $query) => $query->where('etapa', 'like', 'Ensino Fundamental%')),
            'medio' => Tab::make('Ensino Médio')
                ->modifyQueryUsing(fn (Builder $query) => $query->where('etapa', 'Ensino Médio')),
            'todas' => Tab::make('Todas'),
        ];
    }

    public function getDefaultActiveTab(): string|int|null
    {
        return 'fundamental';
    }
}
