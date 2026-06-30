<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class AttemptAnswer extends Model
{
    protected $fillable = ['student_attempt_id', 'quiz_question_id', 'escolhida', 'correta'];

    protected $casts = ['correta' => 'boolean'];
}
