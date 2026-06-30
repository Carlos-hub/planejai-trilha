# PlanejAI + Trilha Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Construir MVP de planejador de aulas BNCC (professor) que deriva trilhas de estudo do aluno (acesso por código, quiz autocorrigido, gamificação) a partir de uma única geração via Anthropic API.

**Architecture:** Laravel monólito modular. Filament para o painel do professor (auth, CRUD, ação de gerar, dashboard). Blade para páginas públicas do aluno (trilha por código + quiz). Um `AiService` encapsula o Anthropic PHP SDK e é o único ponto de IA. BNCC carregada como seed JSON.

**Tech Stack:** Laravel 11, PHP 8.2+, PostgreSQL, Filament 3, Blade, Pest (testes), `anthropic-ai/sdk`, `barryvdh/laravel-dompdf` (export).

## Global Constraints

- Modelo Anthropic: `claude-opus-4-8` (exato, sem sufixo de data).
- Chamada de IA usa structured output (`outputConfig.format` com `json_schema`) — **nunca** assistant prefill (retorna 400 no Opus 4.8).
- API key da Anthropic via env `ANTHROPIC_API_KEY`, nunca hardcoded.
- Acesso do aluno sem login: código curto da trilha + nome digitado. Sem tabela de aluno.
- BNCC vem de seed JSON estático (`database/seeds/bncc.json`), sem chamada externa em runtime.
- Todo teste que toca a Anthropic usa client mockado — nenhum teste faz chamada real à API.
- TDD: teste falhando antes da implementação. Commits frequentes.

---

### Task 1: Scaffold do projeto Laravel + dependências

**Files:**
- Create: app Laravel completo na raiz do repo
- Modify: `.env.example`, `config/services.php`
- Create: `composer.json` (via installer)

**Interfaces:**
- Produces: app Laravel rodável; `config('services.anthropic.key')` e `config('services.anthropic.model')` disponíveis.

- [ ] **Step 1: Instalar Laravel na raiz**

O diretório já contém `docs/` e `spec/`. Instalar Laravel em pasta temporária e mover, para não sobrescrever.

```bash
cd /home/carlos/programming/FIAP/Hackathon
composer create-project laravel/laravel _laravel_tmp
# mover conteúdo (incluindo ocultos) para a raiz, sem sobrescrever docs/ spec/ .git/
shopt -s dotglob
cp -rn _laravel_tmp/* .
rm -rf _laravel_tmp
shopt -u dotglob
```

- [ ] **Step 2: Verificar que a app sobe**

Run: `php artisan --version`
Expected: imprime `Laravel Framework 11.x` (ou superior).

- [ ] **Step 3: Instalar dependências do projeto**

```bash
composer require anthropic-ai/sdk barryvdh/laravel-dompdf
composer require --dev pestphp/pest pestphp/pest-plugin-laravel
php artisan pest:install
```

- [ ] **Step 4: Configurar Postgres e Anthropic no env**

Editar `.env` (e replicar chaves em `.env.example` sem valores secretos):

```dotenv
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=planejai
DB_USERNAME=postgres
DB_PASSWORD=

ANTHROPIC_API_KEY=
ANTHROPIC_MODEL=claude-opus-4-8
```

- [ ] **Step 5: Adicionar config de serviço Anthropic**

Em `config/services.php`, adicionar ao array retornado:

```php
'anthropic' => [
    'key' => env('ANTHROPIC_API_KEY'),
    'model' => env('ANTHROPIC_MODEL', 'claude-opus-4-8'),
],
```

- [ ] **Step 6: Criar o banco e rodar migrations base**

```bash
createdb planejai 2>/dev/null || true
php artisan migrate
```

Expected: tabelas default do Laravel criadas sem erro.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "chore: scaffold Laravel app with Anthropic SDK and Pest"
```

---

### Task 2: Seed da BNCC (modelo + migration + dados)

**Files:**
- Create: `database/migrations/xxxx_create_bncc_skills_table.php`
- Create: `app/Models/BnccSkill.php`
- Create: `database/seeds/bncc.json`
- Create: `database/seeders/BnccSeeder.php`
- Modify: `database/seeders/DatabaseSeeder.php`
- Test: `tests/Feature/BnccSeederTest.php`

**Interfaces:**
- Produces: `BnccSkill` model com colunas `code`, `disciplina`, `ano`, `descricao`. `BnccSeeder` carrega de `database/seeds/bncc.json`.

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/BnccSeederTest.php
use App\Models\BnccSkill;
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('seeder carrega habilidades BNCC do JSON', function () {
    $this->seed(\Database\Seeders\BnccSeeder::class);

    expect(BnccSkill::count())->toBeGreaterThan(0);

    $skill = BnccSkill::where('code', 'EF06MA01')->first();
    expect($skill)->not->toBeNull()
        ->and($skill->disciplina)->toBe('Matemática')
        ->and($skill->ano)->toBe('6º ano');
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=BnccSeederTest`
Expected: FAIL — classe `BnccSeeder` / model `BnccSkill` não existe.

- [ ] **Step 3: Criar migration**

```bash
php artisan make:migration create_bncc_skills_table
```

Conteúdo do `up()`:

```php
Schema::create('bncc_skills', function (Blueprint $table) {
    $table->id();
    $table->string('code')->unique();      // ex: EF06MA01
    $table->string('disciplina');          // ex: Matemática
    $table->string('ano');                 // ex: 6º ano
    $table->text('descricao');
    $table->timestamps();
});
```

- [ ] **Step 4: Criar o model**

```php
<?php
// app/Models/BnccSkill.php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class BnccSkill extends Model
{
    protected $fillable = ['code', 'disciplina', 'ano', 'descricao'];
}
```

- [ ] **Step 5: Criar o JSON de seed**

