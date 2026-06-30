<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class TrailTopic extends Model
{
    protected $fillable = ['study_trail_id', 'ordem', 'titulo', 'resumo'];

    public function trail()
    {
        return $this->belongsTo(StudyTrail::class, 'study_trail_id');
    }
}
