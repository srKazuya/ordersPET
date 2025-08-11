# OrdersPET — Kafka Producer & Consumer Service

Pet-проект, реализующий взаимодействие с **Apache Kafka** для отправки и обработки заказов.  
Написан на **Go**, использует библиотеку `confluent-kafka-go` для интеграции с Kafka.  

При возврате заказа:
- Данные извлекаются из **PostgreSQL** (поднимается в Docker)
При повторном запросе:
- Данные берутся из кеша (in-memory hash-таблица), минуя PostgreSQL, для ускорения ответа.

## Возможности
- **Kafka Producer** — отправка сообщений в заданную тему Kafka
- **Kafka Consumer** — чтение сообщений и сохранение заказов в хранилище
- Логирование с использованием `log/slog`
- Роутинг с помощью `chi`

## Запуск
```bash
CONFIG_PATH=./config/local.yaml go run cmd/orders/main.go
