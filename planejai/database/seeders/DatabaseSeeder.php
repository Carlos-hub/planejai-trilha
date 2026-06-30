<?php

namespace Database\Seeders;

use App\Models\User;
use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class DatabaseSeeder extends Seeder
{
    use WithoutModelEvents;

    /**
     * Seed the application's database.
     */
    public function run(): void
    {
        $this->call(BnccSeeder::class);

        User::updateOrCreate(
            ['email' => 'professor@planejai.test'],
            ['name' => 'Professor Demo', 'password' => Hash::make('password')],
        );
    }
}
