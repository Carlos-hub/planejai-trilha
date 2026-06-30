<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class StudentAttempt extends Model
{
    protected $fillable = ['study_trail_id', 'nome_aluno', 'pontos', 'concluido_em'];

    protected $casts = ['concluido_em' => 'datetime'];

    public function trail()
    {
        return $this->belongsTo(StudyTrail::class, 'study_trail_id');
    }

    public function answers()
    {
        return $this->hasMany(AttemptAnswer::class);
    }
}
