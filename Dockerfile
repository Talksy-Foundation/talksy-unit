# BUILD Stage
FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN go build -o main talksyunit.go

# RUN Stage
FROM golang:1.22

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]