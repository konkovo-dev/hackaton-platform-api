# Hackathon Service Documentation

Полная документация и тестирование Hackathon Service API.

---

## 📚 Содержание

### API Guides
- **[REST API Guide](./rest-guide.md)** - Подробная документация всех REST эндпоинтов
- **[gRPC API Guide](./grpc-guide.md)** - Документация gRPC методов (в разработке)

### Testing
- **[Testing Summary](./TESTING_SUMMARY.md)** - Полный отчет о тестировании
- **[REST Script](./rest-script.sh)** - Автоматизированные Happy Path тесты
- **[REST Fail Cases Script](./rest-script-fail-cases.sh)** - Тесты валидаций и ограничений

### Rules & Specifications
- **[Hackathon Policy Spec](../rules/hackathon.md)** - Бизнес-правила и валидации

---

## 🚀 Quick Start

### 1. Запуск сервисов

```bash
# Из корня проекта
cd deployments
docker-compose up -d

# Проверка статуса
docker-compose ps
```

### 2. Применение миграций

```bash
# Из корня проекта
make hackaton-service-migrate-up
make participation-and-roles-service-migrate-up
make identity-service-migrate-up
make auth-service-migrate-up
```

### 3. Запуск тестов

```bash
cd docs/hackathon

# Основные тесты (Happy Path)
chmod +x rest-script.sh
./rest-script.sh

# Тесты валидаций (Fail Cases)
chmod +x rest-script-fail-cases.sh
./rest-script-fail-cases.sh
```

---

## 🎯 Основные возможности

### Hackathon Management
- ✅ Создание хакатона (DRAFT stage)
- ✅ Обновление с учетом стадии
- ✅ Публикация (DRAFT → Published stages)
- ✅ Валидация перед публикацией
- ✅ Список опубликованных хакатонов

### Task Management
- ✅ Добавление/обновление задания
- ✅ Доступ по ролям и стадиям
- ✅ Ограничения на изменение (запрет на JUDGING/FINISHED)

### Result Management
- ✅ Черновик результата (JUDGING stage)
- ✅ Публикация результата (переход в FINISHED)
- ✅ Публичный доступ после публикации

### Announcements
- ✅ Создание объявлений (только для опубликованных хакатонов)
- ✅ Чтение для staff и participants
- ✅ Обновление и удаление (OWNER/ORGANIZER)

### Access Control
- ✅ DRAFT: только OWNER/ORGANIZER
- ✅ Published: все авторизованные пользователи
- ✅ Task: по ролям и стадиям
- ✅ Result: draft (OWNER/ORG), published (all)
- ✅ Announcements: staff и participants

---

## 📊 Стадии хакатона

```
DRAFT → UPCOMING → REGISTRATION → PRESTART → RUNNING → JUDGING → FINISHED
```

### Переходы между стадиями

| From | To | Trigger |
|------|----|----|
| DRAFT | UPCOMING | Публикация (`published_at` set) |
| UPCOMING | REGISTRATION | `now >= registration_opens_at` |
| REGISTRATION | PRESTART | `now >= registration_closes_at` |
| PRESTART | RUNNING | `now >= starts_at` |
| RUNNING | JUDGING | `now >= ends_at` |
| JUDGING | FINISHED | `result_published_at` set OR `now >= judging_ends_at` |

---

## 🔒 Правила валидации

### По стадиям

| Поле | DRAFT | UPCOMING | REG | PRESTART | RUNNING | JUDGING | FINISHED |
|------|-------|----------|-----|----------|---------|---------|----------|
| name, desc | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| location | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| team_size_max | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| DisableType | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| EnableType | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| task | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |

### Режимы валидации

**Soft Mode (DRAFT):**
- Ошибки возвращаются информационно
- Сохранение всегда происходит
- Позволяет постепенное заполнение

**Strict Mode (Published):**
- Критические ошибки блокируют операцию
- Коды: `REQUIRED`, `TIME_RULE`, `TIME_LOCKED`, `FORBIDDEN`, `POLICY_RULE`
- Гарантирует корректность данных

### TIME_RULE

```
registration_opens_at + 1h < registration_closes_at <= starts_at < ends_at + 1h < judging_ends_at
```

### Типы изменений времени

**TYPE-A** (`registration_opens_at`, `judging_ends_at`):
- Правило: `now < old && now < new`
- Можно менять только будущие даты

**TYPE-B** (`registration_closes_at`, `starts_at`, `ends_at`):
- Правило: `now < old && old < new`
- Можно только продлевать вперед

