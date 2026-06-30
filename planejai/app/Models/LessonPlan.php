<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class LessonPlan extends Model
{
    protected $fillable = ['user_id', 'bncc_skill_id', 'duracao_min', 'objetivos', 'metodologia', 'recursos', 'avaliacao', 'atividade', 'status'];

    public function bnccSkill()
    {
        return $this->belongsTo(BnccSkill::class);
    }

    public function user()
    {
        return $this->belongsTo(User::class);
    }

    public function trail()
    {
        return $this->hasOne(StudyTrail::class);
    }
}
