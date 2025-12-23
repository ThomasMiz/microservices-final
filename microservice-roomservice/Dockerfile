FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -a -o roomservice cmd/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/roomservice .

EXPOSE 8080

CMD ["./roomservice"]
