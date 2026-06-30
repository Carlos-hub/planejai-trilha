<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Quiz extends Model
{
    protected $fillable = ['study_trail_id'];

    public function questions()
    {
        return $this->hasMany(QuizQuestion::class);
    }

    public function trail()
    {
        return $this->belongsTo(StudyTrail::class, 'study_trail_id');
    }
}