```json
[
  { "code": "EF06MA01", "disciplina": "Matemática", "ano": "6º ano", "descricao": "Comparar, ordenar, ler e escrever números naturais e racionais na forma decimal." },
  { "code": "EF06MA02", "disciplina": "Matemática", "ano": "6º ano", "descricao": "Reconhecer o sistema de numeração decimal e o valor posicional dos algarismos." },
  { "code": "EF06CI01", "disciplina": "Ciências", "ano": "6º ano", "descricao": "Classificar materiais em naturais e sintéticos quanto à origem." },
  { "code": "EF06LP01", "disciplina": "Língua Portuguesa", "ano": "6º ano", "descricao": "Reconhecer a impossibilidade de uma neutralidade absoluta no relato de fatos." },
  { "code": "EF07HI01", "disciplina": "História", "ano": "7º ano", "descricao": "Explicar o significado de 'modernidade' e suas lógicas de inclusão e exclusão." }
]
```

> Nota para o implementador: estes 5 registros bastam para a demo. Pode ampliar o JSON depois sem mudar código.

- [ ] **Step 6: Criar o seeder**

```php
<?php
// database/seeders/BnccSeeder.php
namespace Database\Seeders;

use App\Models\BnccSkill;
use Illuminate\Database\Seeder;

class BnccSeeder extends Seeder
{
    public function run(): void
    {
        $path = database_path('seeds/bncc.json');
        $rows = json_decode(file_get_contents($path), true);

        foreach ($rows as $row) {
            BnccSkill::updateOrCreate(['code' => $row['code']], $row);
        }
    }
}
```

- [ ] **Step 7: Registrar no DatabaseSeeder**

Em `database/seeders/DatabaseSeeder.php`, dentro de `run()`:

```php
$this->call(BnccSeeder::class);
```

- [ ] **Step 8: Rodar o teste (deve passar)**

Run: `php artisan test --filter=BnccSeederTest`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: BNCC skills model, migration and JSON seed"
```

---

### Task 3: Migrations e models do domínio

**Files:**
- Create: migrations para `lesson_plans`, `study_trails`, `trail_topics`, `quizzes`, `quiz_questions`, `student_attempts`, `attempt_answers`
- Create: models correspondentes em `app/Models/`
- Test: `tests/Feature/DomainModelsTest.php`

**Interfaces:**
- Produces: models com relações — `LessonPlan` hasOne `StudyTrail`; `StudyTrail` hasMany `TrailTopic`, hasOne `Quiz`, hasMany `StudentAttempt`; `Quiz` hasMany `QuizQuestion`; `StudentAttempt` hasMany `AttemptAnswer`.
- `LessonPlan` colunas: `user_id`, `bncc_skill_id`, `duracao_min`, `objetivos`, `metodologia`, `recursos`, `avaliacao`, `atividade`, `status` (`rascunho`|`gerado`|`falha`).
- `StudyTrail` colunas: `lesson_plan_id`, `codigo` (unique), `publicada_em` (nullable).
- `TrailTopic`: `study_trail_id`, `ordem`, `titulo`, `resumo`.
- `QuizQuestion`: `quiz_id`, `enunciado`, `opcoes` (json array), `correta` (int index).
- `StudentAttempt`: `study_trail_id`, `nome_aluno`, `pontos`, `concluido_em` (nullable).
- `AttemptAnswer`: `student_attempt_id`, `quiz_question_id`, `escolhida` (int), `correta` (bool).

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/DomainModelsTest.php
use App\Models\{User, BnccSkill, LessonPlan, StudyTrail, TrailTopic, Quiz, QuizQuestion, StudentAttempt, AttemptAnswer};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('relações do domínio funcionam ponta a ponta', function () {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);

    $plan = LessonPlan::create([
        'user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50,
        'objetivos' => 'o', 'metodologia' => 'm', 'recursos' => 'r', 'avaliacao' => 'a',
        'atividade' => 'at', 'status' => 'gerado',
    ]);
    $trail = $plan->trail()->create(['codigo' => 'TR-AAA1']);
    $trail->topics()->create(['ordem' => 1, 'titulo' => 't', 'resumo' => 's']);
    $quiz = $trail->quiz()->create([]);
    $q = $quiz->questions()->create(['enunciado' => 'q', 'opcoes' => ['a', 'b'], 'correta' => 1]);

    $attempt = $trail->attempts()->create(['nome_aluno' => 'Ana', 'pontos' => 0]);
    $attempt->answers()->create(['quiz_question_id' => $q->id, 'escolhida' => 1, 'correta' => true]);

    expect($plan->trail->codigo)->toBe('TR-AAA1')
        ->and($trail->topics)->toHaveCount(1)
        ->and($quiz->questions->first()->opcoes)->toBe(['a', 'b'])
        ->and($attempt->answers->first()->correta)->toBeTrue();
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=DomainModelsTest`
Expected: FAIL — models não existem.

- [ ] **Step 3: Criar as migrations**

```bash
php artisan make:migration create_lesson_plans_table
php artisan make:migration create_study_trails_table
php artisan make:migration create_trail_topics_table
php artisan make:migration create_quizzes_table
php artisan make:migration create_quiz_questions_table
php artisan make:migration create_student_attempts_table
php artisan make:migration create_attempt_answers_table
```

Conteúdo dos `up()` (uma migration por tabela):

