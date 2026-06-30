<?php

namespace App\Services;

use App\Contracts\LessonGenerator;
use App\Models\BnccSkill;
use Prism\Prism\Prism;
use Prism\Prism\Schema\ArraySchema;
use Prism\Prism\Schema\NumberSchema;
use Prism\Prism\Schema\ObjectSchema;
use Prism\Prism\Schema\StringSchema;

class PrismLessonGenerator implements LessonGenerator
{
    public function __construct(
        private string $provider,
        private string $model,
    ) {}

    public function generate(BnccSkill $skill, int $duracaoMin): LessonData
    {
        $response = Prism::structured()
            ->using($this->provider, $this->model)
            ->withSchema($this->schema())
            ->withPrompt($this->buildPrompt($skill, $duracaoMin))
            ->asStructured();

        return LessonData::fromArray($response->structured);
    }

    private function buildPrompt(BnccSkill $skill, int $duracaoMin): string
    {
        return "Você é um planejador pedagógico alinhado à BNCC. "
            . "Gere um plano de aula e uma trilha de estudo do aluno.\n\n"
            . "Disciplina: {$skill->disciplina}\n"
            . "Ano/série: {$skill->ano}\n"
            . "Habilidade BNCC: {$skill->code} — {$skill->descricao}\n"
            . "Duração da aula: {$duracaoMin} minutos\n\n"
            . "A trilha deve ter 2 a 4 tópicos curtos e um quiz de 3 a 5 questões "
            . "de múltipla escolha (índice 0-based da opção correta em 'correta').";
    }

    private function schema(): ObjectSchema
    {
        $plano = new ObjectSchema('plano', 'Plano de aula', [
            new StringSchema('objetivos', 'Objetivos da aula'),
            new StringSchema('metodologia', 'Metodologia'),
            new StringSchema('recursos', 'Recursos necessários'),
            new StringSchema('avaliacao', 'Forma de avaliação'),
        ], ['objetivos', 'metodologia', 'recursos', 'avaliacao']);

        $topico = new ObjectSchema('topico', 'Tópico da trilha', [
            new StringSchema('titulo', 'Título do tópico'),
            new StringSchema('resumo', 'Resumo curto'),
        ], ['titulo', 'resumo']);

        $questao = new ObjectSchema('questao', 'Questão do quiz', [
            new StringSchema('enunciado', 'Enunciado'),
            new ArraySchema('opcoes', 'Opções de resposta', new StringSchema('opcao', 'Opção')),
            new NumberSchema('correta', 'Índice 0-based da opção correta'),
        ], ['enunciado', 'opcoes', 'correta']);

        $quiz = new ObjectSchema('quiz', 'Quiz da trilha', [
            new ArraySchema('questoes', 'Questões', $questao),
        ], ['questoes']);

        $trilha = new ObjectSchema('trilha', 'Trilha do aluno', [
            new ArraySchema('topicos', 'Tópicos', $topico),
            $quiz,
        ], ['topicos', 'quiz']);

        return new ObjectSchema('lesson', 'Plano de aula + trilha', [
            $plano,
            new StringSchema('atividade', 'Atividade/quiz da aula'),
            $trilha,
        ], ['plano', 'atividade', 'trilha']);
    }
}
