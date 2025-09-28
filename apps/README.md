# WarmHouse Microservices Architecture

Этот проект демонстрирует постепенный переход от монолита к микросервисной архитектуре с использованием API Gateway.

## Архитектура

### Сервисы

1. **API Gateway** (порт 8084) - единая точка входа для всех запросов
2. **Smart Home Monolith** (порт 8080) - основной сервис с базой данных
3. **Temperature API** (порт 8081) - сервис для получения данных о температуре
4. **Telemetry Service** (порт 8082) - сервис для сбора и хранения телеметрии
5. **Device Management Service** (порт 8083) - сервис для управления устройствами
6. **PostgreSQL** (порт 5432) - база данных для монолита

### Взаимодействие

```
Client → API Gateway → Microservices
                ↓
        Smart Home Monolith → PostgreSQL
```

## Запуск

### Требования

- Docker
- Docker Compose

### Команды

```bash
# Запуск всех сервисов
docker-compose up --build

# Запуск в фоновом режиме
docker-compose up -d --build

# Остановка
docker-compose down

# Просмотр логов
docker-compose logs -f [service_name]
```

## API Endpoints

### API Gateway (порт 8084)

- `GET /health` - проверка здоровья всех сервисов
- `GET /api/temperature/*` - проксирование к Temperature API
- `GET /api/telemetry/*` - проксирование к Telemetry Service
- `GET /api/device/*` - проксирование к Device Management Service
- `GET /api/aggregate/telemetry` - агрегированные данные

### Smart Home Monolith (порт 8080)

- `GET /health` - проверка здоровья
- `GET /api/v1/sensors` - список сенсоров
- `GET /api/v1/sensors/:id` - сенсор по ID
- `POST /api/v1/sensors` - создание сенсора
- `PUT /api/v1/sensors/:id` - обновление сенсора
- `DELETE /api/v1/sensors/:id` - удаление сенсора
- `PATCH /api/v1/sensors/:id/value` - обновление значения сенсора
- `GET /api/v1/sensors/temperature/:location` - температура по локации
- `GET /api/v1/sensors/devices` - список устройств
- `POST /api/v1/sensors/devices` - создание устройства
- `POST /api/v1/sensors/devices/:id/commands` - отправка команды устройству
- `GET /api/v1/sensors/telemetry` - телеметрия
- `POST /api/v1/sensors/telemetry` - создание записи телеметрии

### Temperature API (порт 8081)

- `GET /health` - проверка здоровья
- `GET /temperature?location=...` - температура по локации
- `GET /temperature/:id` - температура по ID сенсора

### Telemetry Service (порт 8082)

- `GET /health` - проверка здоровья
- `GET /telemetry?device_id=...&from=...&to=...` - телеметрия с фильтрами
- `POST /telemetry` - создание записи телеметрии
- `POST /telemetry/bulk` - массовое создание записей
- `GET /telemetry/:id` - запись по ID

### Device Management Service (порт 8083)

- `GET /health` - проверка здоровья
- `GET /devices?house_id=...&status=...` - устройства с фильтрами
- `POST /devices` - создание устройства
- `GET /devices/:id` - устройство по ID
- `PATCH /devices/:id` - обновление устройства
- `PATCH /devices/:id/status` - обновление статуса
- `POST /devices/:id/commands` - отправка команды

## Примеры использования

### Получение температуры

```bash
# Через API Gateway
curl http://localhost:8084/api/temperature/temperature?location=Living%20Room

# Напрямую через Temperature API
curl http://localhost:8081/temperature?location=Living%20Room
```

### Получение устройств

```bash
# Через API Gateway
curl http://localhost:8084/api/device/devices

# Через монолит
curl http://localhost:8080/api/v1/sensors/devices
```

### Создание телеметрии

```bash
curl -X POST http://localhost:8084/api/telemetry/telemetry \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "device-001",
    "metrics_names": ["temperature", "humidity"],
    "metrics_values": [22.5, 45.0]
  }'
```

### Отправка команды устройству

```bash
curl -X POST http://localhost:8080/api/v1/sensors/devices/device-001/commands \
  -H "Content-Type: application/json" \
  -d '{
    "command": "turn_on",
    "parameters": {"brightness": 80}
  }'
```

## Мониторинг

### Проверка здоровья

```bash
# API Gateway
curl http://localhost:8084/health

# Все сервисы
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
```

### Логи

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f api-gateway
docker-compose logs -f app
docker-compose logs -f telemetry-service
docker-compose logs -f device-service
```

## Разработка

### Структура проекта

```
apps/
├── api-gateway/          # API Gateway
├── smart_home/           # Монолит
├── temperature-api/      # Temperature API
├── telemetry-service/    # Telemetry Service
├── device-service/       # Device Management Service
├── docker-compose.yml    # Docker Compose конфигурация
└── README.md            # Этот файл
```

### Добавление нового сервиса

1. Создайте папку для сервиса
2. Добавьте `main.go`, `go.mod`, `Dockerfile`
3. Обновите `docker-compose.yml`
4. Добавьте маршруты в API Gateway
5. Обновите документацию

## Примечания

- Все сервисы используют in-memory хранилище для простоты
- В production следует использовать реальные базы данных
- Добавлена базовая обработка ошибок и логирование
- API Gateway поддерживает CORS для веб-приложений