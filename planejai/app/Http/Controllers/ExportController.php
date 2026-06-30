<?php

namespace App\Http\Controllers;

use App\Models\StudyTrail;
use Barryvdh\DomPDF\Facade\Pdf;

class ExportController extends Controller
{
    public function pdf(string $codigo)
    {
        $trail = StudyTrail::published()->where('codigo', $codigo)
            ->with(['topics', 'lessonPlan.bnccSkill'])->firstOrFail();

        return Pdf::loadView('trail.pdf', ['trail' => $trail])
            ->download("trilha-{$codigo}.pdf");
    }
}
