<?php

namespace Database\Seeders;

use App\Models\CurriculumUnit;
use Illuminate\Database\Seeder;

class CurriculumSeeder extends Seeder
{
    public function run(): void
    {
        $path = database_path('seeds/curriculum.json');
        $rows = json_decode(file_get_contents($path), true);

        foreach ($rows as $row) {
            CurriculumUnit::updateOrCreate(
                [
                    'serie' => $row['serie'],
                    'componente' => $row['componente'],
                    'trimestre' => $row['trimestre'],
                ],
                $row,
            );
        }
    }
}
