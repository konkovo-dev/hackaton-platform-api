# Hackathon Policy Spec — HackathonStaff domain (v1.1, raw)

Нотация правила: `<ACTION> @ <predicate>`

---

## 0) Контекст для проверок

### 0.1 Actor context (Role/Participation service)
- `roles[]`: `{OWNER, ORGANIZER, MENTOR, JURY}`
- `particip.kind`: `{NONE, LOOKING_FOR_TEAM, SINGLE, TEAM}`

### 0.2 Inbox context (Role/Participation service)
Инвайты staff хранятся в Role/Participation service.

`StaffInvitation`:
- `id`
- `hackathon_id`
- `target_user_id`
- `requested_role ∈ {ORGANIZER, MENTOR, JURY}`
- `created_by_user_id`
- `status ∈ {PENDING, ACCEPTED, DECLINED, CANCELED, EXPIRED}`
- `message` (string | null)
- `created_at`
- `expires_at` (timestamp | null)

---

## 1) Инварианты домена

### 1.1 Взаимоисключение staff и participation
В одном хакатоне запрещено одновременно:
- иметь любую роль из `{OWNER, ORGANIZER, MENTOR, JURY}`
- и иметь участие `particip.kind in {LOOKING_FOR_TEAM, SINGLE, TEAM}`

---

## 2) Предикаты

### 2.1 Роли
- `role in {…}`: actor имеет хотя бы одну роль из множества
- `has_role(x)`: пользователь имеет роль `x`

### 2.2 Состояние инвайта
- `invite.status == PENDING`
- `invite.target_user_id == actor.user_id`
- `invite.requested_role in {ORGANIZER, MENTOR, JURY}`

---

## 3) StaffInviteAllowed(target, requested_role)

`StaffInviteAllowed(target, requested_role)` выполняется, если:
- `requested_role in {ORGANIZER, MENTOR, JURY}`
- `target.particip.kind == NONE`
- `target` не имеет роли `requested_role`
- `target` не имеет роли `OWNER`

---

## 4) Таблица действий

| action | includes | описание | условие | описание условия |
|---|---|---|---|---|
| `HackathonStaff.Read` | staff list | Читать список staff хакатона | `HackathonStaff.Read @ role in {OWNER, ORGANIZER, MENTOR, JURY}` | Доступно любому staff |
| `HackathonStaff.Invite` | invitation fields | Создать инвайт на staff-роль | `HackathonStaff.Invite @ has_role(OWNER) && StaffInviteAllowed(target, invite.requested_role)` | Доступно OWNER; нельзя пригласить participant; нельзя пригласить на уже имеющуюся роль |
| `HackathonStaff.CancelInvite` | invitation id | Отменить инвайт на staff-роль | `HackathonStaff.CancelInvite @ has_role(OWNER) && invite.status == PENDING` | Доступно OWNER; отмена только для PENDING |
| `HackathonStaff.RemoveRole` | user_id + role | Удалить staff-роль у пользователя | `HackathonStaff.RemoveRole @ has_role(OWNER) && role_to_remove in {ORGANIZER, MENTOR, JURY}` | Доступно OWNER |
| `HackathonStaff.SelfRemoveRole` | role | Удалить свою staff-роль | `HackathonStaff.SelfRemoveRole @ role in {ORGANIZER, MENTOR, JURY}` | Доступно ORGANIZER/MENTOR/JURY |
| `HackathonStaff.AcceptInvite` | invitation id | Принять инвайт на staff-роль | `HackathonStaff.AcceptInvite @ auth && invite.status == PENDING && invite.target_user_id == actor.user_id && StaffInviteAllowed(actor, invite.requested_role)` | Доступно адресату; при принятии назначается роль |
| `HackathonStaff.RejectInvite` | invitation id | Отклонить инвайт на staff-роль | `HackathonStaff.RejectInvite @ auth && invite.status == PENDING && invite.target_user_id == actor.user_id` | Доступно адресату; перевод в DECLINED |

---

## 5) Побочные эффекты действий

### 5.1 Invite
- создаётся `StaffInvitation` со статусом `PENDING`

### 5.2 AcceptInvite
- инвайт переводится в `ACCEPTED`
- пользователю назначается `requested_role`

### 5.3 RejectInvite
- инвайт переводится в `DECLINED`

### 5.4 CancelInvite
- инвайт переводится в `CANCELED`

### 5.5 RemoveRole / SelfRemoveRole
- роль удаляется у пользователя