```php
// lesson_plans
Schema::create('lesson_plans', function (Blueprint $table) {
    $table->id();
    $table->foreignId('user_id')->constrained()->cascadeOnDelete();
    $table->foreignId('bncc_skill_id')->constrained('bncc_skills');
    $table->unsignedSmallInteger('duracao_min');
    $table->text('objetivos')->nullable();
    $table->text('metodologia')->nullable();
    $table->text('recursos')->nullable();
    $table->text('avaliacao')->nullable();
    $table->text('atividade')->nullable();
    $table->string('status')->default('rascunho');
    $table->timestamps();
});

// study_trails
Schema::create('study_trails', function (Blueprint $table) {
    $table->id();
    $table->foreignId('lesson_plan_id')->constrained()->cascadeOnDelete();
    $table->string('codigo')->unique();
    $table->timestamp('publicada_em')->nullable();
    $table->timestamps();
});

// trail_topics
Schema::create('trail_topics', function (Blueprint $table) {
    $table->id();
    $table->foreignId('study_trail_id')->constrained()->cascadeOnDelete();
    $table->unsignedSmallInteger('ordem');
    $table->string('titulo');
    $table->text('resumo');
    $table->timestamps();
});

// quizzes
Schema::create('quizzes', function (Blueprint $table) {
    $table->id();
    $table->foreignId('study_trail_id')->constrained()->cascadeOnDelete();
    $table->timestamps();
});

// quiz_questions
Schema::create('quiz_questions', function (Blueprint $table) {
    $table->id();
    $table->foreignId('quiz_id')->constrained()->cascadeOnDelete();
    $table->text('enunciado');
    $table->json('opcoes');
    $table->unsignedTinyInteger('correta');
    $table->timestamps();
});

// student_attempts
Schema::create('student_attempts', function (Blueprint $table) {
    $table->id();
    $table->foreignId('study_trail_id')->constrained()->cascadeOnDelete();
    $table->string('nome_aluno');
    $table->unsignedInteger('pontos')->default(0);
    $table->timestamp('concluido_em')->nullable();
    $table->timestamps();
});

// attempt_answers
Schema::create('attempt_answers', function (Blueprint $table) {
    $table->id();
    $table->foreignId('student_attempt_id')->constrained()->cascadeOnDelete();
    $table->foreignId('quiz_question_id')->constrained()->cascadeOnDelete();
    $table->unsignedTinyInteger('escolhida');
    $table->boolean('correta');
    $table->timestamps();
});
```

- [ ] **Step 4: Criar os models**

```php
<?php
// app/Models/LessonPlan.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class LessonPlan extends Model
{
    protected $fillable = ['user_id','bncc_skill_id','duracao_min','objetivos','metodologia','recursos','avaliacao','atividade','status'];
    public function bnccSkill() { return $this->belongsTo(BnccSkill::class); }
    public function user() { return $this->belongsTo(User::class); }
    public function trail() { return $this->hasOne(StudyTrail::class); }
}
```

```php
<?php
// app/Models/StudyTrail.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class StudyTrail extends Model
{
    protected $fillable = ['lesson_plan_id','codigo','publicada_em'];
    protected $casts = ['publicada_em' => 'datetime'];
    public function lessonPlan() { return $this->belongsTo(LessonPlan::class); }
    public function topics() { return $this->hasMany(TrailTopic::class)->orderBy('ordem'); }
    public function quiz() { return $this->hasOne(Quiz::class); }
    public function attempts() { return $this->hasMany(StudentAttempt::class); }
}
```

```php
<?php
// app/Models/TrailTopic.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class TrailTopic extends Model
{
    protected $fillable = ['study_trail_id','ordem','titulo','resumo'];
    public function trail() { return $this->belongsTo(StudyTrail::class, 'study_trail_id'); }
}
```

```php
<?php
// app/Models/Quiz.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class Quiz extends Model
{
    protected $fillable = ['study_trail_id'];
    public function questions() { return $this->hasMany(QuizQuestion::class); }
    public function trail() { return $this->belongsTo(StudyTrail::class, 'study_trail_id'); }
}
```

```php
<?php
// app/Models/QuizQuestion.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class QuizQuestion extends Model
{
    protected $fillable = ['quiz_id','enunciado','opcoes','correta'];
    protected $casts = ['opcoes' => 'array'];
    public function quiz() { return $this->belongsTo(Quiz::class); }
}
```

```php
<?php
// app/Models/StudentAttempt.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class StudentAttempt extends Model
{
    protected $fillable = ['study_trail_id','nome_aluno','pontos','concluido_em'];
    protected $casts = ['concluido_em' => 'datetime'];
    public function trail() { return $this->belongsTo(StudyTrail::class, 'study_trail_id'); }
    public function answers() { return $this->hasMany(AttemptAnswer::class); }
}
```

```php
<?php
// app/Models/AttemptAnswer.php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class AttemptAnswer extends Model
{
    protected $fillable = ['student_attempt_id','quiz_question_id','escolhida','correta'];
    protected $casts = ['correta' => 'boolean'];
}
```

- [ ] **Step 5: Rodar o teste (deve passar)**

Run: `php artisan test --filter=DomainModelsTest`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: domain migrations and Eloquent models"
```

---

### Task 4: AiService com structured output (client mockado)

**Files:**
- Create: `app/Services/AiService.php`
- Create: `app/Services/LessonData.php` (DTO)
- Test: `tests/Unit/AiServiceTest.php`
- Create: `tests/fixtures/lesson_response.json`

**Interfaces:**
- Produces: `AiService::__construct(\Anthropic\Client $client, string $model)`. Método `generateLesson(BnccSkill $skill, int $duracaoMin): LessonData`.
- `LessonData` propriedades públicas: `array $plano` (`objetivos`,`metodologia`,`recursos`,`avaliacao` strings), `string $atividade`, `array $topicos` (cada `['titulo','resumo']`), `array $questoes` (cada `['enunciado','opcoes'=>[],'correta'=>int]`).

> Nota: o construtor recebe o client por injeção para permitir mock. O binding real é registrado em Task 5.

- [ ] **Step 1: Criar a fixture de resposta**

```json
{
  "plano": { "objetivos": "Compreender números racionais decimais.", "metodologia": "Aula expositiva + exercícios.", "recursos": "Quadro, lista impressa.", "avaliacao": "Participação e quiz." },
  "atividade": "Lista de 5 exercícios de ordenação de decimais.",
  "trilha": {
    "topicos": [
      { "titulo": "O que é número decimal", "resumo": "Decimais representam partes do inteiro." },
      { "titulo": "Comparando decimais", "resumo": "Compare casa a casa, da esquerda para a direita." }
    ],
    "quiz": { "questoes": [
      { "enunciado": "Qual é maior?", "opcoes": ["0,5", "0,45"], "correta": 0 }
    ] }
  }
}
```

- [ ] **Step 2: Escrever o teste falhando**

```php
<?php
// tests/Unit/AiServiceTest.php
use App\Models\BnccSkill;
use App\Services\AiService;
use App\Services\LessonData;

