FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/links-helper-bot ./cmd/links-helper-bot

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /out/links-helper-bot /app/links-helper-bot

USER app

ENTRYPOINT ["/app/links-helper-bot"]
