DB_URL ?= postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up
migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down
sqlc:
	sqlc generate
test:
	go test ./...
