# Aulas sequenciais por turma — Design

## Objetivo
Uma turma pode ter várias aulas em sequência, com progressão travada: o aluno
só libera a próxima aula depois de concluir a anterior. Uma mesma aula pode ser
reutilizada em várias turmas. O professor monta e reordena a sequência na página
da turma.

## Decisões
- **Progressão travada (unlock):** aula `i` libera após concluir aula `i-1`.
- **Concluir = enviar o quiz** (qualquer nota).
- **Aula reutilizável:** a mesma aula pode entrar em várias turmas, em posições
  diferentes → tabela de ligação `turma_lessons`.
- **Montagem na página da turma:** adicionar aula existente (status `pronto`) e
  reordenar via ▲▼.

## Modelo de dados
Nova tabela **`turma_lessons`**:

| coluna         | tipo        | notas                                    |
|----------------|-------------|------------------------------------------|
| id             | bigserial   | PK                                       |
| turma_id       | bigint      | FK turmas, ON DELETE CASCADE             |
| lesson_plan_id | bigint      | FK lesson_plans, ON DELETE CASCADE       |
| ordem          | int         | 1-based, contígua por turma              |
| created_at     | timestamptz | default now()                            |

- `UNIQUE (turma_id, lesson_plan_id)` — uma aula aparece no máx. uma vez por turma.
- índice `(turma_id, ordem)`.
- Backfill: para cada `study_trails.turma_id IS NOT NULL`, inserir uma linha
  (turma_id, lesson_plan_id da trilha, ordem por `publicada_em`).
- `study_trails.turma_id` vira legado (não removido nesta iteração). Publish
  continua gerando `codigo`; o vínculo de turma passa a ser `turma_lessons`.

## Conclusão / unlock (derivado — sem tabela de progresso)
- Concluída para um aluno = existe `student_attempts` com
  `concluido_em IS NOT NULL` para `(student_id, study_trail_id da aula)`.
- Unlock: `ordem = 1` sempre liberada; `ordem = i` liberada se a aula de
  `ordem = i-1` está concluída para aquele aluno.

## Backend (endpoints)
Professor (grupo autenticado existente):
- `GET /api/turmas/{id}` — resposta ganha `aulas: [{lesson_plan_id, label,
  ordem, status, codigo}]` (ordenado por `ordem`).
- `POST /api/turmas/{id}/lessons` — body `{lesson_plan_id}`. Anexa no fim
  (`max(ordem)+1`). Valida: aula pertence ao professor e `status = pronto`.
  Garante trilha publicada (gera `codigo` se ainda não tem). Idempotente contra
  duplicata via UNIQUE (retorna 409/erro claro).
- `DELETE /api/turmas/{id}/lessons/{lessonPlanId}` — remove o vínculo e renumera
  as aulas restantes (fecha buracos).
- `PATCH /api/turmas/{id}/lessons` — body `{ordered_ids: [lesson_plan_id,…]}`.
  Renumera atômico (transação); valida que o conjunto bate com as aulas atuais.

Aluno (grupo `RequireStudent`):
- `GET /api/student/lessons` — aulas ordenadas da turma do aluno, cada uma com
  `{ordem, label, codigo, unlocked, concluido, pontos}`.

Seletor do professor usa `GET /api/lessons` (já existe) para escolher a aula.

## Frontend
Professor — `turmas/{id}` ganha seção **"Aulas da turma"**:
- Lista ordenada: badge `ordem` · label · chip de status · ▲▼ mover · remover.
- Botão **"Adicionar aula"** → picker das aulas `pronto` do professor que ainda
  não estão na turma.
- Reordenar chama `PATCH …/lessons`; remover chama `DELETE`.

Aluno — nova página **`/aluno`** (home pós-login):
- Lista ordenada com ícone de estado: cadeado (travada) · seta (liberada) · ✓
  (concluída, mostra pontos).
- Liberada → link para `/t/{codigo}` (fluxo atual da trilha).
- Travada → desabilitada + texto "Conclua a aula anterior".
- `studentLogin` passa a redirecionar para `/aluno`.

## Label da aula
`lesson_plans` não tem título. Label = assunto/código da habilidade BNCC
vinculada; fallback `"Aula #{id}"`.

## Testes (Go)
- Attach: sucesso; rejeita aula de outro professor; rejeita status ≠ pronto;
  duplicata bloqueada pelo UNIQUE.
- Detach: remove e renumera (sem buracos).
- Reorder: renumera atômico; rejeita conjunto que não bate.
- Unlock: aluno sem conclusões vê só a ordem 1 liberada; após concluir a 1ª
  (attempt com `concluido_em`), a 2ª libera.
- `GET /api/student/lessons`: locked/unlocked/concluido corretos.
- Fixtures com TRUNCATE (padrão do projeto para evitar colisão).

## Fora de escopo (YAGNI)
- Nota mínima para liberar (escolhido: qualquer envio).
- Drag-and-drop (usar ▲▼).
- Remover `study_trails.turma_id` (deixado como legado).
- Título editável por aula.
