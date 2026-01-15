# UC-HX-02 — Регистрация на хакатон (выбор режима участия)

## Зачем нужен юзкейс
Пользователь, у которого нет участия в хакатоне (`PART_NONE`), должен зарегистрироваться в одном из разрешённых режимов. Результат регистрации фиксирует состояние участия (`PART_*`) и создаёт “профиль участника хакатона” для матчмейкинга (текст + желаемые роли). Это определяет доступ к разделам `SEC_*` и действиям `CTA_*` в дальнейших юзкейсах.

---

## Участники
- Пользователь (залогинен)

---

## Триггер
Пользователь нажимает `CTA_REGISTER` и выбирает вариант регистрации на экране регистрации.

---

## Предусловия
- `auth == true`
- `PART_* == PART_NONE`

---

## Эндпоинты
- `POST /v1/hackathons/{hackathon_id}/register:individual`
- `POST /v1/hackathons/{hackathon_id}/register:lookingForTeam`
- `POST /v1/hackathons/{hackathon_id}/register:createTeam`

---

## Что возвращаем
- Обновлённый контекст пользователя в хакатоне: новый `PART_*` и `team_id` (если появился).

---

## Правила экрана регистрации (какие варианты показать после `CTA_REGISTER`)
| Условие | Показать варианты регистрации |
|---|---|
| `REG_ALLOW_INDIVIDUAL == true AND REG_ALLOW_TEAM == true` | `PART_INDIVIDUAL_ACTIVE`, `PART_TEAM_CAPTAIN`, `PART_LOOKING_FOR_TEAM` |
| `REG_ALLOW_INDIVIDUAL == true AND REG_ALLOW_TEAM == false` | `PART_INDIVIDUAL_ACTIVE` |
| `REG_ALLOW_INDIVIDUAL == false AND REG_ALLOW_TEAM == true` | `PART_TEAM_CAPTAIN`, `PART_LOOKING_FOR_TEAM` |

---

## Правила обработки выбора варианта (результирующий `PART_*`)
| Условие | Действие пользователя | Результат (`PART_*`) |
|---|---|---|
| `REG_ALLOW_INDIVIDUAL == true AND PART_* == PART_NONE` | Выбирает индивидуальную регистрацию (`register:individual`) | `PART_INDIVIDUAL_ACTIVE` |
| `REG_ALLOW_TEAM == true AND PART_* == PART_NONE` | Выбирает “зарегистрироваться и найти команду позже” (`register:lookingForTeam`) | `PART_LOOKING_FOR_TEAM` |
| `REG_ALLOW_TEAM == true AND PART_* == PART_NONE` | Выбирает “создать команду” (`register:createTeam`) | `PART_TEAM_CAPTAIN` |

---

```mermaid
sequenceDiagram
  autonumber
  actor Пользователь
  participant UI as UI
  participant Gateway as gateway
  participant ParticipationRoles as participation-and-roles
  participant Team as team

  Пользователь->>UI: Выбирает вариант регистрации
  UI->>Gateway: Запрос регистрации

  alt Выбран PART_INDIVIDUAL_ACTIVE
    Gateway->>ParticipationRoles: register:individual
    ParticipationRoles-->>Gateway: ok
    Gateway-->>UI: ok
  else Выбран PART_LOOKING_FOR_TEAM
    Gateway->>ParticipationRoles: register:lookingForTeam
    ParticipationRoles-->>Gateway: ok
    Gateway-->>UI: ok
  else Выбран PART_TEAM_CAPTAIN
    Gateway->>Team: register:createTeam
    Team-->>Gateway: ok (team_id)
    Gateway-->>UI: ok
    Team->>Team: Записать событие в outbox
    Team-->>ParticipationRoles: Событие "команда создана для хакатона"
    ParticipationRoles-->>ParticipationRoles: Зафиксировать PART_TEAM_CAPTAIN и team_id
  end

  UI-->>Пользователь: Показать экран
```