---

## 🧪 Тестовые сценарии

### rest-script.sh (20 тестов)
1. ✅ Регистрация пользователей
2. ✅ Создание хакатона (DRAFT)
3. ✅ Получение DRAFT (owner vs non-owner)
4. ✅ Добавление задания
5. ✅ Доступ к заданию по ролям
6. ✅ Обновление хакатона в DRAFT
7. ✅ Валидация для публикации
8. ✅ Публикация хакатона
9. ✅ Получение с `include_task`
10. ✅ Обновление location на UPCOMING
11. ✅ Обновление team_size_max на UPCOMING
12. ✅ Отключение типа регистрации на UPCOMING
13. ✅ Доступ к опубликованному хакатону
14. ✅ Список хакатонов
15-20. ✅ CRUD announcements

### rest-script-fail-cases.sh (5 тестов)
1. ✅ Location изменение на RUNNING (FAIL)
2. ✅ TeamSizeMax изменение на RUNNING (FAIL)
3. ✅ DisableType на RUNNING (FAIL)
4. ✅ Task обновление на JUDGING (FAIL)
5. ✅ Публикация без обязательных полей (FAIL)

---

## 📋 API Endpoints Summary

### Hackathon
- `POST /v1/hackathons` - Создать
- `GET /v1/hackathons/{id}` - Получить
- `PUT /v1/hackathons/{id}` - Обновить
- `GET /v1/hackathons/{id}:validate` - Валидировать
- `POST /v1/hackathons/{id}:publish` - Опубликовать
- `GET /v1/hackathons` - Список

### Task
- `GET /v1/hackathons/{id}:task` - Получить задание
- `PUT /v1/hackathons/{id}:task` - Обновить задание

### Result
- `GET /v1/hackathons/{id}:result` - Получить результат
- `PUT /v1/hackathons/{id}:result` - Обновить черновик
- `POST /v1/hackathons/{id}:result:publish` - Опубликовать результат

### Announcements
- `GET /v1/hackathons/{id}/announcements` - Список
- `POST /v1/hackathons/{id}/announcements` - Создать
- `PUT /v1/hackathons/{id}/announcements/{aid}` - Обновить
- `DELETE /v1/hackathons/{id}/announcements/{aid}` - Удалить

---

## 🔍 Debugging

### Проверка логов

```bash
# Hackathon service
docker-compose logs -f hackaton-service

# Participation & Roles service
docker-compose logs -f participation-and-roles-service

# Gateway
docker-compose logs -f gateway
```

### Проверка базы данных

```bash
# Подключение к PostgreSQL
docker-compose exec postgres psql -U postgres -d hackathon_db

# Проверка хакатонов
SELECT id, name, stage, state, published_at FROM hackathon.hackathons;

# Проверка ролей
SELECT * FROM participation_roles.hackathon_roles;
```

### Общие проблемы

**Проблема:** `authentication service error`
- **Решение:** Проверить, что `SERVICE_AUTH_TOKEN` одинаковый во всех сервисах

**Проблема:** `role not found` после создания
- **Решение:** Подождать 1-2 секунды для обработки outbox events

**Проблема:** `access denied` для DRAFT hackathon
- **Решение:** Ожидаемое поведение, DRAFT доступен только OWNER/ORGANIZER

---

## 📖 Дополнительная документация

- [Docker Setup Guide](../docker-setup.md) - Настройка окружения
- [Hackathon Rules](../rules/hackathon.md) - Подробные бизнес-правила
- [Testing Summary](./TESTING_SUMMARY.md) - Детальный отчет о тестировании

---

## ✅ Checklist для разработчиков

- [ ] Сервисы запущены (`docker-compose up -d`)
- [ ] Миграции применены (все сервисы)
- [ ] REST Happy Path тесты проходят (`./rest-script.sh`)
- [ ] REST Fail Cases тесты проходят (`./rest-script-fail-cases.sh`)
- [ ] Проверены логи на наличие ошибок
- [ ] Код следует стилю Identity/Hackathon services
- [ ] Все TODO items completed

---

## 📝 Changelog

### 2026-01-25 - Major Refactoring
- ✅ Task и Result как отдельные ресурсы
- ✅ Stage-based validation rules
- ✅ Soft vs Strict validation modes
- ✅ Complete access control policies
- ✅ Comprehensive test coverage

---

Для вопросов и предложений обращайтесь к документации в [docs/rules/hackathon.md](../rules/hackathon.md).
