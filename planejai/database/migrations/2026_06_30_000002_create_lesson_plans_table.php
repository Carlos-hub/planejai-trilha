<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('lesson_plans', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->foreignId('bncc_skill_id')->constrained('bncc_skills');
            $table->unsignedSmallInteger('duracao_min');
            $table->text('objetivos')->nullable();
            $table->text('metodologia')->nullable();
            $table->text('recursos')->nullable();
            $table->text('avaliacao')->nullable();
            $table->text('atividade')->nullable();
            $table->string('status')->default('rascunho');
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('lesson_plans');
    }
};
