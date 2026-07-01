<?php

namespace App\Filament\Resources\CurriculumUnits;

use App\Filament\Resources\CurriculumUnits\Pages\ListCurriculumUnits;
use App\Filament\Resources\CurriculumUnits\Schemas\CurriculumUnitForm;
use App\Filament\Resources\CurriculumUnits\Tables\CurriculumUnitsTable;
use App\Models\CurriculumUnit;
use BackedEnum;
use Filament\Resources\Resource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class CurriculumUnitResource extends Resource
{
    protected static ?string $model = CurriculumUnit::class;

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBookOpen;

    protected static ?string $navigationLabel = 'Currículo (BNCC + Extras)';

    protected static ?string $modelLabel = 'unidade curricular';

    protected static ?string $pluralModelLabel = 'currículo';

    protected static ?string $recordTitleAttribute = 'componente';

    public static function form(Schema $schema): Schema
    {
        return CurriculumUnitForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return CurriculumUnitsTable::configure($table);
    }

    public static function getRelations(): array
    {
        return [
            //
        ];
    }

    // Currículo é somente leitura: não pode ser criado, editado ou excluído.
    // O professor apenas duplica uma unidade para um plano de aula.
    public static function canCreate(): bool
    {
        return false;
    }

    public static function canEdit($record): bool
    {
        return false;
    }

    public static function canDelete($record): bool
    {
        return false;
    }

    public static function canDeleteAny(): bool
    {
        return false;
    }

    public static function getPages(): array
    {
        return [
            'index' => ListCurriculumUnits::route('/'),
        ];
    }
}
