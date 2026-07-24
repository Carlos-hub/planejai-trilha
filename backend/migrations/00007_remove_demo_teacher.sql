-- +goose Up
-- The demo teacher is removed from the system; signup replaces it. Cascades to
-- any lessons/turmas/students that belonged to the demo account.
DELETE FROM users WHERE email = 'prof@demo.com';

-- +goose Down
-- Irreversible: the demo teacher and its cascaded data are intentionally gone.
SELECT 1;
