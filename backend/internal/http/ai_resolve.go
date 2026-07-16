package http

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/lesson"
)

// errNoAIToken signals the professor has no usable AI token; callers map it to 503.
var errNoAIToken = errors.New("no ai token configured")

// generatorForUser loads the professor's stored AI token, decrypts it, and
// builds a Generator via d.NewGen. Returns errNoAIToken (→503) when the secret
// box is unavailable or the professor has no token.
func (d Deps) generatorForUser(ctx context.Context, userID int64) (lesson.Generator, error) {
	if d.Secret == nil {
		return nil, errNoAIToken
	}
	tok, err := d.Store.GetAIToken(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errNoAIToken
		}
		return nil, err
	}
	plaintext, err := d.Secret.Open(tok.TokenCiphertext, tok.TokenNonce)
	if err != nil {
		return nil, err
	}
	return d.NewGen(ctx, tok.Provider, string(plaintext))
}