test('generateLesson parseia structured output em LessonData', function () {
    $payload = file_get_contents(base_path('tests/fixtures/lesson_response.json'));

    // bloco de texto que a API retorna quando output_config.format é json_schema
    $textBlock = (object) ['type' => 'text', 'text' => $payload];
    $message = (object) ['content' => [$textBlock]];

    $messages = Mockery::mock();
    $messages->shouldReceive('create')->once()->andReturn($message);
    $client = Mockery::mock(\Anthropic\Client::class);
    $client->messages = $messages;

    $skill = new BnccSkill(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'decimais']);

    $service = new AiService($client, 'claude-opus-4-8');
    $data = $service->generateLesson($skill, 50);

    expect($data)->toBeInstanceOf(LessonData::class)
        ->and($data->plano['objetivos'])->toBe('Compreender números racionais decimais.')
        ->and($data->atividade)->toContain('5 exercícios')
        ->and($data->topicos)->toHaveCount(2)
        ->and($data->questoes[0]['correta'])->toBe(0);
});
```

- [ ] **Step 3: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=AiServiceTest`
Expected: FAIL — `AiService` / `LessonData` não existem.

- [ ] **Step 4: Criar o DTO LessonData**

```php
<?php
// app/Services/LessonData.php
namespace App\Services;

class LessonData
{
    public function __construct(
        public array $plano,
        public string $atividade,
        public array $topicos,
        public array $questoes,
    ) {}

    public static function fromArray(array $d): self
    {
        return new self(
            plano: $d['plano'],
            atividade: $d['atividade'],
            topicos: $d['trilha']['topicos'],
            questoes: $d['trilha']['quiz']['questoes'],
        );
    }
}
```

- [ ] **Step 5: Criar o AiService**

```php
<?php
// app/Services/AiService.php
namespace App\Services;

use App\Models\BnccSkill;

class AiService
{
    public function __construct(
        private \Anthropic\Client $client,
        private string $model,
    ) {}

    public function generateLesson(BnccSkill $skill, int $duracaoMin): LessonData
    {
        $prompt = $this->buildPrompt($skill, $duracaoMin);

        $message = $this->client->messages->create(
            model: $this->model,
            maxTokens: 8000,
            outputConfig: ['format' => ['type' => 'json_schema', 'schema' => $this->schema()]],
            messages: [['role' => 'user', 'content' => $prompt]],
        );

        $text = '';
        foreach ($message->content as $block) {
            if (($block->type ?? null) === 'text') {
                $text = $block->text;
                break;
            }
        }

        $data = json_decode($text, true, 512, JSON_THROW_ON_ERROR);
        return LessonData::fromArray($data);
    }

    private function buildPrompt(BnccSkill $skill, int $duracaoMin): string
    {
        return "Você é um planejador pedagógico alinhado à BNCC. "
            . "Gere um plano de aula e uma trilha de estudo do aluno.\n\n"
            . "Disciplina: {$skill->disciplina}\n"
            . "Ano/série: {$skill->ano}\n"
            . "Habilidade BNCC: {$skill->code} — {$skill->descricao}\n"
            . "Duração da aula: {$duracaoMin} minutos\n\n"
            . "Responda APENAS no formato JSON definido pelo schema. "
            . "A trilha deve ter 2 a 4 tópicos curtos e um quiz de 3 a 5 questões "
            . "de múltipla escolha (índice da opção correta em 'correta').";
    }

    private function schema(): array
    {
        return [
            'type' => 'object',
            'additionalProperties' => false,
            'required' => ['plano', 'atividade', 'trilha'],
            'properties' => [
                'plano' => [
                    'type' => 'object',
                    'additionalProperties' => false,
                    'required' => ['objetivos', 'metodologia', 'recursos', 'avaliacao'],
                    'properties' => [
                        'objetivos' => ['type' => 'string'],
                        'metodologia' => ['type' => 'string'],
                        'recursos' => ['type' => 'string'],
                        'avaliacao' => ['type' => 'string'],
                    ],
                ],
                'atividade' => ['type' => 'string'],
                'trilha' => [
                    'type' => 'object',
                    'additionalProperties' => false,
                    'required' => ['topicos', 'quiz'],
                    'properties' => [
                        'topicos' => [
                            'type' => 'array',
                            'items' => [
                                'type' => 'object',
                                'additionalProperties' => false,
                                'required' => ['titulo', 'resumo'],
                                'properties' => [
                                    'titulo' => ['type' => 'string'],
                                    'resumo' => ['type' => 'string'],
                                ],
                            ],
                        ],
                        'quiz' => [
                            'type' => 'object',
                            'additionalProperties' => false,
                            'required' => ['questoes'],
                            'properties' => [
                                'questoes' => [
                                    'type' => 'array',
                                    'items' => [
                                        'type' => 'object',
                                        'additionalProperties' => false,
                                        'required' => ['enunciado', 'opcoes', 'correta'],
                                        'properties' => [
                                            'enunciado' => ['type' => 'string'],
                                            'opcoes' => ['type' => 'array', 'items' => ['type' => 'string']],
                                            'correta' => ['type' => 'integer'],
                                        ],
                                    ],
                                ],
                            ],
                        ],
                    ],
                ],
            ],
        ];
    }
}
```

> Nota: se a assinatura real de `messages->create` no SDK divergir (ex.: nome da chave nested), ajuste apenas a chamada — o teste mocka `create` e não valida argumentos, então o contrato do DTO permanece. Confirme a forma de `outputConfig` no README PHP do SDK instalado antes de rodar contra a API real.

- [ ] **Step 6: Rodar o teste (deve passar)**

