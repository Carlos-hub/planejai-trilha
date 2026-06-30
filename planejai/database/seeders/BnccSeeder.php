<?php

namespace Database\Seeders;

use App\Models\BnccSkill;
use Illuminate\Database\Seeder;

class BnccSeeder extends Seeder
{
    public function run(): void
    {
        $path = database_path('seeds/bncc.json');
        $rows = json_decode(file_get_contents($path), true);

        foreach ($rows as $row) {
            BnccSkill::updateOrCreate(['code' => $row['code']], $row);
        }
    }
}
