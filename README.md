# ordersPET
KAFKA/Go
Запуск: ``` CONFIG_PATH=./config/local.yaml go run cmd/orders/main.go ```

goose -dir db/migrations postgres "postgres://postgres:postgres@localhost:5436/orders?sslmode=disable" up