Run: `php artisan test --filter=AiServiceTest`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: AiService with structured output and LessonData DTO"
```

---

### Task 5: Filament + auth do professor + binding do AiService

**Files:**
- Modify: `app/Providers/AppServiceProvider.php`
- Create: painel Filament + usuário admin
- Test: `tests/Feature/AiServiceBindingTest.php`

**Interfaces:**
- Consumes: `AiService` (Task 4).
- Produces: `app(AiService::class)` resolvível com client real configurado a partir de `config('services.anthropic')`. Painel Filament acessível em `/admin` com login.

- [ ] **Step 1: Instalar Filament**

```bash
composer require filament/filament:"^3.2"
php artisan filament:install --panels
```

Aceite o painel padrão `admin`.

- [ ] **Step 2: Criar usuário professor**

```bash
php artisan make:filament-user
```

Informe nome, email e senha quando solicitado (estes são as credenciais de demo do professor).

- [ ] **Step 3: Escrever o teste falhando do binding**

```php
<?php
// tests/Feature/AiServiceBindingTest.php
use App\Services\AiService;

test('AiService é resolvível pelo container', function () {
    config()->set('services.anthropic.key', 'sk-test');
    config()->set('services.anthropic.model', 'claude-opus-4-8');

    $service = app(AiService::class);

    expect($service)->toBeInstanceOf(AiService::class);
});
```

- [ ] **Step 4: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=AiServiceBindingTest`
Expected: FAIL — container não sabe construir `AiService` (precisa do client e do model).

- [ ] **Step 5: Registrar o binding**

Em `app/Providers/AppServiceProvider.php`, método `register()`:

```php
use App\Services\AiService;

$this->app->singleton(AiService::class, function ($app) {
    $client = \Anthropic::client(config('services.anthropic.key'));
    return new AiService($client, config('services.anthropic.model'));
});
```

> Nota: confirme o factory real do SDK PHP. Conforme o README do `anthropic-ai/sdk`, o client é `new \Anthropic\Client(apiKey: ...)`. Se a facade `\Anthropic::client()` não existir na versão instalada, use `new \Anthropic\Client(apiKey: config('services.anthropic.key'))`.

- [ ] **Step 6: Rodar o teste (deve passar)**

Run: `php artisan test --filter=AiServiceBindingTest`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: Filament panel, professor auth, AiService container binding"
```

---

### Task 6: Resource Filament de LessonPlan com ação "Gerar com IA"

**Files:**
- Create: `app/Filament/Resources/LessonPlanResource.php` (+ páginas geradas)
- Create: `app/Actions/GenerateLessonAction.php`
- Test: `tests/Feature/GenerateLessonActionTest.php`

**Interfaces:**
- Consumes: `AiService::generateLesson` (Task 4), models (Task 3).
- Produces: `GenerateLessonAction::execute(LessonPlan $plan): void` — chama IA, popula campos do plano, cria `StudyTrail` (com código), `TrailTopic`s, `Quiz` + `QuizQuestion`s, marca `status='gerado'`. Em exceção, `status='falha'`.
- Código da trilha: `StudyTrail::gerarCodigo(): string` formato `TR-XXXX` (4 chars alfanuméricos maiúsculos, único).

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/GenerateLessonActionTest.php
use App\Actions\GenerateLessonAction;
use App\Models\{User, BnccSkill, LessonPlan};
use App\Services\{AiService, LessonData};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('ação gera plano, trilha, tópicos e quiz', function () {
    $data = new LessonData(
        plano: ['objetivos' => 'o', 'metodologia' => 'm', 'recursos' => 'r', 'avaliacao' => 'a'],
        atividade: 'atividade x',
        topicos: [['titulo' => 't1', 'resumo' => 's1'], ['titulo' => 't2', 'resumo' => 's2']],
        questoes: [['enunciado' => 'q', 'opcoes' => ['a', 'b'], 'correta' => 1]],
    );
    $ai = Mockery::mock(AiService::class);
    $ai->shouldReceive('generateLesson')->once()->andReturn($data);
    app()->instance(AiService::class, $ai);

    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create([
        'user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'rascunho',
    ]);

    app(GenerateLessonAction::class)->execute($plan);
    $plan->refresh();

    expect($plan->status)->toBe('gerado')
        ->and($plan->objetivos)->toBe('o')
        ->and($plan->atividade)->toBe('atividade x')
        ->and($plan->trail)->not->toBeNull()
        ->and($plan->trail->codigo)->toStartWith('TR-')
        ->and($plan->trail->topics)->toHaveCount(2)
        ->and($plan->trail->quiz->questions)->toHaveCount(1)
        ->and($plan->trail->quiz->questions->first()->correta)->toBe(1);
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=GenerateLessonActionTest`
Expected: FAIL — `GenerateLessonAction` não existe.

- [ ] **Step 3: Adicionar gerador de código no StudyTrail**

Em `app/Models/StudyTrail.php`, adicionar:

```php
public static function gerarCodigo(): string
{
    do {
        $codigo = 'TR-' . strtoupper(\Illuminate\Support\Str::random(4));
    } while (self::where('codigo', $codigo)->exists());
    return $codigo;
}
```

- [ ] **Step 4: Criar a action**

```php
<?php
// app/Actions/GenerateLessonAction.php
namespace App\Actions;

use App\Models\LessonPlan;
use App\Models\StudyTrail;
use App\Services\AiService;
use Illuminate\Support\Facades\DB;
use Throwable;

class GenerateLessonAction
{
    public function __construct(private AiService $ai) {}

    public function execute(LessonPlan $plan): void
    {
        try {
            $data = $this->ai->generateLesson($plan->bnccSkill, $plan->duracao_min);

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
                        'correta' => $q['correta'],
                    ]);
                }
            });
        } catch (Throwable $e) {
            $plan->update(['status' => 'falha']);
            report($e);
        }
    }
}
```

- [ ] **Step 5: Rodar o teste (deve passar)**

Run: `php artisan test --filter=GenerateLessonActionTest`
Expected: PASS.

- [ ] **Step 6: Criar o resource Filament**

```bash
php artisan make:filament-resource LessonPlan --generate
```

