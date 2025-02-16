# avito-shop

# Запуск
### 1. Необходимо создать .env файл и заполнить пустые поля нужными данными
.env file:
```dotenv
PASSWORD_SALT=
SIGNING_KEY=
TOKEN_TTL_HOURS=
MONEY_FOR_START=1000
####
DATABASE_PORT=5433
DATABASE_USER=postgres
DATABASE_PASSWORD=#password
DATABASE_NAME=shop
DATABASE_HOST=localhost
SERVER_PORT=8080
```
### 2. 
```bash
docker-compose build
docker-compose up -d
```

# Тестирование и линтер
### Команда для запуска тестов 
```bash
make test 
```
### Запуск интеграцинных тестов
```bash
make test-integration
```
### Запуск тестов с получением покрыти в файл cover.out
```bash
make get-coverage
```
### Запуск линтера
```bash
make lint
```

# Вопросы.проблемы
## 1. Можно ли менять докер файлы, которые вы приложили к заданию?
#### Я убрала енвы в .env, мне кажется так безопаснее
## 2. В Dockerfile указано, что cmd должна лежать в internal, обычно я видела, что cmd лежит в корне проекта
#### Оставила так, как просит Dockerfile