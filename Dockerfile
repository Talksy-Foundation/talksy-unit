# BUILD Stage
FROM golang:1.22 AS builder

WORKDIR /app
COPY . .
COPY vendor ./vendor


RUN go build -mod vendor -o main talksy_unit.go

# RUN Stage
FROM golang:1.22

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]