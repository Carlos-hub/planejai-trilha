<?php

namespace App\Filament\Resources\LessonPlans;

use App\Filament\Resources\LessonPlans\Pages\CreateLessonPlan;
use App\Filament\Resources\LessonPlans\Pages\EditLessonPlan;
use App\Filament\Resources\LessonPlans\Pages\ListLessonPlans;
use App\Filament\Resources\LessonPlans\Pages\Turma;
use App\Filament\Resources\LessonPlans\Schemas\LessonPlanForm;
use App\Filament\Resources\LessonPlans\Tables\LessonPlansTable;
use App\Models\LessonPlan;
use BackedEnum;
use Filament\Resources\Resource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class LessonPlanResource extends Resource
{
    protected static ?string $model = LessonPlan::class;

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    public static function form(Schema $schema): Schema
    {
        return LessonPlanForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return LessonPlansTable::configure($table);
    }

    public static function getRelations(): array
    {
        return [
            //
        ];
    }

    public static function getPages(): array
    {
        return [
            'index' => ListLessonPlans::route('/'),
            'create' => CreateLessonPlan::route('/create'),
            'edit' => EditLessonPlan::route('/{record}/edit'),
            'turma' => Turma::route('/{record}/turma'),
        ];
    }
}
