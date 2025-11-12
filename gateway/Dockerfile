FROM golang:1.24.0-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

RUN go env -w GOPROXY=https://goproxy.io,direct

COPY gateway/go.mod gateway/go.sum ./

RUN go mod download

COPY gateway/cmd ./cmd

COPY gateway/internal ./internal

COPY gateway/pkg ./pkg

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/app

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .

COPY .env .

EXPOSE 8080

CMD ["./main"]