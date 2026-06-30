<?php

namespace App\Actions;

use App\Contracts\LessonGenerator;
use App\Models\LessonPlan;
use App\Models\StudyTrail;
use Illuminate\Support\Facades\DB;
use Throwable;

class GenerateLessonAction
{
    public function __construct(private LessonGenerator $ai) {}

    public function execute(LessonPlan $plan): void
    {
        try {
            $data = $this->ai->generate($plan->bnccSkill, $plan->duracao_min);

            DB::transaction(function () use ($plan, $data) {
                $plan->update([
                    'objetivos' => $data->plano['objetivos'],
                    'metodologia' => $data->plano['metodologia'],
                    'recursos' => $data->plano['recursos'],
                    'avaliacao' => $data->plano['avaliacao'],
                    'atividade' => $data->atividade,
                    'status' => 'gerado',
                ]);

                $trail = $plan->trail()->create(['codigo' => StudyTrail::gerarCodigo()]);

                foreach ($data->topicos as $i => $t) {
                    $trail->topics()->create([
                        'ordem' => $i + 1,
                        'titulo' => $t['titulo'],
                        'resumo' => $t['resumo'],
                    ]);
                }

                $quiz = $trail->quiz()->create([]);
                foreach ($data->questoes as $q) {
                    $quiz->questions()->create([
                        'enunciado' => $q['enunciado'],
                        'opcoes' => $q['opcoes'],
                        'correta' => (int) $q['correta'],
                    ]);
                }
            });
        } catch (Throwable $e) {
            $plan->update(['status' => 'falha']);
            report($e);
        }
    }
}
