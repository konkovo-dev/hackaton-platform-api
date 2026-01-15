# UC-HX-19 — Управление оргсоставом хакатона (пригласить/удалить/изменить роли)

## Зачем нужен юзкейс
Владелец хакатона (`HX_ROLE_OWNER`) управляет оргсоставом: назначает и снимает роли `HX_ROLE_ORGANIZER`, `HX_ROLE_MENTOR`, `HX_ROLE_JUDGE`. Назначение роли происходит через единый механизм приглашений (`INV_ROLE_ASSIGNMENT`), чтобы роль появлялась только после принятия приглашения.

---

## Участники
- Владелец хакатона (залогинен, имеет `HX_ROLE_OWNER`)
- Получатель приглашения (залогинен)

---

## Триггер
Владелец открывает `SEC_ORGANIZER` и управляет оргсоставом (добавить роль / убрать роль / посмотреть список).

---

## Предусловия
- `auth == true`
- `HAS(HX_ROLE_OWNER)`

---

## Эндпоинты
- `GET /v1/hackathons/{hackathon_id}/staff`
- `POST /v1/hackathons/{hackathon_id}/staff:invite`
- `DELETE /v1/hackathons/{hackathon_id}/staff/{user_id}/roles/{role}`

---

## Что возвращаем
- Для `GET`: список пользователей оргсостава и их роли `HX_ROLE_*`, включая ожидающие назначения (через `INV_ROLE_ASSIGNMENT` со статусом `INV_PENDING`).
- Для `POST`: uuid созданного приглашения.
- Для `DELETE`: подтверждение удаления роли.

---

## Правила
| Условие | Результат |
|---|---|
| `auth == true AND HAS(HX_ROLE_OWNER)` | Разрешено управлять оргсоставом. |
| `auth == true AND NOT HAS(HX_ROLE_OWNER)` | Доступ запрещён. |
| `POST staff:invite` | Создаётся `INV_ROLE_ASSIGNMENT` со статусом `INV_PENDING`. |
| `INV_ROLE_ASSIGNMENT` принят получателем | Назначается соответствующая роль `HX_ROLE_*` получателю. |
| `DELETE role` | Роль снимается сразу (без приглашения). |

---

```mermaid
sequenceDiagram
  autonumber
  actor Владелец
  actor Получатель
  participant UI as UI
  participant Gateway as gateway
  participant ParticipationRoles as participation-and-roles

  Владелец->>UI: Открыть SEC_ORGANIZER и управление оргсоставом
  UI->>Gateway: GET /v1/hackathons/{hackathon_id}/staff
  Gateway->>ParticipationRoles: Список HX_ROLE_* + INV_ROLE_ASSIGNMENT (INV_PENDING)
  ParticipationRoles-->>Gateway: OK
  Gateway-->>UI: OK

  Владелец->>UI: Пригласить на роль (INV_ROLE_ASSIGNMENT)
  UI->>Gateway: POST /v1/hackathons/{hackathon_id}/staff:invite
  Gateway->>ParticipationRoles: Создать INV_ROLE_ASSIGNMENT (INV_PENDING)
  ParticipationRoles-->>Gateway: OK (invitation_id)
  Gateway-->>UI: OK

  Получатель->>UI: Принять приглашение
  UI->>Gateway: Принять INV_ROLE_ASSIGNMENT
  Gateway->>ParticipationRoles: Принять INV_ROLE_ASSIGNMENT
  ParticipationRoles-->>ParticipationRoles: Назначить HX_ROLE_*
  ParticipationRoles-->>Gateway: OK
  Gateway-->>UI: OK

  Владелец->>UI: Снять роль
  UI->>Gateway: DELETE /v1/hackathons/{hackathon_id}/staff/{user_id}/roles/{role}
  Gateway->>ParticipationRoles: Снять HX_ROLE_*
  ParticipationRoles-->>Gateway: OK
  Gateway-->>UI: OK
```