Editar `app/Filament/Resources/LessonPlanResource.php` para:
- Form: `Select` de `bncc_skill_id` (options `BnccSkill::pluck('code'... )` — exibir `code — disciplina`), `TextInput` numérico `duracao_min`. Campos gerados (`objetivos`, etc.) como `Textarea` read-only/visíveis.
- Table: colunas `bnccSkill.code`, `status`, `trail.codigo`.
- Header action "Gerar com IA" na página de edição que chama a action:

```php
use App\Actions\GenerateLessonAction;
use Filament\Actions\Action;

Action::make('gerar')
    ->label('Gerar com IA')
    ->icon('heroicon-o-sparkles')
    ->requiresConfirmation()
    ->action(function () {
        app(GenerateLessonAction::class)->execute($this->record);
        $this->refreshFormData(['objetivos','metodologia','recursos','avaliacao','atividade','status']);
    });
```

> Nota: ajuste de UI do Filament é flexível; o critério de "pronto" é: criar plano escolhendo BNCC+duração, clicar "Gerar com IA", e ver os campos preenchidos + trilha com código.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: LessonPlan Filament resource and generate-with-AI action"
```

---

### Task 7: Publicar trilha + página pública do aluno (Blade)

**Files:**
- Create: `app/Http/Controllers/TrailController.php`
- Create: `resources/views/trail/show.blade.php`
- Modify: `routes/web.php`
- Modify: `app/Models/StudyTrail.php` (escopo de publicada)
- Test: `tests/Feature/TrailAccessTest.php`

**Interfaces:**
- Consumes: `StudyTrail`, `TrailTopic`, `Quiz` (Task 3).
- Produces: rota `GET /t/{codigo}` (nome `trail.show`) renderiza trilha publicada; 404 se não publicada/inexistente. `StudyTrail::publicar()` seta `publicada_em`. Scope `published()`.

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/TrailAccessTest.php
use App\Models\{User, BnccSkill, LessonPlan, StudyTrail};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

function fazerTrilha(bool $publicada): StudyTrail {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-DEMO']);
    $trail->topics()->create(['ordem' => 1, 'titulo' => 'Tópico 1', 'resumo' => 'resumo aqui']);
    if ($publicada) $trail->publicar();
    return $trail;
}

test('trilha publicada é acessível por código', function () {
    fazerTrilha(true);
    $this->get('/t/TR-DEMO')->assertOk()->assertSee('Tópico 1');
});

test('trilha não publicada retorna 404', function () {
    fazerTrilha(false);
    $this->get('/t/TR-DEMO')->assertNotFound();
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=TrailAccessTest`
Expected: FAIL — rota/método não existem.

- [ ] **Step 3: Adicionar publicar() e scope no StudyTrail**

```php
public function publicar(): void
{
    $this->update(['publicada_em' => now()]);
}

public function scopePublished($query)
{
    return $query->whereNotNull('publicada_em');
}
```

- [ ] **Step 4: Criar controller**

```php
<?php
// app/Http/Controllers/TrailController.php
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
```

- [ ] **Step 5: Registrar rota**

Em `routes/web.php`:

```php
use App\Http\Controllers\TrailController;

Route::get('/t/{codigo}', [TrailController::class, 'show'])->name('trail.show');
```

- [ ] **Step 6: Criar a view (modo foco básico)**

```blade
{{-- resources/views/trail/show.blade.php --}}
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Trilha {{ $trail->codigo }}</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 640px; margin: 2rem auto; padding: 0 1rem; line-height: 1.5; }
        .topico { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; margin: 1rem 0; }
        h1 { font-size: 1.4rem; }
    </style>
</head>
<body>
    <h1>{{ $trail->lessonPlan->bnccSkill->disciplina }} — {{ $trail->lessonPlan->bnccSkill->ano }}</h1>
    <p>Código da trilha: <strong>{{ $trail->codigo }}</strong></p>

    @foreach ($trail->topics as $topico)
        <div class="topico">
            <h2>{{ $topico->ordem }}. {{ $topico->titulo }}</h2>
            <p>{{ $topico->resumo }}</p>
        </div>
    @endforeach

    @if ($trail->quiz)
        <a href="{{ route('quiz.show', $trail->codigo) }}">Fazer o quiz →</a>
    @endif
</body>
</html>
```

> Nota: a rota `quiz.show` é criada na Task 8; deixe o link aqui — o teste desta task não navega até ele.

- [ ] **Step 7: Adicionar action "Publicar" no Filament**

Em `LessonPlanResource`, header action (página de edição), visível quando `status === 'gerado'`:

```php
Action::make('publicar')
    ->label('Publicar trilha')
    ->visible(fn () => $this->record->trail && ! $this->record->trail->publicada_em)
    ->action(fn () => $this->record->trail->publicar());
```

- [ ] **Step 8: Rodar o teste (deve passar)**

Run: `php artisan test --filter=TrailAccessTest`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: publish trail and public student trail page"
```

---

### Task 8: Quiz autocorrigido + gamificação

**Files:**
- Create: `app/Http/Controllers/QuizController.php`
- Create: `resources/views/quiz/show.blade.php`, `resources/views/quiz/result.blade.php`
- Modify: `routes/web.php`
- Test: `tests/Feature/QuizSubmissionTest.php`

**Interfaces:**
- Consumes: `StudyTrail`, `Quiz`, `QuizQuestion`, `StudentAttempt`, `AttemptAnswer` (Task 3).
- Produces: `GET /t/{codigo}/quiz` (`quiz.show`) form com nome + questões; `POST /t/{codigo}/quiz` (`quiz.submit`) cria `StudentAttempt`, corrige cada resposta, calcula `pontos` (10 por acerto), seta `concluido_em`, renderiza resultado com pontos e barra de progresso (acertos/total).

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/QuizSubmissionTest.php
use App\Models\{User, BnccSkill, LessonPlan, StudentAttempt};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

function trilhaComQuiz(): string {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-QUIZ']);
    $trail->publicar();
    $quiz = $trail->quiz()->create([]);
    $quiz->questions()->create(['enunciado' => 'Q1', 'opcoes' => ['a', 'b'], 'correta' => 0]);
    $quiz->questions()->create(['enunciado' => 'Q2', 'opcoes' => ['x', 'y'], 'correta' => 1]);
    return 'TR-QUIZ';
}

test('submissão corrige respostas e pontua', function () {
    $codigo = trilhaComQuiz();
    $perguntas = \App\Models\StudyTrail::where('codigo', $codigo)->first()->quiz->questions;

    $resp = $this->post("/t/{$codigo}/quiz", [
        'nome_aluno' => 'Ana',
        'respostas' => [
            $perguntas[0]->id => 0, // correta
            $perguntas[1]->id => 0, // errada
        ],
    ]);

    $resp->assertOk()->assertSee('10'); // 1 acerto = 10 pontos

    $attempt = StudentAttempt::where('nome_aluno', 'Ana')->first();
    expect($attempt->pontos)->toBe(10)
        ->and($attempt->concluido_em)->not->toBeNull()
        ->and($attempt->answers)->toHaveCount(2)
        ->and($attempt->answers->where('correta', true))->toHaveCount(1);
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=QuizSubmissionTest`
Expected: FAIL — rotas/controller não existem.

