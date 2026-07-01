<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class LessonPlan extends Model
{
    protected $fillable = ['user_id', 'bncc_skill_id', 'curriculum_unit_id', 'titulo', 'duracao_min', 'objetivos', 'metodologia', 'recursos', 'avaliacao', 'atividade', 'status'];

    public function bnccSkill()
    {
        return $this->belongsTo(BnccSkill::class);
    }

    public function curriculumUnit()
    {
        return $this->belongsTo(CurriculumUnit::class);
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
