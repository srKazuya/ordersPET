FROM golang:1.24.5

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o orders cmd/orders/main.go

CMD ["./orders"]