- [ ] **Step 3: Criar controller**

```php
<?php
// app/Http/Controllers/QuizController.php
namespace App\Http\Controllers;

use App\Models\StudyTrail;
use Illuminate\Http\Request;

class QuizController extends Controller
{
    public function show(string $codigo)
    {
        $trail = StudyTrail::published()->where('codigo', $codigo)
            ->with('quiz.questions')->firstOrFail();

        return view('quiz.show', ['trail' => $trail]);
    }

    public function submit(Request $request, string $codigo)
    {
        $trail = StudyTrail::published()->where('codigo', $codigo)
            ->with('quiz.questions')->firstOrFail();

        $validated = $request->validate([
            'nome_aluno' => 'required|string|max:120',
            'respostas' => 'required|array',
        ]);

        $attempt = $trail->attempts()->create([
            'nome_aluno' => $validated['nome_aluno'],
            'pontos' => 0,
        ]);

        $acertos = 0;
        foreach ($trail->quiz->questions as $q) {
            $escolhida = (int) ($validated['respostas'][$q->id] ?? -1);
            $correta = $escolhida === (int) $q->correta;
            if ($correta) $acertos++;
            $attempt->answers()->create([
                'quiz_question_id' => $q->id,
                'escolhida' => max(0, $escolhida),
                'correta' => $correta,
            ]);
        }

        $total = $trail->quiz->questions->count();
        $attempt->update(['pontos' => $acertos * 10, 'concluido_em' => now()]);

        return view('quiz.result', [
            'trail' => $trail,
            'attempt' => $attempt->fresh(),
            'acertos' => $acertos,
            'total' => $total,
        ]);
    }
}
```

- [ ] **Step 4: Registrar rotas**

Em `routes/web.php`:

```php
use App\Http\Controllers\QuizController;

Route::get('/t/{codigo}/quiz', [QuizController::class, 'show'])->name('quiz.show');
Route::post('/t/{codigo}/quiz', [QuizController::class, 'submit'])->name('quiz.submit');
```

- [ ] **Step 5: Criar view do quiz**

```blade
{{-- resources/views/quiz/show.blade.php --}}
<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Quiz {{ $trail->codigo }}</title>
<style>body{font-family:system-ui,sans-serif;max-width:640px;margin:2rem auto;padding:0 1rem;line-height:1.5}.q{border:1px solid #ddd;border-radius:8px;padding:1rem;margin:1rem 0}button{padding:.6rem 1rem;font-size:1rem}</style>
</head>
<body>
    <h1>Quiz — {{ $trail->codigo }}</h1>
    <form method="POST" action="{{ route('quiz.submit', $trail->codigo) }}">
        @csrf
        <p><label>Seu nome: <input type="text" name="nome_aluno" required></label></p>
        @foreach ($trail->quiz->questions as $q)
            <div class="q">
                <p><strong>{{ $loop->iteration }}. {{ $q->enunciado }}</strong></p>
                @foreach ($q->opcoes as $i => $opcao)
                    <label style="display:block">
                        <input type="radio" name="respostas[{{ $q->id }}]" value="{{ $i }}" required> {{ $opcao }}
                    </label>
                @endforeach
            </div>
        @endforeach
        <button type="submit">Enviar respostas</button>
    </form>
</body>
</html>
```

- [ ] **Step 6: Criar view de resultado (gamificação)**

```blade
{{-- resources/views/quiz/result.blade.php --}}
<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Resultado</title>
<style>body{font-family:system-ui,sans-serif;max-width:640px;margin:2rem auto;padding:0 1rem;line-height:1.5}.bar{background:#eee;border-radius:8px;overflow:hidden;height:24px}.fill{background:#4caf50;height:100%;text-align:center;color:#fff;font-size:.8rem;line-height:24px}</style>
</head>
<body>
    <h1>Mandou bem, {{ $attempt->nome_aluno }}!</h1>
    <p>Pontuação: <strong>{{ $attempt->pontos }}</strong> pontos</p>
    <p>{{ $acertos }} de {{ $total }} acertos</p>
    <div class="bar">
        <div class="fill" style="width: {{ $total ? round($acertos / $total * 100) : 0 }}%">
            {{ $total ? round($acertos / $total * 100) : 0 }}%
        </div>
    </div>
    <p><a href="{{ route('trail.show', $trail->codigo) }}">← Voltar à trilha</a></p>
</body>
</html>
```

- [ ] **Step 7: Rodar o teste (deve passar)**

