<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('curriculum_units', function (Blueprint $table) {
            $table->id();
            $table->string('etapa');        // ex: Ensino Fundamental - Anos Iniciais
            $table->string('serie');        // ex: 1º ano
            $table->string('componente');   // ex: Matemática
            $table->unsignedTinyInteger('trimestre'); // 1, 2 ou 3
            $table->json('objetos');        // lista de objetos de conhecimento / conteúdos
            $table->timestamps();

            $table->unique(['serie', 'componente', 'trimestre']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('curriculum_units');
    }
};
