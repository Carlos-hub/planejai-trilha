-- name: UpsertAIToken :one
INSERT INTO ai_tokens (user_id, provider, token_ciphertext, token_nonce)
VALUES ($1,$2,$3,$4)
ON CONFLICT (user_id) DO UPDATE
  SET provider=EXCLUDED.provider,
      token_ciphertext=EXCLUDED.token_ciphertext,
      token_nonce=EXCLUDED.token_nonce,
      updated_at=now()
RETURNING *;
-- name: GetAIToken :one
SELECT * FROM ai_tokens WHERE user_id=$1;
-- name: DeleteAIToken :exec
DELETE FROM ai_tokens WHERE user_id=$1;
