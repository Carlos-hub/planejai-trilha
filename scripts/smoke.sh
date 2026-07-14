#!/usr/bin/env bash
# End-to-end smoke test for PlanejAI+Trilha.
#
# Brings up the stack (db + api, and web if not skipped), then drives the
# full teacher/student flow against the real HTTP API using the MANUAL
# lesson-creation path (no dependency on the LLM /generate endpoint, which
# requires a real ANTHROPIC_API_KEY).
#
# Usage:
#   bash scripts/smoke.sh            # brings up db, api, web
#   SMOKE_SKIP_WEB=1 bash scripts/smoke.sh   # brings up only db, api
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COOKIE_JAR="$(mktemp -d)/smoke-cookies.txt"
trap 'rm -rf "$(dirname "$COOKIE_JAR")"' EXIT

log() { printf '\n\033[1;34m==>\033[0m %s\n' "$1"; }
fail() { printf '\n\033[1;31mFAIL:\033[0m %s\n' "$1" >&2; exit 1; }
ok() { printf '\033[1;32m✓\033[0m %s\n' "$1"; }

cd "$ROOT_DIR"

if [[ ! -f .env ]]; then
  log "No .env found, copying from .env.example"
  cp .env.example .env
fi

log "Starting stack (docker compose up -d --build)"
if [[ "${SMOKE_SKIP_WEB:-0}" == "1" ]]; then
  docker compose up -d --build db api
else
  docker compose up -d --build db api web
fi

log "Waiting for db to be healthy"
for i in $(seq 1 60); do
  status="$(docker compose ps db --format '{{.Health}}' 2>/dev/null || true)"
  if [[ "$status" == "healthy" ]]; then
    ok "db healthy"
    break
  fi
  if [[ "$i" == 60 ]]; then
    fail "db did not become healthy in time"
  fi
  sleep 2
done

log "Waiting for api /healthz"
for i in $(seq 1 60); do
  if curl -sf "$API_URL/healthz" >/dev/null 2>&1; then
    ok "api is up"
    break
  fi
  if [[ "$i" == 60 ]]; then
    fail "api did not become healthy in time"
  fi
  sleep 2
done

# ---------------------------------------------------------------------------
# 1. Login as demo teacher
# ---------------------------------------------------------------------------
log "Logging in as prof@demo.com"
login_status="$(curl -s -o /tmp/smoke_login.json -w '%{http_code}' \
  -c "$COOKIE_JAR" -H 'Content-Type: application/json' \
  -d '{"Email":"prof@demo.com","Senha":"demo1234"}' \
  "$API_URL/api/auth/login")"
[[ "$login_status" == "200" ]] || fail "login expected 200, got $login_status: $(cat /tmp/smoke_login.json)"
ok "login OK (200)"

# ---------------------------------------------------------------------------
# 2. Create a MANUAL lesson (plano + atividade + topicos + quiz) — no LLM
# ---------------------------------------------------------------------------
log "Creating manual lesson"
lesson_payload='{
  "duracao": 50,
  "plano": {
    "objetivos": "Compreender operações básicas de adição e subtração.",
    "metodologia": "Aula expositiva com exercícios práticos em grupo.",
    "recursos": "Quadro, giz, folhas de exercício.",
    "avaliacao": "Quiz autocorrigido ao final da trilha."
  },
  "atividade": "Resolver 10 problemas de adição e subtração em duplas.",
  "trilha": {
    "topicos": [
      {"titulo": "Adição de números naturais", "resumo": "Como somar números naturais passo a passo."},
      {"titulo": "Subtração de números naturais", "resumo": "Como subtrair números naturais passo a passo."}
    ],
    "quiz": {
      "questoes": [
        {"enunciado": "Quanto é 2 + 2?", "opcoes": ["3", "4", "5", "6"], "correta": 1},
        {"enunciado": "Quanto é 9 - 3?", "opcoes": ["5", "7", "6", "4"], "correta": 2}
      ]
    }
  }
}'

lesson_status="$(curl -s -o /tmp/smoke_lesson.json -w '%{http_code}' \
  -b "$COOKIE_JAR" -H 'Content-Type: application/json' \
  -d "$lesson_payload" "$API_URL/api/lessons")"
[[ "$lesson_status" == "201" ]] || fail "create lesson expected 201, got $lesson_status: $(cat /tmp/smoke_lesson.json)"

lesson_id="$(jq -r '.id' /tmp/smoke_lesson.json)"
lesson_status_field="$(jq -r '.status' /tmp/smoke_lesson.json)"
[[ "$lesson_id" != "null" && -n "$lesson_id" ]] || fail "could not extract lesson id: $(cat /tmp/smoke_lesson.json)"
[[ "$lesson_status_field" == "pronto" ]] || fail "expected lesson status 'pronto', got '$lesson_status_field'"
ok "lesson created (id=$lesson_id, status=$lesson_status_field)"

# ---------------------------------------------------------------------------
# 3. Publish the trail
# ---------------------------------------------------------------------------
log "Publishing trail for lesson $lesson_id"
publish_status="$(curl -s -o /tmp/smoke_publish.json -w '%{http_code}' \
  -b "$COOKIE_JAR" -X POST "$API_URL/api/trails/$lesson_id/publish")"
[[ "$publish_status" == "200" ]] || fail "publish expected 200, got $publish_status: $(cat /tmp/smoke_publish.json)"

