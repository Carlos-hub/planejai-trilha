<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('trail_topics', function (Blueprint $table) {
            $table->id();
            $table->foreignId('study_trail_id')->constrained()->cascadeOnDelete();
            $table->unsignedSmallInteger('ordem');
            $table->string('titulo');
            $table->text('resumo');
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('trail_topics');
    }
};
