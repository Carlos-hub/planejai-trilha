<?php

namespace App\Http\Controllers;

use App\Models\StudyTrail;

class TrailController extends Controller
{
    public function show(string $codigo)
    {
        $trail = StudyTrail::published()
            ->where('codigo', $codigo)
            ->with(['topics', 'quiz.questions', 'lessonPlan.bnccSkill'])
            ->firstOrFail();

        return view('trail.show', ['trail' => $trail]);
    }
}
