<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class QuizQuestion extends Model
{
    protected $fillable = ['quiz_id', 'enunciado', 'opcoes', 'correta'];

    protected $casts = ['opcoes' => 'array'];

    public function quiz()
    {
        return $this->belongsTo(Quiz::class);
    }
}
