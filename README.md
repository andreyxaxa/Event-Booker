# EventBooker - сервис бронирования с дедлайнами
Сервис позволяет создавать мероприятия, бронировать места, подтверждать их, и автоматически отменять неподтвержденные брони через заданный интервал времени.

[Старт]()

## Обзор

- UI - http://localhost:8080/v1
- Документация API - Swagger - http://localhost:8080/swagger
- Конфиг - [config/config.go](https://github.com/andreyxaxa/Event-Booker/blob/main/config/config.go). Читается из `.env` файла.

## Запуск

Сервис отправляет оповещения на почту через net/smtp.

Соответственно, требуется :
Gmail аккаунт, с него сервис будет слать оповещения. + app-pasword этого аккаунта. [Как создать пароль приложения](https://support.google.com/accounts/answer/185833?hl=ru)

1. Клонируйте репозиторий
2. В корне создайте `.env` файл, скопируйте туда содержимое [env.example](https://github.com/andreyxaxa/Event-Booker/blob/main/.env.example), подставив в `SMTP_USERNAME` ваш gmail, в `SMTP_PASSWORD` ваш app-password.
   ```
   cp .env.example .env
   ```
3. Выполните, дождитесь запуска сервиса
   ```
   make compose-up
   ```
4. Перейдите на http://localhost:8080/v1 и пользуйтесь сервисом.
<img width="1684" height="991" alt="image" src="https://github.com/user-attachments/assets/1e071257-fa24-44f7-8977-5ad6912285ff" />

- Перейдите на http://localhost:8080/swagger и ознакомьтесь с API, если хотите взаимодействовать с сервисом вручную или из стороннего сервиса.



## Прочие `make` команды
Зависимости:
```
make deps
```
docker compose down -v:
```
make compose-down
```
