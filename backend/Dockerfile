FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/api ./cmd/api

FROM alpine:3.20
COPY --from=build /bin/api /bin/api
COPY --from=build /src/seed /seed
COPY --from=build /src/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