codigo="$(jq -r '.codigo' /tmp/smoke_publish.json)"
[[ -n "$codigo" && "$codigo" != "null" ]] || fail "could not extract codigo: $(cat /tmp/smoke_publish.json)"
ok "trail published (codigo=$codigo)"

# ---------------------------------------------------------------------------
# 4. Public read (no cookie) — must NOT leak the answer key
# ---------------------------------------------------------------------------
log "Reading public trail (unauthenticated)"
public_status="$(curl -s -o /tmp/smoke_public.json -w '%{http_code}' "$API_URL/api/t/$codigo")"
[[ "$public_status" == "200" ]] || fail "public trail read expected 200, got $public_status: $(cat /tmp/smoke_public.json)"
if grep -qi 'correta' /tmp/smoke_public.json; then
  fail "public trail response leaks the answer key ('correta' field present)"
fi
ok "public trail read OK, no answer key leaked"

# ---------------------------------------------------------------------------
# 5. Start attempt
# ---------------------------------------------------------------------------
log "Starting student attempt"
attempt_status="$(curl -s -o /tmp/smoke_attempt.json -w '%{http_code}' \
  -H 'Content-Type: application/json' -d '{"nome":"Aluno Teste"}' \
  "$API_URL/api/t/$codigo/attempt")"
[[ "$attempt_status" == "201" ]] || fail "start attempt expected 201, got $attempt_status: $(cat /tmp/smoke_attempt.json)"

attempt_id="$(jq -r '.attempt_id' /tmp/smoke_attempt.json)"
[[ -n "$attempt_id" && "$attempt_id" != "null" ]] || fail "could not extract attempt_id: $(cat /tmp/smoke_attempt.json)"
ok "attempt started (attempt_id=$attempt_id)"

# ---------------------------------------------------------------------------
# 6. Submit answers (all correct, using question IDs from public read)
# ---------------------------------------------------------------------------
log "Submitting answers (all correct)"
# Map: question 1 (2+2) -> option index 1 ("4"); question 2 (9-3) -> option index 2 ("6")
answers_payload="$(jq -c '
  {
    answers: [
      {quiz_question_id: (.quiz.questoes[0].id), escolhida: 1},
      {quiz_question_id: (.quiz.questoes[1].id), escolhida: 2}
    ]
  }
' /tmp/smoke_public.json)"

expected_total="$(jq -r '.quiz.questoes | length' /tmp/smoke_public.json)"

answers_status="$(curl -s -o /tmp/smoke_answers.json -w '%{http_code}' \
  -H 'Content-Type: application/json' -d "$answers_payload" \
  "$API_URL/api/attempts/$attempt_id/answers")"
[[ "$answers_status" == "200" ]] || fail "submit answers expected 200, got $answers_status: $(cat /tmp/smoke_answers.json)"

pontos="$(jq -r '.pontos' /tmp/smoke_answers.json)"
acertos="$(jq -r '.acertos' /tmp/smoke_answers.json)"
total="$(jq -r '.total' /tmp/smoke_answers.json)"

expected_pontos=$((10 * expected_total))
[[ "$total" == "$expected_total" ]] || fail "expected total=$expected_total, got $total"
[[ "$acertos" == "$total" ]] || fail "expected acertos == total ($total), got acertos=$acertos"
[[ "$pontos" == "$expected_pontos" ]] || fail "expected pontos=$expected_pontos (10 * $expected_total), got $pontos"
ok "answers scored correctly (pontos=$pontos, acertos=$acertos/$total)"

# ---------------------------------------------------------------------------
# 7. Stats (teacher-authenticated)
# ---------------------------------------------------------------------------
log "Fetching class stats"
stats_status="$(curl -s -o /tmp/smoke_stats.json -w '%{http_code}' \
  -b "$COOKIE_JAR" "$API_URL/api/trails/$lesson_id/stats")"
[[ "$stats_status" == "200" ]] || fail "stats expected 200, got $stats_status: $(cat /tmp/smoke_stats.json)"

total_alunos="$(jq -r '.total_alunos' /tmp/smoke_stats.json)"
[[ "$total_alunos" -ge 1 ]] || fail "expected total_alunos >= 1, got $total_alunos"
ok "stats OK (total_alunos=$total_alunos)"

# ---------------------------------------------------------------------------
# 8. PDF export
# ---------------------------------------------------------------------------
log "Fetching PDF export"
pdf_headers="$(mktemp)"
pdf_status="$(curl -s -o /tmp/smoke.pdf -D "$pdf_headers" -w '%{http_code}' "$API_URL/api/t/$codigo/export.pdf")"
[[ "$pdf_status" == "200" ]] || fail "PDF export expected 200, got $pdf_status"
content_type="$(grep -i '^content-type:' "$pdf_headers" | tr -d '\r' | cut -d' ' -f2-)"
[[ "$content_type" == "application/pdf"* ]] || fail "expected content-type application/pdf, got '$content_type'"
rm -f "$pdf_headers"
echo "PDF OK"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
log "SMOKE TEST SUCCESS"
cat <<SUMMARY
  lesson_id     = $lesson_id
  codigo        = $codigo
  attempt_id    = $attempt_id
  pontos        = $pontos ($acertos/$total acertos)
  total_alunos  = $total_alunos
  PDF export    = OK ($content_type)
SUMMARY

exit 0
