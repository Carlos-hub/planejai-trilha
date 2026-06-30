<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bncc_skills', function (Blueprint $table) {
            $table->id();
            $table->string('code')->unique();      // ex: EF06MA01
            $table->string('disciplina');          // ex: Matemática
            $table->string('ano');                 // ex: 6º ano
            $table->text('descricao');
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bncc_skills');
    }
};
