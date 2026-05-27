FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o forum .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/forum .
COPY templates/ templates/
COPY static/ static/

RUN mkdir -p static/uploads

EXPOSE 8080

CMD ["./forum"]
