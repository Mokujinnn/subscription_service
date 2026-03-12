# Subscription Service

REST API сервис для агрегации данных об онлайн подписках пользователей.

## Конфигурация

Скопируйте файл конфигурации:
```bash
cp .env.example .env
```

## Swagger документация

Документация генерируется автоматически. Для генерации выполните:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/main.go
```

## Запуск

```bash
docker-compose up
```
