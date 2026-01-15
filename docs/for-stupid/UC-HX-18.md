# UC-HX-18 — Редактировать хакатон (основные поля)

## Зачем нужен юзкейс
`HX_ROLE_OWNER` или `HX_ROLE_ORGANIZER` меняет основные данные хакатона, чтобы подготовить его к публикации и управлять регистрацией (в том числе через `REG_ALLOW_INDIVIDUAL` и `REG_ALLOW_TEAM`).

---

## Участники
- Пользователь (залогинен)
- Gateway (HTTP API)
- Auth Service (introspect)
- Hackathon Service
- Participation&Roles Service

---

## Триггер
Пользователь открывает `SEC_ORGANIZER` и сохраняет изменения.

---

## Предусловия
- Пользователь отправляет запрос с `Authorization: Bearer <token>`.

---

## Авторизация (обязательное правило)
Эндпоинт защищён авторизацией на уровне gateway:
- перед выполнением handler’а gateway обязан провалидировать токен через `AuthService.IntrospectToken`
- если токен невалиден/истёк/неподдерживаемый — запрос отклоняется до вызова доменных сервисов

---

## Эндпоинт
- `PUT /v1/hackathons/{hackathon_id}`

---

## Что возвращаем
- `OK` (успешное сохранение)

---

## Правила
| Условие | Результат |
|---|---|
| `AuthService.IntrospectToken` вернул `valid == false` | `401 Unauthorized`, доменные сервисы не вызываются |
| `valid == true` и `HAS(HX_ROLE_OWNER) OR HAS(HX_ROLE_ORGANIZER)` | Изменения сохраняются |
| `valid == true` и `NOT HAS(HX_ROLE_OWNER) AND NOT HAS(HX_ROLE_ORGANIZER)` | `403 Forbidden` |

> Важно: проверка ролей выполняется **внутри hackathon-service**, а не в gateway.

---

## Sequence

```mermaid
sequenceDiagram
  autonumber
  actor Пользователь
  participant UI as UI
  participant Gateway as gateway
  participant Auth as auth-service
  participant Hackathon as hackathon
  participant ParticipationRoles as participation-and-roles

  Пользователь->>UI: Изменить поля в SEC_ORGANIZER
  UI->>Gateway: PUT /v1/hackathons/{hackathon_id} (Authorization: Bearer token, payload)

  Gateway->>Auth: IntrospectToken(token)
  alt token invalid
    Auth-->>Gateway: valid=false
    Gateway-->>UI: 401 Unauthorized
  else token valid
    Auth-->>Gateway: valid=true, user_id

    Gateway->>Hackathon: UpdateHackathon(hackathon_id, actor_user_id=user_id, payload)

    Hackathon->>ParticipationRoles: CheckHackathonRole(hackathon_id, user_id, [HX_ROLE_OWNER,HX_ROLE_ORGANIZER])
    alt no required role
      ParticipationRoles-->>Hackathon: allowed=false
      Hackathon-->>Gateway: 403 Forbidden
      Gateway-->>UI: 403 Forbidden
    else allowed
      ParticipationRoles-->>Hackathon: allowed=true
      Hackathon->>Hackathon: Сохранить изменения (Update)
      Hackathon-->>Gateway: OK
      Gateway-->>UI: OK
    end
  end

  UI-->>Пользователь: Показать результат
```
