# Identity Service Documentation

Документация для тестирования Identity Service разделена на отдельные файлы по сервисам и протоколам.

## MeService (Мой профиль)

Сервис для управления собственным профилем пользователя. **Требует авторизации** (Bearer token).

### Endpoints

- **GetMe** — получить свой профиль
- **UpdateMe** — обновить профиль
- **UpdateMySkills** — обновить навыки (replace-all)
- **UpdateMyContacts** — обновить контакты (replace-all)

### Guides

- 📡 **gRPC**: [me-service-grpc-guide.md](./me-service-grpc-guide.md)
- 🌐 **REST**: [me-service-rest-guide.md](./me-service-rest-guide.md)

---
## UsersService (Профили пользователей)

API для просмотра профилей других пользователей. **Требует авторизации** — доступен только залогиненным участникам платформы.

### Endpoints

- **GetUser** — получить профиль пользователя (с visibility правилами)
- **BatchGetUsers** — получить несколько профилей за один запрос
- **ListUsers** — поиск и фильтрация пользователей с пагинацией

### Guides

- 📡 **gRPC**: [users-service-grpc-guide.md](./users-service-grpc-guide.md)
- 🌐 **REST**: [users-service-rest-guide.md](./users-service-rest-guide.md)

> **Важно**: Перед тестированием UsersService выполните SQL-скрипты для добавления тестовых данных (см. раздел "Подготовка тестовых данных" в guides).

---

## SkillsService (Каталог навыков)

API для просмотра каталога доступных навыков. **Требует авторизации**.

### Endpoints

- **ListSkillCatalog** — получить список навыков из каталога с фильтрацией и пагинацией

### Guides

- 📡 **gRPC**: [skills-service-grpc-guide.md](./skills-service-grpc-guide.md)
- 🌐 **REST**: [skills-service-rest-guide.md](./skills-service-rest-guide.md)

---

## Visibility Rules

### MeService
Возвращает **все** навыки и контакты текущего пользователя, включая:
- Глобальную видимость (`VisibilitySettings`)
- Per-contact видимость для каждого контакта

### UsersService
Применяет правила видимости:
- Если `skills_visibility = PRIVATE` → `skills = []`
- Если `contacts_visibility = PRIVATE` → `contacts = []`
- Если `contacts_visibility = PUBLIC` → только контакты с `visibility = PUBLIC`
- Поле `visibility` у контактов **не возвращается** (публичный API)

---

## Общее предусловие: Регистрация через Auth Service

Перед тестированием рекомендуется зарегистрировать тестового пользователя и сохранить токен:

```bash
# gRPC
RESPONSE=$(grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "testuser@example.com",
  "password": "SecurePass123",
  "first_name": "Test",
  "last_name": "User",
  "timezone": "UTC"
}' localhost:50051 auth.v1.AuthService/Register)

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken')
```

```bash
# REST
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "testuser@example.com",
    "password": "SecurePass123",
    "first_name": "Test",
    "last_name": "User",
    "timezone": "UTC"
  }')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
```

> Подробнее: [../auth/grpc-guide.md](../auth/grpc-guide.md) | [../auth/rest-guide.md](../auth/rest-guide.md)

---

## Тестовые данные

### MeService
- **Требует**: авторизацию (Bearer token)
- **Использует**: реальную регистрацию через `auth-service`
- **Тестовые пользователи**: `alice_me`, `testuser_me`, `testuser_combined`

### UsersService
- **Требует**: авторизацию (Bearer token) + тестовые данные в БД (см. guides)
- **Тестовые пользователи**:
  - `bob_public` (`b0b00000-...-001`) — все PUBLIC
  - `charlie_mixed` (`c4a41e00-...-002`) — skills PRIVATE, contacts PUBLIC
  - `diana_private` (`d1a4a000-...-003`) — все PRIVATE
  - `eve_golang` (`e5e00000-...-004`) — все PUBLIC, для поиска по навыкам

**Важно**: Тестовые UUID для UsersService не конфликтуют с реальными пользователями из MeService.

---

## Дополнительно

- **Docker setup**: [../docker-setup.md](../docker-setup.md)
- **Auth service**: [../auth/](../auth/)
- **gRPC Endpoint**: `localhost:50051`
- **REST Gateway**: `http://localhost:8080`

