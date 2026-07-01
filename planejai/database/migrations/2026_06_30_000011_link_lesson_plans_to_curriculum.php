<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('lesson_plans', function (Blueprint $table) {
            // plano pode nascer de uma habilidade BNCC OU de uma unidade do currículo
            $table->foreignId('bncc_skill_id')->nullable()->change();
            $table->foreignId('curriculum_unit_id')->nullable()->after('bncc_skill_id')
                ->constrained('curriculum_units')->nullOnDelete();
            $table->string('titulo')->nullable()->after('curriculum_unit_id');
        });
    }

    public function down(): void
    {
        Schema::table('lesson_plans', function (Blueprint $table) {
            $table->dropConstrainedForeignId('curriculum_unit_id');
            $table->dropColumn('titulo');
            $table->foreignId('bncc_skill_id')->nullable(false)->change();
        });
    }
};
