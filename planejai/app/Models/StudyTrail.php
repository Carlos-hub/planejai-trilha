<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Str;

class StudyTrail extends Model
{
    protected $fillable = ['lesson_plan_id', 'codigo', 'publicada_em'];

    protected $casts = ['publicada_em' => 'datetime'];

    public function lessonPlan()
    {
        return $this->belongsTo(LessonPlan::class);
    }

    public function topics()
    {
        return $this->hasMany(TrailTopic::class)->orderBy('ordem');
    }

    public function quiz()
    {
        return $this->hasOne(Quiz::class);
    }

    public function attempts()
    {
        return $this->hasMany(StudentAttempt::class);
    }

    public static function gerarCodigo(): string
    {
        do {
            $codigo = 'TR-' . strtoupper(Str::random(4));
        } while (self::where('codigo', $codigo)->exists());

        return $codigo;
    }

    public function publicar(): void
    {
        $this->update(['publicada_em' => now()]);
    }

    public function scopePublished($query)
    {
        return $query->whereNotNull('publicada_em');
    }

    public function estatisticas(): array
    {
        return [
            'total_alunos' => $this->attempts()->count(),
            'media_pontos' => round((float) $this->attempts()->avg('pontos'), 1),
            'concluidos' => $this->attempts()->whereNotNull('concluido_em')->count(),
        ];
    }
}
