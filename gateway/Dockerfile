FROM golang:1.24.0-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

RUN go env -w GOPROXY=https://goproxy.io,direct

COPY gateway/go.mod gateway/go.sum ./

RUN go mod download

COPY gateway/cmd ./cmd

COPY gateway/internal ./internal

COPY gateway/pkg ./pkg

COPY gateway/bootstrap ./bootstrap

COPY gateway/wire ./wire

RUN go install github.com/google/wire/cmd/wire@latest
RUN wire ./wire

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g ./cmd/app/main.go

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/app

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/main .

RUN mkdir -p /app/logs && chown -R appuser:appgroup /app/logs

USER appuser

EXPOSE 8080

VOLUME ["/app/logs"]

CMD ["./main"]