Run: `php artisan test --filter=QuizSubmissionTest`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: self-graded quiz with scoring and progress bar"
```

---

### Task 9: Dashboard de turma (Filament)

**Files:**
- Create: `app/Filament/Resources/LessonPlanResource/Pages/Turma.php` (ou widget) — listagem de tentativas da trilha
- Test: `tests/Feature/TurmaStatsTest.php`

**Interfaces:**
- Consumes: `StudentAttempt` (Task 3).
- Produces: `StudyTrail::estatisticas(): array` com `['total_alunos','media_pontos','concluidos']`. Página Filament exibindo a tabela de tentativas + estatísticas.

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/TurmaStatsTest.php
use App\Models\{User, BnccSkill, LessonPlan};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('estatísticas da turma somam tentativas', function () {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-TURMA']);
    $trail->attempts()->create(['nome_aluno' => 'A', 'pontos' => 20, 'concluido_em' => now()]);
    $trail->attempts()->create(['nome_aluno' => 'B', 'pontos' => 10, 'concluido_em' => now()]);

    $stats = $trail->estatisticas();

    expect($stats['total_alunos'])->toBe(2)
        ->and($stats['media_pontos'])->toBe(15.0)
        ->and($stats['concluidos'])->toBe(2);
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=TurmaStatsTest`
Expected: FAIL — método `estatisticas` não existe.

- [ ] **Step 3: Adicionar estatisticas() no StudyTrail**

```php
public function estatisticas(): array
{
    $attempts = $this->attempts();
    return [
        'total_alunos' => $attempts->count(),
        'media_pontos' => round((float) $attempts->avg('pontos'), 1),
        'concluidos' => $attempts->whereNotNull('concluido_em')->count(),
    ];
}
```

- [ ] **Step 4: Rodar o teste (deve passar)**

Run: `php artisan test --filter=TurmaStatsTest`
Expected: PASS.

- [ ] **Step 5: Criar página Filament de turma**

```bash
php artisan make:filament-page Pages/Turma --resource=LessonPlanResource --type=custom
```

Na página, expor `$this->record->trail->estatisticas()` e uma tabela das `attempts` (`nome_aluno`, `pontos`, `concluido_em`). UI flexível; critério de pronto: professor abre a turma de um plano publicado e vê alunos + média.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: class dashboard with attempt stats"
```

---

### Task 10: Export PDF + link WhatsApp

**Files:**
- Create: `app/Http/Controllers/ExportController.php`
- Create: `resources/views/trail/pdf.blade.php`
- Modify: `routes/web.php`, `resources/views/trail/show.blade.php`
- Test: `tests/Feature/ExportTest.php`

**Interfaces:**
- Consumes: `StudyTrail` (Task 3), `barryvdh/laravel-dompdf`.
- Produces: `GET /t/{codigo}/pdf` (`trail.pdf`) retorna PDF da trilha (Content-Type `application/pdf`). Página da trilha mostra botão "Compartilhar no WhatsApp" via link `https://wa.me/?text=`.

- [ ] **Step 1: Escrever o teste falhando**

```php
<?php
// tests/Feature/ExportTest.php
use App\Models\{User, BnccSkill, LessonPlan};
use Illuminate\Foundation\Testing\RefreshDatabase;

uses(RefreshDatabase::class);

test('export PDF retorna application/pdf', function () {
    $user = User::factory()->create();
    $skill = BnccSkill::create(['code' => 'EF06MA01', 'disciplina' => 'Matemática', 'ano' => '6º ano', 'descricao' => 'x']);
    $plan = LessonPlan::create(['user_id' => $user->id, 'bncc_skill_id' => $skill->id, 'duracao_min' => 50, 'status' => 'gerado']);
    $trail = $plan->trail()->create(['codigo' => 'TR-PDF']);
    $trail->topics()->create(['ordem' => 1, 'titulo' => 'T1', 'resumo' => 'r1']);
    $trail->publicar();

    $resp = $this->get('/t/TR-PDF/pdf');
    $resp->assertOk();
    expect($resp->headers->get('content-type'))->toContain('application/pdf');
});
```

- [ ] **Step 2: Rodar o teste (deve falhar)**

Run: `php artisan test --filter=ExportTest`
Expected: FAIL — rota não existe.

- [ ] **Step 3: Criar view do PDF**

```blade
{{-- resources/views/trail/pdf.blade.php --}}
<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="utf-8"><style>body{font-family:DejaVu Sans,sans-serif;font-size:12px}h1{font-size:16px}.t{margin:8px 0}</style></head>
<body>
    <h1>Trilha {{ $trail->codigo }} — {{ $trail->lessonPlan->bnccSkill->disciplina }}</h1>
    @foreach ($trail->topics as $topico)
        <div class="t"><strong>{{ $topico->ordem }}. {{ $topico->titulo }}</strong><br>{{ $topico->resumo }}</div>
    @endforeach
</body>
</html>
```

- [ ] **Step 4: Criar controller**

```php
<?php
// app/Http/Controllers/ExportController.php
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
```

- [ ] **Step 5: Registrar rota**

Em `routes/web.php`:

```php
use App\Http\Controllers\ExportController;

Route::get('/t/{codigo}/pdf', [ExportController::class, 'pdf'])->name('trail.pdf');
```

- [ ] **Step 6: Adicionar botões na página da trilha**

Em `resources/views/trail/show.blade.php`, antes de `</body>`:

```blade
<p>
    <a href="{{ route('trail.pdf', $trail->codigo) }}">Baixar PDF</a>
    &nbsp;|&nbsp;
    <a href="https://wa.me/?text={{ urlencode('Estude esta trilha: ' . route('trail.show', $trail->codigo)) }}" target="_blank" rel="noopener">Compartilhar no WhatsApp</a>
</p>
```

- [ ] **Step 7: Rodar o teste (deve passar)**

Run: `php artisan test --filter=ExportTest`
Expected: PASS.

- [ ] **Step 8: Rodar a suíte completa**

Run: `php artisan test`
Expected: todos os testes PASS.

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat: PDF export and WhatsApp share link"
```

---

## Notas de fechamento (não são tarefas)

- README do projeto + diagrama de arquitetura simples são exigidos no relatório do hackathon — produzir após o MVP funcional.
- Tornar o repositório GitHub público antes da entrega (11/04/2026 no PDF; confirmar data real do hackathon).
- Para o vídeo do MVP: roteiro = login professor → criar plano (BNCC+duração) → "Gerar com IA" → publicar → abrir link do aluno → fazer quiz → ver pontuação → dashboard da turma.
